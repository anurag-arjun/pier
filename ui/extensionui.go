package ui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	apptheme "github.com/lighto/pier/app"
	"github.com/lighto/pier/session"
)

// ExtensionUIHandler renders modal dialogs for extension_ui_request events.
type ExtensionUIHandler struct {
	theme    apptheme.Theme
	matTheme *material.Theme

	// Active dialog
	request    *session.ExtensionUIRequest
	selectBtns []widget.Clickable
	confirmBtn widget.Clickable
	cancelBtn  widget.Clickable
	inputEditor widget.Editor
	submitBtn  widget.Clickable

	// Callback to send response back to pi
	OnResponse func(resp session.ExtensionUIResponse)
}

// NewExtensionUIHandler creates an extension UI handler.
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

// SetRequest sets the active dialog request.
func (h *ExtensionUIHandler) SetRequest(req *session.ExtensionUIRequest) {
	h.request = req
	h.selectBtns = nil
	if req != nil && req.Method == "select" {
		h.selectBtns = make([]widget.Clickable, len(req.Options))
	}
	if req != nil && (req.Method == "input" || req.Method == "editor") {
		h.inputEditor.SetText(req.Prefill)
	}
}

// HasActiveDialog returns whether a dialog is showing.
func (h *ExtensionUIHandler) HasActiveDialog() bool {
	return h.request != nil
}

// Layout renders the active dialog if any. Returns true if a response was sent.
func (h *ExtensionUIHandler) Layout(gtx layout.Context) (layout.Dimensions, bool) {
	if h.request == nil {
		return layout.Dimensions{}, false
	}

	switch h.request.Method {
	case "select":
		return h.layoutSelect(gtx)
	case "confirm":
		return h.layoutConfirm(gtx)
	case "input", "editor":
		return h.layoutInput(gtx)
	case "notify":
		// Fire-and-forget: just show briefly, no response needed
		h.request = nil
		return layout.Dimensions{}, false
	default:
		// Unknown method — cancel
		h.sendCancelled()
		return layout.Dimensions{}, true
	}
}

func (h *ExtensionUIHandler) layoutSelect(gtx layout.Context) (layout.Dimensions, bool) {
	responded := false

	for i := range h.request.Options {
		if i < len(h.selectBtns) && h.selectBtns[i].Clicked(gtx) {
			h.sendValue(h.request.Options[i])
			responded = true
		}
	}
	if h.cancelBtn.Clicked(gtx) {
		h.sendCancelled()
		responded = true
	}

	dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H4(h.matTheme, h.request.Title)
			lbl.Color = h.theme.Palette.Text
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			list := layout.List{Axis: layout.Vertical}
			return list.Layout(gtx, len(h.request.Options), func(gtx layout.Context, i int) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(h.matTheme, &h.selectBtns[i], h.request.Options[i])
					btn.Background = h.theme.Palette.SurfaceAlt
					btn.Color = h.theme.Palette.Text
					return btn.Layout(gtx)
				})
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(h.matTheme, &h.cancelBtn, "Cancel")
			btn.Background = h.theme.Palette.Border
			btn.Color = h.theme.Palette.TextSecondary
			return btn.Layout(gtx)
		}),
	)
	return dims, responded
}

func (h *ExtensionUIHandler) layoutConfirm(gtx layout.Context) (layout.Dimensions, bool) {
	responded := false

	if h.confirmBtn.Clicked(gtx) {
		h.sendConfirm(true)
		responded = true
	}
	if h.cancelBtn.Clicked(gtx) {
		h.sendConfirm(false)
		responded = true
	}

	dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H4(h.matTheme, h.request.Title)
			lbl.Color = h.theme.Palette.Text
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body1(h.matTheme, h.request.Message)
			lbl.Color = h.theme.Palette.TextSecondary
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(h.matTheme, &h.confirmBtn, "Yes")
					btn.Background = h.theme.Palette.Accent
					return btn.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(h.matTheme, &h.cancelBtn, "No")
					btn.Background = h.theme.Palette.Border
					btn.Color = h.theme.Palette.TextSecondary
					return btn.Layout(gtx)
				}),
			)
		}),
	)
	return dims, responded
}

func (h *ExtensionUIHandler) layoutInput(gtx layout.Context) (layout.Dimensions, bool) {
	responded := false

	for {
		evt, ok := h.inputEditor.Update(gtx)
		if !ok {
			break
		}
		if _, isSubmit := evt.(widget.SubmitEvent); isSubmit {
			h.sendValue(h.inputEditor.Text())
			responded = true
		}
	}
	if h.submitBtn.Clicked(gtx) {
		h.sendValue(h.inputEditor.Text())
		responded = true
	}
	if h.cancelBtn.Clicked(gtx) {
		h.sendCancelled()
		responded = true
	}

	dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H4(h.matTheme, h.request.Title)
			lbl.Color = h.theme.Palette.Text
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			placeholder := h.request.Placeholder
			if placeholder == "" {
				placeholder = "Enter text..."
			}
			e := material.Editor(h.matTheme, &h.inputEditor, placeholder)
			e.Color = h.theme.Palette.Text
			return e.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(h.matTheme, &h.submitBtn, "Submit")
					btn.Background = h.theme.Palette.Accent
					return btn.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(h.matTheme, &h.cancelBtn, "Cancel")
					btn.Background = h.theme.Palette.Border
					btn.Color = h.theme.Palette.TextSecondary
					return btn.Layout(gtx)
				}),
			)
		}),
	)
	return dims, responded
}

func (h *ExtensionUIHandler) sendValue(value string) {
	if h.OnResponse != nil && h.request != nil {
		h.OnResponse(session.ExtensionUIResponse{
			Type:  "extension_ui_response",
			ID:    h.request.ID,
			Value: value,
		})
	}
	h.request = nil
}

func (h *ExtensionUIHandler) sendConfirm(confirmed bool) {
	if h.OnResponse != nil && h.request != nil {
		h.OnResponse(session.ExtensionUIResponse{
			Type:      "extension_ui_response",
			ID:        h.request.ID,
			Confirmed: &confirmed,
		})
	}
	h.request = nil
}

func (h *ExtensionUIHandler) sendCancelled() {
	if h.OnResponse != nil && h.request != nil {
		h.OnResponse(session.ExtensionUIResponse{
			Type:      "extension_ui_response",
			ID:        h.request.ID,
			Cancelled: true,
		})
	}
	h.request = nil
}
