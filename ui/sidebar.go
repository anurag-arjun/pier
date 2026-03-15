package ui

import (
	"fmt"
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	apptheme "github.com/lighto/pier/app"
	"github.com/lighto/pier/ui/widgets"
)

// SessionEntry is a session card in the sidebar.
type SessionEntry struct {
	ID     string
	Model  string
	Status string
	TaskID string
	Unread bool
}

// WorkspaceEntry is a sidebar item.
type WorkspaceEntry struct {
	Name     string
	Path     string
	Sessions int
	Status   string
}

// Sidebar is the left navigation panel.
type Sidebar struct {
	theme    apptheme.Theme
	matTheme *material.Theme

	workspaces  []WorkspaceEntry
	activeWsIdx int
	wsClicks    []widget.Clickable
	wsHovers    []widgets.HoverState
	addWsBtn    widget.Clickable

	sessions      map[int][]SessionEntry
	activeSession map[int]int

	OnAddWorkspace func()
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

func (s *Sidebar) SetWorkspaces(ws []WorkspaceEntry) {
	s.workspaces = ws
	for len(s.wsClicks) < len(ws) {
		s.wsClicks = append(s.wsClicks, widget.Clickable{})
	}
	for len(s.wsHovers) < len(ws) {
		s.wsHovers = append(s.wsHovers, widgets.HoverState{})
	}
}

func (s *Sidebar) SetSessions(wsIdx int, sessions []SessionEntry) {
	s.sessions[wsIdx] = sessions
}

func (s *Sidebar) ActiveWorkspaceIndex() int  { return s.activeWsIdx }
func (s *Sidebar) ActiveSessionIndex() int    { return s.activeSession[s.activeWsIdx] }

func (s *Sidebar) Layout(gtx layout.Context) layout.Dimensions {
	width := gtx.Dp(s.theme.Space.SidebarWidth)
	gtx.Constraints.Min.X = width
	gtx.Constraints.Max.X = width

	// Background
	paint.FillShape(gtx.Ops, s.theme.Palette.Surface, clip.Rect(image.Rect(0, 0, width, gtx.Constraints.Max.Y)).Op())

	for i := range s.workspaces {
		if i < len(s.wsClicks) && s.wsClicks[i].Clicked(gtx) {
			s.activeWsIdx = i
		}
	}

	return layout.Inset{Top: unit.Dp(12), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Title
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(16), Right: unit.Dp(16), Bottom: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(s.matTheme, s.theme.Typo.H2, "Pier")
					lbl.Font.Weight = font.Bold
					lbl.Color = s.theme.Palette.Text
					return lbl.Layout(gtx)
				})
			}),
			// Section header
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(16), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(s.matTheme, s.theme.Typo.Caption, "WORKSPACES")
					lbl.Color = s.theme.Palette.TextTertiary
					lbl.Font.Weight = font.Medium
					return lbl.Layout(gtx)
				})
			}),
			// Workspace list
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				list := layout.List{Axis: layout.Vertical}
				return list.Layout(gtx, len(s.workspaces), func(gtx layout.Context, i int) layout.Dimensions {
					return s.layoutWorkspaceItem(gtx, i)
				})
			}),
			// Add workspace button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8), Top: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					if s.addWsBtn.Clicked(gtx) && s.OnAddWorkspace != nil {
						s.OnAddWorkspace()
					}
					btn := material.Button(s.matTheme, &s.addWsBtn, "+ workspace")
					btn.Background = s.theme.Palette.SurfaceAlt
					btn.Color = s.theme.Palette.TextSecondary
					btn.TextSize = s.theme.Typo.BodySmall
					btn.Inset = layout.Inset{Left: unit.Dp(12), Right: unit.Dp(12), Top: unit.Dp(6), Bottom: unit.Dp(6)}
					return btn.Layout(gtx)
				})
			}),
		)
	})
}

func (s *Sidebar) layoutWorkspaceItem(gtx layout.Context, i int) layout.Dimensions {
	ws := s.workspaces[i]
	isActive := i == s.activeWsIdx

	return material.Clickable(gtx, &s.wsClicks[i], func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Left: unit.Dp(4), Right: unit.Dp(4), Top: unit.Dp(1), Bottom: unit.Dp(1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			// Track hover
			s.wsHovers[i].Update(gtx)
			hovered := s.wsHovers[i].Hovered()

			return layout.Stack{}.Layout(gtx,
				// Background: hover or selected
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					if !isActive && !hovered {
						return layout.Dimensions{}
					}
					rr := gtx.Dp(unit.Dp(4))
					size := gtx.Constraints.Min
					if size.X == 0 {
						size.X = gtx.Constraints.Max.X
					}
					if size.Y == 0 {
						size.Y = gtx.Constraints.Max.Y
					}
					bg := s.theme.Palette.SurfaceAlt
					if hovered && !isActive {
						bg = widgets.WithAlpha(s.theme.Palette.SurfaceAlt, 160)
					}
					paint.FillShape(gtx.Ops, bg, clip.UniformRRect(image.Rect(0, 0, size.X, size.Y), rr).Op(gtx.Ops))
					return layout.Dimensions{Size: size}
				}),
				// Left accent bar (selected)
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{}.Layout(gtx,
						// Accent bar
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if !isActive {
								return layout.Dimensions{Size: image.Pt(gtx.Dp(unit.Dp(3)), 0)}
							}
							barW := gtx.Dp(unit.Dp(3))
							barH := gtx.Constraints.Max.Y
							if barH <= 0 {
								barH = gtx.Dp(unit.Dp(40))
							}
							rr := gtx.Dp(unit.Dp(2))
							paint.FillShape(gtx.Ops, s.theme.Palette.Accent, clip.UniformRRect(image.Rect(0, 4, barW, barH-4), rr).Op(gtx.Ops))
							return layout.Dimensions{Size: image.Pt(barW, barH)}
						}),
						// Content
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8), Top: unit.Dp(8), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									// Name + status dot
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
											layout.Rigid(func(gtx layout.Context) layout.Dimensions {
												return layoutStatusDot(gtx, s.theme, ws.Status)
											}),
											layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
											layout.Rigid(func(gtx layout.Context) layout.Dimensions {
												lbl := material.Label(s.matTheme, s.theme.Typo.Body, ws.Name)
												lbl.Font.Weight = font.Medium
												lbl.Color = s.theme.Palette.Text
												return lbl.Layout(gtx)
											}),
										)
									}),
									// Path + session count
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return layout.Inset{Top: unit.Dp(2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
											text := ws.Path
											if ws.Sessions > 0 {
												text += fmt.Sprintf("  ·  %d sessions", ws.Sessions)
											}
											lbl := material.Label(s.matTheme, s.theme.Typo.Caption, text)
											lbl.Color = s.theme.Palette.TextTertiary
											return lbl.Layout(gtx)
										})
									}),
									// Session cards
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										if !isActive {
											return layout.Dimensions{}
										}
										sessions := s.sessions[i]
										if len(sessions) == 0 {
											return layout.Dimensions{}
										}
										return layout.Inset{Top: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
											return s.layoutSessionCards(gtx, sessions)
										})
									}),
								)
							})
						}),
					)
				}),
			)
		})
	})
}

func (s *Sidebar) layoutSessionCards(gtx layout.Context, sessions []SessionEntry) layout.Dimensions {
	list := layout.List{Axis: layout.Vertical}
	return list.Layout(gtx, len(sessions), func(gtx layout.Context, i int) layout.Dimensions {
		sess := sessions[i]
		return layout.Inset{Top: unit.Dp(2), Bottom: unit.Dp(2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutStatusDot(gtx, s.theme, sess.Status)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
				// Model as pill badge
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					model := sess.Model
					if model == "" {
						model = "session"
					}
					return s.layoutPillBadge(gtx, model)
				}),
				// Task ID
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if sess.TaskID == "" {
						return layout.Dimensions{}
					}
					return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(s.matTheme, s.theme.Typo.Caption, sess.TaskID)
						lbl.Color = s.theme.Palette.TextTertiary
						lbl.Font.Typeface = s.theme.Typo.MonoFace
						return lbl.Layout(gtx)
					})
				}),
				// Unread dot
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !sess.Unread {
						return layout.Dimensions{}
					}
					return layout.Inset{Left: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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

func (s *Sidebar) layoutPillBadge(gtx layout.Context, text string) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			rr := gtx.Dp(s.theme.Space.BadgeRadius)
			size := gtx.Constraints.Min
			paint.FillShape(gtx.Ops, s.theme.Palette.SurfaceAlt, clip.UniformRRect(image.Rect(0, 0, size.X, size.Y), rr).Op(gtx.Ops))
			return layout.Dimensions{Size: size}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: unit.Dp(6), Right: unit.Dp(6), Top: unit.Dp(2), Bottom: unit.Dp(2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(s.matTheme, s.theme.Typo.MonoSmall, text)
				lbl.Color = s.theme.Palette.TextSecondary
				lbl.Font.Typeface = s.theme.Typo.MonoFace
				return lbl.Layout(gtx)
			})
		}),
	)
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
		c = theme.Palette.TextTertiary
	}
	defer clip.Ellipse{Max: image.Pt(size, size)}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, c)
	return layout.Dimensions{Size: image.Pt(size, size)}
}
