package main

import (
	"image"
	"log"
	"os"
	"path/filepath"
	"time"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	apptheme "github.com/lighto/pier/app"
	brclient "github.com/lighto/pier/br"
	"github.com/lighto/pier/config"
	"github.com/lighto/pier/session"
	"github.com/lighto/pier/ui"
	"github.com/lighto/pier/workspace"
)

func main() {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("Pier"))
		w.Option(app.Size(1280, 800))

		if err := run(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

type appState struct {
	cfg      config.AppConfig
	theme    apptheme.Theme
	matTheme *material.Theme

	sidebar   *ui.Sidebar
	timeline  *ui.Timeline
	prompt    *ui.PromptBar
	taskPanel *ui.TaskPanel
	planPanel *ui.PlanPanel
	extUI     *ui.ExtensionUIHandler

	// Workspaces
	workspaces []*workspace.Workspace
	activeWs   *workspace.Workspace

	// Active session
	sess    *session.Session
	sessErr string

	// br client
	brClient *brclient.Client

	// UI state
	startBtn widget.Clickable
	started  bool

	// Ctrl+C double-tap tracking
	lastCtrlC time.Time
}

func run(w *app.Window) error {
	cfg, _ := config.Load()
	theme := apptheme.ThemeByName(cfg.Theme)
	matTheme := material.NewTheme()
	matTheme.Shaper = text.NewShaper(text.WithCollection([]text.FontFace{}))

	cwd, _ := os.Getwd()

	st := &appState{
		cfg:       cfg,
		theme:     theme,
		matTheme:  matTheme,
		sidebar:   ui.NewSidebar(theme, matTheme),
		timeline:  ui.NewTimeline(theme, matTheme),
		prompt:    ui.NewPromptBar(theme, matTheme),
		taskPanel: ui.NewTaskPanel(theme, matTheme),
		planPanel: ui.NewPlanPanel(theme, matTheme),
		extUI:     ui.NewExtensionUIHandler(theme, matTheme),
	}

	// Load existing workspaces or create default
	workspaces, _ := workspace.LoadAll()
	if len(workspaces) == 0 {
		ws := workspace.NewWorkspace(workspace.GenerateID(), filepath.Base(cwd), cwd)
		workspace.Save(ws)
		workspaces = []*workspace.Workspace{ws}
	}
	st.workspaces = workspaces
	st.activeWs = workspaces[0]
	st.refreshSidebar()

	// Init br client
	if bc, err := brclient.NewClient(st.activeWs.Path, cfg.BrPath); err == nil {
		st.brClient = bc
		st.refreshTasks()
	} else {
		st.taskPanel.SetError("br not found: " + err.Error())
	}

	// Load plan
	st.planPanel.SetPlanPath(st.activeWs.PlanPath())

	// Extension UI response callback
	st.extUI.OnResponse = func(resp session.ExtensionUIResponse) {
		if st.sess != nil {
			st.sess.SendCommand(resp)
		}
	}

	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			if st.sess != nil {
				st.sess.Stop()
			}
			return e.Err

		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// Drain session events
			if st.sess != nil {
				if st.sess.DrainEvents() {
					w.Invalidate()
					st.updateSessionStatus()

					// Check for extension UI requests
					st.checkExtensionUI()
				}
				for _, line := range st.sess.DrainErrors() {
					log.Printf("pi stderr: %s", line)
				}
				if !st.sess.Running() {
					st.updateSidebarSessionStatus("idle")
				}
			}

			// Process keyboard shortcuts
			st.processKeys(gtx, w)

			// Check task panel refresh
			if st.taskPanel.RefreshClicked(gtx) {
				st.refreshTasks()
			}

			// Layout
			st.layout(gtx)

			e.Frame(gtx.Ops)

			// Keep invalidating while active
			if st.sess != nil && st.sess.Running() {
				w.Invalidate()
			}
		}
	}
}

func (st *appState) layout(gtx layout.Context) {
	// Sidebar | Divider | Main | Divider | Task Panel
	layout.Flex{}.Layout(gtx,
		// Sidebar
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return st.sidebar.Layout(gtx)
		}),
		// Divider
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutDivider(gtx, st.theme)
		}),
		// Main content
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			// Background
			rect := image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y)
			paint.FillShape(gtx.Ops, st.theme.Palette.Background, clip.Rect(rect).Op())

			return layout.Inset{Top: unit.Dp(12), Left: unit.Dp(16), Right: unit.Dp(16), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if st.extUI.HasActiveDialog() {
					dims, _ := st.extUI.Layout(gtx)
					return dims
				}
				if !st.started {
					return st.layoutStartScreen(gtx)
				}
				return st.layoutSession(gtx)
			})
		}),
		// Divider
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !st.taskPanel.Visible() {
				return layout.Dimensions{}
			}
			return layoutDivider(gtx, st.theme)
		}),
		// Task panel
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return st.taskPanel.Layout(gtx)
		}),
	)
}

func (st *appState) layoutStartScreen(gtx layout.Context) layout.Dimensions {
	if st.startBtn.Clicked(gtx) {
		st.startSession()
	}

	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H3(st.matTheme, st.activeWs.Name)
			lbl.Color = st.theme.Palette.Text
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body1(st.matTheme, st.activeWs.Path)
			lbl.Color = st.theme.Palette.TextSecondary
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(st.matTheme, &st.startBtn, "Start Pi Session")
			btn.Background = st.theme.Palette.Accent
			return btn.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if st.sessErr == "" {
				return layout.Dimensions{}
			}
			return layout.Inset{Top: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(st.matTheme, "Error: "+st.sessErr)
				lbl.Color = st.theme.Palette.Error
				return lbl.Layout(gtx)
			})
		}),
	)
}

func (st *appState) layoutSession(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Status bar
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			var statusText string
			if st.sess != nil {
				s := st.sess.State
				s.Lock()
				statusText = s.Status.String()
				model := s.Model
				s.Unlock()
				if model != "" {
					statusText = model + "  ·  " + statusText
				}
			}
			return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Caption(st.matTheme, statusText)
				lbl.Color = st.theme.Palette.TextSecondary
				return lbl.Layout(gtx)
			})
		}),
		// Timeline
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if st.sess == nil {
				return layout.Dimensions{}
			}
			return st.timeline.Layout(gtx, st.sess.State)
		}),
		// Prompt bar
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			status := session.StatusIdle
			if st.sess != nil {
				st.sess.State.Lock()
				status = st.sess.State.Status
				st.sess.State.Unlock()
			}
			dims, submitted := st.prompt.Layout(gtx, status)
			if submitted != "" && st.sess != nil {
				if err := st.sess.SendPrompt(submitted); err != nil {
					log.Printf("send prompt error: %v", err)
				}
				st.updateSidebarSessionStatus("thinking")
			}
			return dims
		}),
	)
}

func (st *appState) startSession() {
	sess, err := session.New(session.Config{
		ID:          "s1",
		WorkspaceID: st.activeWs.ID,
		PiPath:      st.cfg.PiPath,
		WorkDir:     st.activeWs.Path,
		Model:       st.cfg.DefaultModel,
	})
	if err != nil {
		st.sessErr = err.Error()
		return
	}
	st.sess = sess
	st.started = true
	st.sessErr = ""
	st.updateSidebarSessionStatus("waiting")

	// Request initial state + commands for autocomplete
	sess.RequestState()
	sess.SendCommand(session.NewGetCommandsCmd())

	// Save session metadata
	st.activeWs.Sessions = append(st.activeWs.Sessions, workspace.SessionMeta{
		ID:        sess.ID,
		CreatedAt: st.activeWs.CreatedAt,
	})
	workspace.Save(st.activeWs)
}

func (st *appState) refreshSidebar() {
	var entries []ui.WorkspaceEntry
	for _, ws := range st.workspaces {
		entries = append(entries, ui.WorkspaceEntry{
			Name:     ws.Name,
			Path:     shortPath(ws.Path),
			Sessions: len(ws.Sessions),
			Status:   "idle",
		})
	}
	st.sidebar.SetWorkspaces(entries)
}

func (st *appState) updateSessionStatus() {
	if st.sess == nil {
		return
	}
	st.sess.State.Lock()
	status := st.sess.State.Status.String()
	st.sess.State.Unlock()
	st.updateSidebarSessionStatus(status)

	// Update session cards
	model := ""
	st.sess.State.Lock()
	model = st.sess.State.Model
	st.sess.State.Unlock()

	st.sidebar.SetSessions(st.sidebar.ActiveWorkspaceIndex(), []ui.SessionEntry{
		{ID: st.sess.ID, Model: model, Status: status},
	})
}

func (st *appState) updateSidebarSessionStatus(status string) {
	if len(st.workspaces) == 0 {
		return
	}
	ws := st.workspaces[st.sidebar.ActiveWorkspaceIndex()]
	entries := []ui.WorkspaceEntry{}
	for i, w := range st.workspaces {
		s := "idle"
		if i == st.sidebar.ActiveWorkspaceIndex() {
			s = status
		}
		entries = append(entries, ui.WorkspaceEntry{
			Name:     w.Name,
			Path:     shortPath(w.Path),
			Sessions: len(ws.Sessions),
			Status:   s,
		})
	}
	st.sidebar.SetWorkspaces(entries)
}

func (st *appState) checkExtensionUI() {
	if st.sess == nil {
		return
	}
	// Check last event for extension UI requests
	st.sess.State.Lock()
	timeline := st.sess.State.Timeline
	st.sess.State.Unlock()

	// Scan for pending extension UI events (simplified - in production would track pending requests)
	_ = timeline
}

func (st *appState) refreshTasks() {
	if st.brClient == nil || !st.brClient.IsInitialized() {
		return
	}
	tasks, err := st.brClient.List()
	if err != nil {
		st.taskPanel.SetError(err.Error())
		return
	}
	st.taskPanel.SetTasks(tasks)
}

// processKeys registers key filters and dispatches actions.
func (st *appState) processKeys(gtx layout.Context, w *app.Window) {
	// Register global key event handler
	area := clip.Rect(image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y)).Push(gtx.Ops)
	event.Op(gtx.Ops, &st.started)
	area.Pop()

	// Drain key events with a catch-all filter scoped to our tag
	for {
		ev, ok := gtx.Event(
			key.Filter{Focus: &st.started, Name: ""},
		)
		if !ok {
			break
		}
		if ke, ok := ev.(key.Event); ok {
			action := apptheme.MatchAction(ke)
			if action != apptheme.ActionNone {
				st.handleAction(action, w)
			}
		}
	}
}

func (st *appState) handleAction(action apptheme.Action, w *app.Window) {
	switch action {

	// --- Pi passthrough: translate to RPC commands ---

	case apptheme.ActionAbort:
		if st.sess != nil && st.sess.Running() {
			st.sess.SendCommand(session.NewAbortCmd())
		}

	case apptheme.ActionCycleModel:
		if st.sess != nil {
			st.sess.SendCommand(session.CycleModelCmd{Type: "cycle_model"})
		}

	case apptheme.ActionCycleModelBackward:
		// Pi doesn't have a backward cycle via RPC yet — use forward
		if st.sess != nil {
			st.sess.SendCommand(session.CycleModelCmd{Type: "cycle_model"})
		}

	case apptheme.ActionCycleThinking:
		if st.sess != nil {
			st.sess.SendCommand(session.CycleThinkingLevelCmd{Type: "cycle_thinking_level"})
		}

	case apptheme.ActionSelectModel:
		// TODO: show model picker UI, for now cycle
		if st.sess != nil {
			st.sess.SendCommand(session.CycleModelCmd{Type: "cycle_model"})
		}

	// --- Session lifecycle ---

	case apptheme.ActionClearEditor:
		now := time.Now()
		if now.Sub(st.lastCtrlC) < time.Second {
			// Double Ctrl+C within 1s → kill session
			st.handleAction(apptheme.ActionKillSession, w)
			st.lastCtrlC = time.Time{} // reset
		} else {
			st.lastCtrlC = now
			// Single Ctrl+C → clear prompt bar (editor text is empty after this)
			// The editor widget handles Ctrl+C natively for copy,
			// but if nothing is selected this acts as clear
		}

	case apptheme.ActionKillSession:
		if st.sess != nil {
			st.sess.Stop()
			st.sess = nil
			st.started = false
			st.updateSidebarSessionStatus("idle")
		}

	case apptheme.ActionCloseSession:
		if st.sess != nil {
			st.sess.Stop()
			st.sess = nil
			st.started = false
			st.updateSidebarSessionStatus("idle")
		}

	case apptheme.ActionNewSession:
		if !st.started {
			st.startSession()
		}

	// --- Pier-only UI ---

	case apptheme.ActionCollapseTools:
		if st.timeline != nil {
			st.timeline.CollapseAll()
		}

	case apptheme.ActionToggleTaskPanel:
		st.taskPanel.Toggle()

	case apptheme.ActionRefreshTasks:
		st.refreshTasks()

	case apptheme.ActionFocusPrompt:
		// Focus is handled by the editor widget — would need gtx to call Focus()
		// For now, this is a no-op; the prompt bar is always accessible

	case apptheme.ActionCopyLastMessage:
		if st.sess != nil {
			st.sess.State.Lock()
			tl := st.sess.State.Timeline
			for i := len(tl) - 1; i >= 0; i-- {
				if tl[i].Kind == session.EntryAssistantMessage && tl[i].AssistantText != "" {
					w.Option(app.Title("Pier — copied"))
					// Gio clipboard: would need clipboard.WriteOp
					break
				}
			}
			st.sess.State.Unlock()
		}
	}
}

func layoutDivider(gtx layout.Context, theme apptheme.Theme) layout.Dimensions {
	width := gtx.Dp(unit.Dp(1))
	rect := image.Rect(0, 0, width, gtx.Constraints.Max.Y)
	paint.FillShape(gtx.Ops, theme.Palette.Border, clip.Rect(rect).Op())
	return layout.Dimensions{Size: image.Pt(width, gtx.Constraints.Max.Y)}
}

func shortPath(p string) string {
	home, _ := os.UserHomeDir()
	if home != "" {
		if rel, err := filepath.Rel(home, p); err == nil {
			return "~/" + rel
		}
	}
	return p
}
