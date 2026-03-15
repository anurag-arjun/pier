package ui

import (
	"os"
	"os/exec"
	"runtime"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/richtext"

	apptheme "github.com/lighto/pier/app"
	"github.com/lighto/pier/ui/widgets"
)

// PlanPanel displays plan.md as rendered markdown.
type PlanPanel struct {
	theme    apptheme.Theme
	matTheme *material.Theme
	mdRender *widgets.MarkdownRenderer

	planPath    string
	planContent string
	planLoaded  bool
	richState   richtext.InteractiveText
	list        widget.List

	editBtn widget.Clickable
}

// NewPlanPanel creates a plan panel.
func NewPlanPanel(theme apptheme.Theme, matTheme *material.Theme) *PlanPanel {
	return &PlanPanel{
		theme:    theme,
		matTheme: matTheme,
		mdRender: widgets.NewMarkdownRenderer(
			matTheme.Shaper,
			theme.Palette.Text,
			theme.Palette.Accent,
		),
		list: widget.List{List: layout.List{Axis: layout.Vertical}},
	}
}

// SetPlanPath sets the plan file path and triggers a reload.
func (pp *PlanPanel) SetPlanPath(path string) {
	pp.planPath = path
	pp.Reload()
}

// Reload reads plan.md from disk.
func (pp *PlanPanel) Reload() {
	if pp.planPath == "" {
		pp.planContent = ""
		pp.planLoaded = false
		return
	}
	data, err := os.ReadFile(pp.planPath)
	if err != nil {
		pp.planContent = ""
		pp.planLoaded = false
		return
	}
	pp.planContent = string(data)
	pp.planLoaded = true
}

// Layout renders the plan panel.
func (pp *PlanPanel) Layout(gtx layout.Context) layout.Dimensions {
	// Check edit button
	if pp.editBtn.Clicked(gtx) {
		pp.openInEditor()
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Caption(pp.matTheme, "PLAN")
					lbl.Color = pp.theme.Palette.TextSecondary
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !pp.planLoaded {
						return layout.Dimensions{}
					}
					btn := material.Button(pp.matTheme, &pp.editBtn, "Open in $EDITOR")
					btn.Background = pp.theme.Palette.SurfaceAlt
					btn.Color = pp.theme.Palette.TextSecondary
					btn.Inset = layout.Inset{
						Left: unit.Dp(8), Right: unit.Dp(8),
						Top: unit.Dp(2), Bottom: unit.Dp(2),
					}
					return btn.Layout(gtx)
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		// Content
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if !pp.planLoaded {
				lbl := material.Body2(pp.matTheme, "No plan yet. Use Discover to generate one.")
				lbl.Color = pp.theme.Palette.TextSecondary
				return lbl.Layout(gtx)
			}
			return pp.mdRender.Layout(gtx, &pp.richState, pp.planContent)
		}),
	)
}

func (pp *PlanPanel) openInEditor() {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		switch runtime.GOOS {
		case "darwin":
			editor = "open"
		case "windows":
			editor = "notepad"
		default:
			editor = "xdg-open"
		}
	}
	cmd := exec.Command(editor, pp.planPath)
	cmd.Start() // fire and forget
}
