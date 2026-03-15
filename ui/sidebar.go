package ui

import (
	"fmt"
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	apptheme "github.com/lighto/pier/app"
)

// SessionEntry is a session card in the sidebar.
type SessionEntry struct {
	ID     string
	Model  string
	Status string // "thinking"|"streaming"|"waiting"|"error"|"idle"
	TaskID string // linked task ID if any
	Unread bool   // has unread agent_end while unfocused
}

// WorkspaceEntry is a sidebar item.
type WorkspaceEntry struct {
	Name     string
	Path     string
	Sessions int
	Status   string // most active session status
}

// Sidebar is the left navigation panel.
type Sidebar struct {
	theme         apptheme.Theme
	matTheme      *material.Theme

	workspaces    []WorkspaceEntry
	activeWsIdx   int
	wsClicks      []widget.Clickable
	addWsBtn      widget.Clickable

	// Per-workspace session data
	sessions      map[int][]SessionEntry // workspace index → sessions
	activeSession map[int]int            // workspace index → active session index
	sessionClicks []widget.Clickable     // flat list of session clickables
	addSessBtn    widget.Clickable

	OnAddWorkspace func()
	OnAddSession   func(wsIdx int)
}

// NewSidebar creates the sidebar.
func NewSidebar(theme apptheme.Theme, matTheme *material.Theme) *Sidebar {
	return &Sidebar{
		theme:         theme,
		matTheme:      matTheme,
		sessions:      make(map[int][]SessionEntry),
		activeSession: make(map[int]int),
	}
}

// SetWorkspaces sets the workspace list.
func (s *Sidebar) SetWorkspaces(ws []WorkspaceEntry) {
	s.workspaces = ws
	for len(s.wsClicks) < len(ws) {
		s.wsClicks = append(s.wsClicks, widget.Clickable{})
	}
}

// SetSessions sets sessions for a workspace.
func (s *Sidebar) SetSessions(wsIdx int, sessions []SessionEntry) {
	s.sessions[wsIdx] = sessions
}

// ActiveWorkspaceIndex returns the selected workspace.
func (s *Sidebar) ActiveWorkspaceIndex() int {
	return s.activeWsIdx
}

// ActiveSessionIndex returns the focused session in the active workspace.
func (s *Sidebar) ActiveSessionIndex() int {
	return s.activeSession[s.activeWsIdx]
}

// Layout renders the sidebar.
func (s *Sidebar) Layout(gtx layout.Context) layout.Dimensions {
	width := gtx.Dp(s.theme.Space.SidebarWidth)
	gtx.Constraints.Min.X = width
	gtx.Constraints.Max.X = width

	// Background
	rect := image.Rect(0, 0, width, gtx.Constraints.Max.Y)
	paint.FillShape(gtx.Ops, s.theme.Palette.Surface, clip.Rect(rect).Op())

	// Process clicks
	for i := range s.workspaces {
		if i < len(s.wsClicks) && s.wsClicks[i].Clicked(gtx) {
			s.activeWsIdx = i
		}
	}

	return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Title
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(12), Right: unit.Dp(12), Bottom: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.H4(s.matTheme, "Pier")
					lbl.Color = s.theme.Palette.Text
					return lbl.Layout(gtx)
				})
			}),
			// Section header
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(12), Bottom: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Caption(s.matTheme, "WORKSPACES")
					lbl.Color = s.theme.Palette.TextSecondary
					return lbl.Layout(gtx)
				})
			}),
			// Workspace list
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return s.layoutWorkspaceList(gtx)
			}),
			// Add workspace button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8), Top: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					if s.addWsBtn.Clicked(gtx) && s.OnAddWorkspace != nil {
						s.OnAddWorkspace()
					}
					btn := material.Button(s.matTheme, &s.addWsBtn, "+ workspace")
					btn.Background = s.theme.Palette.SurfaceAlt
					btn.Color = s.theme.Palette.TextSecondary
					btn.Inset = layout.Inset{
						Left: unit.Dp(8), Right: unit.Dp(8),
						Top: unit.Dp(4), Bottom: unit.Dp(4),
					}
					return btn.Layout(gtx)
				})
			}),
		)
	})
}

func (s *Sidebar) layoutWorkspaceList(gtx layout.Context) layout.Dimensions {
	list := layout.List{Axis: layout.Vertical}
	return list.Layout(gtx, len(s.workspaces), func(gtx layout.Context, i int) layout.Dimensions {
		ws := s.workspaces[i]
		isActive := i == s.activeWsIdx

		return material.Clickable(gtx, &s.wsClicks[i], func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: unit.Dp(4), Right: unit.Dp(4), Top: unit.Dp(1), Bottom: unit.Dp(1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Stack{}.Layout(gtx,
					// Active highlight
					layout.Expanded(func(gtx layout.Context) layout.Dimensions {
						if !isActive {
							return layout.Dimensions{}
						}
						rr := gtx.Dp(unit.Dp(4))
						bounds := image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y)
						rrect := clip.RRect{Rect: bounds, NE: rr, NW: rr, SE: rr, SW: rr}
						defer rrect.Push(gtx.Ops).Pop()
						paint.Fill(gtx.Ops, s.theme.Palette.SurfaceAlt)
						return layout.Dimensions{Size: bounds.Max}
					}),
					layout.Stacked(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8), Top: unit.Dp(6), Bottom: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								// Name + status dot
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return layoutStatusDot(gtx, s.theme, ws.Status)
										}),
										layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											lbl := material.Body1(s.matTheme, ws.Name)
											lbl.Color = s.theme.Palette.Text
											return lbl.Layout(gtx)
										}),
									)
								}),
								// Path + session count
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									text := ws.Path
									if ws.Sessions > 0 {
										text += fmt.Sprintf("  ·  %d sessions", ws.Sessions)
									}
									lbl := material.Caption(s.matTheme, text)
									lbl.Color = s.theme.Palette.TextSecondary
									return lbl.Layout(gtx)
								}),
								// Session cards (only for active workspace)
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									if !isActive {
										return layout.Dimensions{}
									}
									sessions := s.sessions[i]
									if len(sessions) == 0 {
										return layout.Dimensions{}
									}
									return layout.Inset{Top: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return s.layoutSessionCards(gtx, i, sessions)
									})
								}),
							)
						})
					}),
				)
			})
		})
	})
}

func (s *Sidebar) layoutSessionCards(gtx layout.Context, wsIdx int, sessions []SessionEntry) layout.Dimensions {
	list := layout.List{Axis: layout.Vertical}
	return list.Layout(gtx, len(sessions), func(gtx layout.Context, i int) layout.Dimensions {
		sess := sessions[i]
		isActive := i == s.activeSession[wsIdx]

		return layout.Inset{Top: unit.Dp(2), Bottom: unit.Dp(2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutStatusDot(gtx, s.theme, sess.Status)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					model := sess.Model
					if model == "" {
						model = "session"
					}
					lbl := material.Caption(s.matTheme, model)
					if isActive {
						lbl.Color = s.theme.Palette.Text
					} else {
						lbl.Color = s.theme.Palette.TextSecondary
					}
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if sess.TaskID == "" {
						return layout.Dimensions{}
					}
					return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Caption(s.matTheme, sess.TaskID)
						lbl.Color = s.theme.Palette.TextSecondary
						return lbl.Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !sess.Unread {
						return layout.Dimensions{}
					}
					return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						size := gtx.Dp(unit.Dp(6))
						defer clip.Ellipse{Max: image.Pt(size, size)}.Push(gtx.Ops).Pop()
						paint.Fill(gtx.Ops, s.theme.Palette.Accent)
						return layout.Dimensions{Size: image.Pt(size, size)}
					})
				}),
			)
		})
	})
}

func layoutStatusDot(gtx layout.Context, theme apptheme.Theme, status string) layout.Dimensions {
	size := gtx.Dp(unit.Dp(8))
	var c color.NRGBA
	switch status {
	case "thinking":
		c = theme.Palette.StatusThinking
	case "streaming":
		c = theme.Palette.StatusStreaming
	case "waiting":
		c = theme.Palette.StatusWaiting
	case "error":
		c = theme.Palette.StatusError
	case "compacting":
		c = theme.Palette.StatusCompacting
	case "retrying":
		c = theme.Palette.StatusRetrying
	default:
		c = theme.Palette.TextSecondary
	}
	defer clip.Ellipse{Max: image.Pt(size, size)}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, c)
	return layout.Dimensions{Size: image.Pt(size, size)}
}
