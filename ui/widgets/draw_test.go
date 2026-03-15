package widgets

import (
	"image/color"
	"testing"
)

func TestWithAlpha(t *testing.T) {
	c := color.NRGBA{R: 255, G: 128, B: 0, A: 255}
	got := WithAlpha(c, 128)
	if got.R != 255 || got.G != 128 || got.B != 0 || got.A != 128 {
		t.Errorf("WithAlpha: got %v", got)
	}
}

func TestMulAlpha(t *testing.T) {
	c := color.NRGBA{R: 255, G: 128, B: 0, A: 200}
	got := MulAlpha(c, 128)
	// 200 * 128 / 255 ≈ 100
	if got.A < 99 || got.A > 101 {
		t.Errorf("MulAlpha: expected ~100, got %d", got.A)
	}
	if got.R != 255 || got.G != 128 || got.B != 0 {
		t.Errorf("MulAlpha: RGB changed: %v", got)
	}
}

func TestMulAlphaZero(t *testing.T) {
	c := color.NRGBA{R: 255, G: 255, B: 255, A: 0}
	got := MulAlpha(c, 255)
	if got.A != 0 {
		t.Errorf("MulAlpha with zero alpha: expected 0, got %d", got.A)
	}
}

func TestHoveredDark(t *testing.T) {
	// Dark color should blend toward white
	dark := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	hov := Hovered(dark)
	if hov.R <= dark.R || hov.G <= dark.G || hov.B <= dark.B {
		t.Errorf("Hovered dark: expected lighter, got %v from %v", hov, dark)
	}
}

func TestHoveredLight(t *testing.T) {
	// Light color should blend toward black
	light := color.NRGBA{R: 220, G: 220, B: 220, A: 255}
	hov := Hovered(light)
	if hov.R >= light.R || hov.G >= light.G || hov.B >= light.B {
		t.Errorf("Hovered light: expected darker, got %v from %v", hov, light)
	}
}

func TestHoveredTransparent(t *testing.T) {
	got := Hovered(color.NRGBA{})
	if got.A == 0 {
		t.Error("Hovered transparent: expected non-zero alpha")
	}
}

func TestDisabled(t *testing.T) {
	c := color.NRGBA{R: 100, G: 200, B: 50, A: 255}
	d := Disabled(c)
	// Should have reduced alpha
	if d.A >= c.A {
		t.Errorf("Disabled: expected reduced alpha, got %d >= %d", d.A, c.A)
	}
}

func TestApproxLuminance(t *testing.T) {
	black := approxLuminance(color.NRGBA{R: 0, G: 0, B: 0})
	white := approxLuminance(color.NRGBA{R: 255, G: 255, B: 255})
	if black != 0 {
		t.Errorf("black luminance: expected 0, got %d", black)
	}
	if white < 250 {
		t.Errorf("white luminance: expected ~255, got %d", white)
	}
}
