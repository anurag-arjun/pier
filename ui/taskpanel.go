package ui

import (
	"fmt"
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	apptheme "github.com/lighto/pier/app"
	"github.com/lighto/pier/br"
	"github.com/lighto/pier/ui/widgets"
)

// TaskPanel displays br tasks grouped by status.
type TaskPanel struct {
	theme    apptheme.Theme
	matTheme *material.Theme

	visible    bool
	tasks      []br.Task
	brErr      string
	list       widget.List
	refreshBtn widget.Clickable
	taskHovers []widgets.HoverState

	OnTaskClick func(task br.Task)
}

func NewTaskPanel(theme apptheme.Theme, matTheme *material.Theme) *TaskPanel {
	return &TaskPanel{
		theme:    theme,
		matTheme: matTheme,
		visible:  true,
		list:     widget.List{List: layout.List{Axis: layout.Vertical}},
	}
}

func (tp *TaskPanel) SetVisible(v bool) { tp.visible = v }
func (tp *TaskPanel) Visible() bool     { return tp.visible }
func (tp *TaskPanel) Toggle()           { tp.visible = !tp.visible }

func (tp *TaskPanel) SetTasks(tasks []br.Task) {
	tp.tasks = tasks
	tp.brErr = ""
	for len(tp.taskHovers) < len(tasks) {
		tp.taskHovers = append(tp.taskHovers, widgets.HoverState{})
	}
}

func (tp *TaskPanel) SetError(err string) { tp.brErr = err; tp.tasks = nil }

func (tp *TaskPanel) RefreshClicked(gtx layout.Context) bool {
	return tp.refreshBtn.Clicked(gtx)
}

func (tp *TaskPanel) Layout(gtx layout.Context) layout.Dimensions {
	if !tp.visible {
		return layout.Dimensions{}
	}

	width := gtx.Dp(unit.Dp(260))
	gtx.Constraints.Min.X = width
	gtx.Constraints.Max.X = width

	paint.FillShape(gtx.Ops, tp.theme.Palette.Surface, clip.Rect(image.Rect(0, 0, width, gtx.Constraints.Max.Y)).Op())

	return layout.Inset{Top: unit.Dp(12), Left: unit.Dp(12), Right: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Header
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(tp.matTheme, tp.theme.Typo.Caption, "TASKS")
						lbl.Color = tp.theme.Palette.TextTertiary
						lbl.Font.Weight = font.Medium
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(tp.matTheme, &tp.refreshBtn, "↻")
						btn.Background = tp.theme.Palette.SurfaceAlt
						btn.Color = tp.theme.Palette.TextSecondary
						btn.TextSize = tp.theme.Typo.BodySmall
						btn.Inset = layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8), Top: unit.Dp(3), Bottom: unit.Dp(3)}
						return btn.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			// Error state
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if tp.brErr == "" {
					return layout.Dimensions{}
				}
				return tp.layoutError(gtx)
			}),
			// Task list
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				if len(tp.tasks) == 0 && tp.brErr == "" {
					lbl := material.Label(tp.matTheme, tp.theme.Typo.BodySmall, "No tasks yet.")
					lbl.Color = tp.theme.Palette.TextTertiary
					return lbl.Layout(gtx)
				}
				return tp.layoutTaskList(gtx)
			}),
		)
	})
}

func (tp *TaskPanel) layoutError(gtx layout.Context) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			rr := gtx.Dp(tp.theme.Space.BorderRadius)
			size := gtx.Constraints.Min
			if size.X == 0 {
				size.X = gtx.Constraints.Max.X
			}
			widgets.DrawBorderedRect(gtx, tp.theme.Palette.ToolBlockBg, tp.theme.Palette.Error, size, rr, 1)
			// Left red accent
			barW := gtx.Dp(unit.Dp(3))
			paint.FillShape(gtx.Ops, tp.theme.Palette.Error, clip.Rect(image.Rect(0, 0, barW, size.Y)).Op())
			return layout.Dimensions{Size: size}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: unit.Dp(12), Right: unit.Dp(8), Top: unit.Dp(8), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(tp.matTheme, tp.theme.Typo.BodySmall, tp.brErr)
				lbl.Color = tp.theme.Palette.Error
				return lbl.Layout(gtx)
			})
		}),
	)
}

func (tp *TaskPanel) layoutTaskList(gtx layout.Context) layout.Dimensions {
	ready, inProgress, blocked := tp.groupTasks()

	type section struct {
		label string
		tasks []br.Task
	}
	sections := []section{
		{"READY (%d)", ready},
		{"IN PROGRESS (%d)", inProgress},
		{"BLOCKED (%d)", blocked},
	}

	var items []layout.FlexChild
	for _, sec := range sections {
		sec := sec
		if len(sec.tasks) == 0 {
			continue
		}
		items = append(items, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(tp.matTheme, tp.theme.Typo.Caption, fmt.Sprintf(sec.label, len(sec.tasks)))
				lbl.Color = tp.theme.Palette.TextTertiary
				lbl.Font.Weight = font.Medium
				return lbl.Layout(gtx)
			})
		}))
		for _, task := range sec.tasks {
			task := task
			items = append(items, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return tp.layoutTaskCard(gtx, task)
			}))
		}
	}

	if len(items) == 0 {
		return layout.Dimensions{}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, items...)
}

func (tp *TaskPanel) groupTasks() (ready, inProgress, blocked []br.Task) {
	for _, t := range tp.tasks {
		switch t.Status {
		case "in_progress":
			inProgress = append(inProgress, t)
		case "open":
			ready = append(ready, t)
		default:
			blocked = append(blocked, t)
		}
	}
	return
}

func (tp *TaskPanel) layoutTaskCard(gtx layout.Context, task br.Task) layout.Dimensions {
	// Find hover index
	idx := -1
	for i, t := range tp.tasks {
		if t.ID == task.ID {
			idx = i
			break
		}
	}

	isBlocked := task.Status != "open" && task.Status != "in_progress"

	return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		borderColor := tp.theme.Palette.BorderSubtle
		if idx >= 0 && idx < len(tp.taskHovers) {
			tp.taskHovers[idx].Update(gtx)
			if tp.taskHovers[idx].Hovered() {
				borderColor = tp.theme.Palette.Border
			}
		}

		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				rr := gtx.Dp(tp.theme.Space.BorderRadius)
				size := gtx.Constraints.Min
				if size.X == 0 {
					size.X = gtx.Constraints.Max.X
				}
				if size.Y == 0 {
					size.Y = gtx.Constraints.Max.Y
				}
				bg := tp.theme.Palette.SurfaceAlt
				if isBlocked {
					bg = widgets.MulAlpha(bg, 128)
				}
				widgets.DrawBorderedRect(gtx, bg, borderColor, size, rr, 1)
				return layout.Dimensions{Size: size}
			}),
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(10), Right: unit.Dp(10), Top: unit.Dp(8), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						// Priority dot + ID
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
								// Priority dot
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									c := tp.theme.Palette.PriorityP3
									if task.Priority <= 1 {
										c = tp.theme.Palette.PriorityP1
									} else if task.Priority == 2 {
										c = tp.theme.Palette.PriorityP2
									}
									size := gtx.Dp(unit.Dp(8))
									defer clip.Ellipse{Max: image.Pt(size, size)}.Push(gtx.Ops).Pop()
									paint.Fill(gtx.Ops, c)
									return layout.Dimensions{Size: image.Pt(size, size)}
								}),
								layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
								// ID
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									id := task.ID
									if len(id) > 12 {
										id = id[len(id)-8:]
									}
									lbl := material.Label(tp.matTheme, tp.theme.Typo.Caption, id)
									lbl.Color = tp.theme.Palette.TextTertiary
									lbl.Font.Typeface = tp.theme.Typo.MonoFace
									return lbl.Layout(gtx)
								}),
							)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
						// Title
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(tp.matTheme, tp.theme.Typo.BodySmall, task.Title)
							lbl.Font.Weight = font.Medium
							lbl.Color = tp.theme.Palette.Text
							if isBlocked {
								lbl.Color = tp.theme.Palette.TextSecondary
							}
							lbl.MaxLines = 2
							return lbl.Layout(gtx)
						}),
					)
				})
			}),
		)
	})
}
