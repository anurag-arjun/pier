package widgets

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
)

// HoverState tracks mouse enter/leave for a layout region.
//
// Usage:
//
//	var hover widgets.HoverState
//
//	// In layout:
//	hover.Update(gtx)
//	if hover.Hovered() {
//	    // draw hover background
//	}
//
// HoverState uses pointer.PassOp so it does not consume events —
// clickable widgets underneath still work.
type HoverState struct {
	hovering bool
	tag      bool // used as event tag identity
}

// Update registers the hover tracking area and drains pointer events.
// Must be called within a clip region that defines the hover area.
// Typically called after pushing a clip.Rect or clip.RRect.
func (h *HoverState) Update(gtx layout.Context) {
	// Register pointer event tracking
	defer pointer.PassOp{}.Push(gtx.Ops).Pop()
	defer clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, &h.tag)

	for {
		ev, ok := gtx.Event(pointer.Filter{
			Target: &h.tag,
			Kinds:  pointer.Enter | pointer.Leave | pointer.Cancel,
		})
		if !ok {
			break
		}
		if pe, ok := ev.(pointer.Event); ok {
			switch pe.Kind {
			case pointer.Enter:
				h.hovering = true
			case pointer.Leave, pointer.Cancel:
				h.hovering = false
			}
		}
	}
}

// Hovered returns whether the mouse is currently over the tracked area.
func (h *HoverState) Hovered() bool {
	return h.hovering
}
