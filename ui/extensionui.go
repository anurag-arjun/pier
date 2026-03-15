package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	apptheme "github.com/lighto/pier/app"
	"github.com/lighto/pier/session"
	"github.com/lighto/pier/ui/widgets"
)

// ExtensionUIHandler renders modal dialogs for extension_ui_request events.
type ExtensionUIHandler struct {
	theme    apptheme.Theme
	matTheme *material.Theme

	request     *session.ExtensionUIRequest
	selectBtns  []widget.Clickable
	selectHover []widgets.HoverState
	confirmBtn  widget.Clickable
	cancelBtn   widget.Clickable
	inputEditor widget.Editor
	submitBtn   widget.Clickable

	OnResponse func(resp session.ExtensionUIResponse)
}

func NewExtensionUIHandler(theme apptheme.Theme, matTheme *material.Theme) *ExtensionUIHandler {
	return &ExtensionUIHandler{
		theme:    theme,
		matTheme: matTheme,
		inputEditor: widget.Editor{
			SingleLine: true,
			Submit:     true,
		},
	}
}

func (h *ExtensionUIHandler) SetRequest(req *session.ExtensionUIRequest) {
	h.request = req
	h.selectBtns = nil
	h.selectHover = nil
	if req != nil && req.Method == "select" {
		h.selectBtns = make([]widget.Clickable, len(req.Options))
		h.selectHover = make([]widgets.HoverState, len(req.Options))
	}
	if req != nil && (req.Method == "input" || req.Method == "editor") {
		h.inputEditor.SetText(req.Prefill)
	}
}

func (h *ExtensionUIHandler) HasActiveDialog() bool { return h.request != nil }

// Layout renders backdrop + centered dialog card.
func (h *ExtensionUIHandler) Layout(gtx layout.Context) (layout.Dimensions, bool) {
	if h.request == nil {
		return layout.Dimensions{}, false
	}

	// Full-screen backdrop
	dims := layout.Stack{Alignment: layout.Center}.Layout(gtx,
		// Dim backdrop
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)
			paint.FillShape(gtx.Ops, widgets.WithAlpha(h.theme.Palette.Background, 180), clip.Rect(image.Rect(0, 0, size.X, size.Y)).Op())
			return layout.Dimensions{Size: size}
		}),
		// Dialog card
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			// Constrain dialog width
			maxW := gtx.Dp(unit.Dp(480))
			minW := gtx.Dp(unit.Dp(320))
			if gtx.Constraints.Max.X > maxW {
				gtx.Constraints.Max.X = maxW
			}
			if gtx.Constraints.Min.X < minW {
				gtx.Constraints.Min.X = minW
			}

			return layout.Stack{}.Layout(gtx,
				// Card background
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					rr := gtx.Dp(unit.Dp(12))
					size := gtx.Constraints.Min
					if size.X == 0 {
						size.X = gtx.Constraints.Max.X
					}
					if size.Y == 0 {
						size.Y = gtx.Constraints.Max.Y
					}
					widgets.DrawBorderedRect(gtx, h.theme.Palette.SurfaceRaised, h.theme.Palette.Border, size, rr, 1)
					return layout.Dimensions{Size: size}
				}),
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(24)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return h.layoutDialogContent(gtx)
					})
				}),
			)
		}),
	)

	return dims, false
}

func (h *ExtensionUIHandler) layoutDialogContent(gtx layout.Context) layout.Dimensions {
	switch h.request.Method {
	case "select":
		return h.layoutSelect(gtx)
	case "confirm":
		return h.layoutConfirm(gtx)
	case "input", "editor":
		return h.layoutInput(gtx)
	default:
		h.sendCancelled()
		return layout.Dimensions{}
	}
}

func (h *ExtensionUIHandler) layoutSelect(gtx layout.Context) layout.Dimensions {
	for i := range h.request.Options {
		if i < len(h.selectBtns) && h.selectBtns[i].Clicked(gtx) {
			h.sendValue(h.request.Options[i])
		}
	}
	if h.cancelBtn.Clicked(gtx) {
		h.sendCancelled()
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(h.matTheme, h.theme.Typo.H3, h.request.Title)
			lbl.Color = h.theme.Palette.Text
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			list := layout.List{Axis: layout.Vertical}
			return list.Layout(gtx, len(h.request.Options), func(gtx layout.Context, i int) layout.Dimensions {
				if i < len(h.selectHover) {
					h.selectHover[i].Update(gtx)
				}
				return material.Clickable(gtx, &h.selectBtns[i], func(gtx layout.Context) layout.Dimensions {
					return layout.Stack{}.Layout(gtx,
						layout.Expanded(func(gtx layout.Context) layout.Dimensions {
							if i < len(h.selectHover) && h.selectHover[i].Hovered() {
								rr := gtx.Dp(unit.Dp(4))
								size := gtx.Constraints.Min
								if size.X == 0 {
									size.X = gtx.Constraints.Max.X
								}
								paint.FillShape(gtx.Ops, h.theme.Palette.SurfaceAlt, clip.UniformRRect(image.Rect(0, 0, size.X, size.Y), rr).Op(gtx.Ops))
								return layout.Dimensions{Size: size}
							}
							return layout.Dimensions{}
						}),
						layout.Stacked(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8), Top: unit.Dp(8), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(h.matTheme, h.theme.Typo.Body, h.request.Options[i])
								lbl.Color = h.theme.Palette.Text
								return lbl.Layout(gtx)
							})
						}),
					)
				})
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return h.layoutCancelButton(gtx)
		}),
	)
}

func (h *ExtensionUIHandler) layoutConfirm(gtx layout.Context) layout.Dimensions {
	if h.confirmBtn.Clicked(gtx) {
		h.sendConfirm(true)
	}
	if h.cancelBtn.Clicked(gtx) {
		h.sendConfirm(false)
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(h.matTheme, h.theme.Typo.H3, h.request.Title)
			lbl.Color = h.theme.Palette.Text
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(h.matTheme, h.theme.Typo.Body, h.request.Message)
			lbl.Color = h.theme.Palette.TextSecondary
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(h.matTheme, &h.confirmBtn, "Yes")
					btn.Background = h.theme.Palette.Accent
					btn.Inset = layout.Inset{Left: unit.Dp(16), Right: unit.Dp(16), Top: unit.Dp(8), Bottom: unit.Dp(8)}
					return btn.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return h.layoutCancelButton(gtx)
				}),
			)
		}),
	)
}

func (h *ExtensionUIHandler) layoutInput(gtx layout.Context) layout.Dimensions {
	for {
		evt, ok := h.inputEditor.Update(gtx)
		if !ok {
			break
		}
		if _, isSubmit := evt.(widget.SubmitEvent); isSubmit {
			h.sendValue(h.inputEditor.Text())
		}
	}
	if h.submitBtn.Clicked(gtx) {
		h.sendValue(h.inputEditor.Text())
	}
	if h.cancelBtn.Clicked(gtx) {
		h.sendCancelled()
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(h.matTheme, h.theme.Typo.H3, h.request.Title)
			lbl.Color = h.theme.Palette.Text
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
		// Bordered input
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Stack{}.Layout(gtx,
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					rr := gtx.Dp(h.theme.Space.BorderRadius)
					size := gtx.Constraints.Min
					if size.X == 0 {
						size.X = gtx.Constraints.Max.X
					}
					if size.Y == 0 {
						size.Y = gtx.Constraints.Max.Y
					}
					widgets.DrawBorderedRect(gtx, h.theme.Palette.Surface, h.theme.Palette.Border, size, rr, 1)
					return layout.Dimensions{Size: size}
				}),
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: unit.Dp(12), Right: unit.Dp(12), Top: unit.Dp(10), Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						placeholder := h.request.Placeholder
						if placeholder == "" {
							placeholder = "Enter text..."
						}
						e := material.Editor(h.matTheme, &h.inputEditor, placeholder)
						e.Color = h.theme.Palette.Text
						e.HintColor = h.theme.Palette.TextTertiary
						return e.Layout(gtx)
					})
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(h.matTheme, &h.submitBtn, "Submit")
					btn.Background = h.theme.Palette.Accent
					btn.Inset = layout.Inset{Left: unit.Dp(16), Right: unit.Dp(16), Top: unit.Dp(8), Bottom: unit.Dp(8)}
					return btn.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return h.layoutCancelButton(gtx)
				}),
			)
		}),
	)
}

func (h *ExtensionUIHandler) layoutCancelButton(gtx layout.Context) layout.Dimensions {
	btn := material.Button(h.matTheme, &h.cancelBtn, "Cancel")
	btn.Background = h.theme.Palette.SurfaceAlt
	btn.Color = h.theme.Palette.TextSecondary
	btn.Inset = layout.Inset{Left: unit.Dp(16), Right: unit.Dp(16), Top: unit.Dp(8), Bottom: unit.Dp(8)}
	return btn.Layout(gtx)
}

func (h *ExtensionUIHandler) sendValue(value string) {
	if h.OnResponse != nil && h.request != nil {
		h.OnResponse(session.ExtensionUIResponse{
			Type: "extension_ui_response", ID: h.request.ID, Value: value,
		})
	}
	h.request = nil
}

func (h *ExtensionUIHandler) sendConfirm(confirmed bool) {
	if h.OnResponse != nil && h.request != nil {
		h.OnResponse(session.ExtensionUIResponse{
			Type: "extension_ui_response", ID: h.request.ID, Confirmed: &confirmed,
		})
	}
	h.request = nil
}

func (h *ExtensionUIHandler) sendCancelled() {
	if h.OnResponse != nil && h.request != nil {
		h.OnResponse(session.ExtensionUIResponse{
			Type: "extension_ui_response", ID: h.request.ID, Cancelled: true,
		})
	}
	h.request = nil
}
