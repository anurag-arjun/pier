package ui

import (
	"fmt"
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	apptheme "github.com/lighto/pier/app"
	"github.com/lighto/pier/br"
)

// TaskPanel displays br tasks grouped by status.
type TaskPanel struct {
	theme    apptheme.Theme
	matTheme *material.Theme

	visible   bool
	tasks     []br.Task // cached task list
	brErr     string    // error message (e.g., "br not found")
	list      widget.List

	refreshBtn widget.Clickable
	taskClicks []widget.Clickable

	// Callback when a task card is clicked
	OnTaskClick func(task br.Task)
}

// NewTaskPanel creates a task panel.
func NewTaskPanel(theme apptheme.Theme, matTheme *material.Theme) *TaskPanel {
	return &TaskPanel{
		theme:    theme,
		matTheme: matTheme,
		visible:  true,
		list:     widget.List{List: layout.List{Axis: layout.Vertical}},
	}
}

// SetVisible shows/hides the panel.
func (tp *TaskPanel) SetVisible(v bool) { tp.visible = v }

// Visible returns whether the panel is visible.
func (tp *TaskPanel) Visible() bool { return tp.visible }

// Toggle toggles panel visibility.
func (tp *TaskPanel) Toggle() { tp.visible = !tp.visible }

// SetTasks updates the cached task list.
func (tp *TaskPanel) SetTasks(tasks []br.Task) {
	tp.tasks = tasks
	tp.brErr = ""
	for len(tp.taskClicks) < len(tasks) {
		tp.taskClicks = append(tp.taskClicks, widget.Clickable{})
	}
}

// SetError sets an error message (e.g., br not found).
func (tp *TaskPanel) SetError(err string) {
	tp.brErr = err
	tp.tasks = nil
}

// RefreshClicked returns true if the refresh button was clicked.
func (tp *TaskPanel) RefreshClicked(gtx layout.Context) bool {
	return tp.refreshBtn.Clicked(gtx)
}

// Layout renders the task panel.
func (tp *TaskPanel) Layout(gtx layout.Context) layout.Dimensions {
	if !tp.visible {
		return layout.Dimensions{}
	}

	width := gtx.Dp(unit.Dp(280))
	gtx.Constraints.Min.X = width
	gtx.Constraints.Max.X = width

	// Background
	rect := image.Rect(0, 0, width, gtx.Constraints.Max.Y)
	paint.FillShape(gtx.Ops, tp.theme.Palette.Surface, clip.Rect(rect).Op())

	// Check task clicks
	for i := range tp.tasks {
		if i < len(tp.taskClicks) && tp.taskClicks[i].Clicked(gtx) {
			if tp.OnTaskClick != nil {
				tp.OnTaskClick(tp.tasks[i])
			}
		}
	}

	return layout.Inset{
		Top: unit.Dp(8), Left: unit.Dp(8), Right: unit.Dp(8),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Header
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Caption(tp.matTheme, "TASKS")
						lbl.Color = tp.theme.Palette.TextSecondary
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(tp.matTheme, &tp.refreshBtn, "↻")
						btn.Background = tp.theme.Palette.SurfaceAlt
						btn.Color = tp.theme.Palette.TextSecondary
						btn.Inset = layout.Inset{
							Left: unit.Dp(8), Right: unit.Dp(8),
							Top: unit.Dp(2), Bottom: unit.Dp(2),
						}
						return btn.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			// Error state
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if tp.brErr == "" {
					return layout.Dimensions{}
				}
				lbl := material.Body2(tp.matTheme, tp.brErr)
				lbl.Color = tp.theme.Palette.Error
				return lbl.Layout(gtx)
			}),
			// Task list
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return tp.layoutTaskList(gtx)
			}),
		)
	})
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
		// Section header
		items = append(items, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Caption(tp.matTheme, fmt.Sprintf(sec.label, len(sec.tasks)))
				lbl.Color = tp.theme.Palette.TextSecondary
				lbl.Font.Weight = 700
				return lbl.Layout(gtx)
			})
		}))
		// Task cards
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
			// "deferred" or other → blocked-ish
			blocked = append(blocked, t)
		}
	}
	return
}

func (tp *TaskPanel) layoutTaskCard(gtx layout.Context, task br.Task) layout.Dimensions {
	return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				rr := gtx.Dp(unit.Dp(4))
				bounds := image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y)
				rrect := clip.RRect{Rect: bounds, NE: rr, NW: rr, SE: rr, SW: rr}
				defer rrect.Push(gtx.Ops).Pop()
				paint.Fill(gtx.Ops, tp.theme.Palette.SurfaceAlt)
				return layout.Dimensions{Size: bounds.Max}
			}),
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Left: unit.Dp(8), Right: unit.Dp(8),
					Top: unit.Dp(6), Bottom: unit.Dp(6),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						// Priority + ID + Title
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
								// Priority badge
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									c := tp.theme.Palette.PriorityP3
									if task.Priority <= 1 {
										c = tp.theme.Palette.PriorityP1
									} else if task.Priority == 2 {
										c = tp.theme.Palette.PriorityP2
									}
									lbl := material.Caption(tp.matTheme, fmt.Sprintf("P%d", task.Priority))
									lbl.Color = c
									lbl.Font.Weight = 700
									return lbl.Layout(gtx)
								}),
								layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
								// ID
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									id := task.ID
									if len(id) > 12 {
										id = id[len(id)-8:]
									}
									lbl := material.Caption(tp.matTheme, id)
									lbl.Color = tp.theme.Palette.TextSecondary
									return lbl.Layout(gtx)
								}),
							)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(2)}.Layout),
						// Title
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body2(tp.matTheme, task.Title)
							lbl.Color = tp.theme.Palette.Text
							lbl.MaxLines = 2
							return lbl.Layout(gtx)
						}),
					)
				})
			}),
		)
	})
}
