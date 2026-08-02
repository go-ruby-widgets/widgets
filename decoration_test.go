// Copyright (c) 2026 the go-ruby-widgets/widgets authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package widgets

import (
	"testing"

	"github.com/go-widgets/painter"
	"github.com/go-widgets/toolkit"
)

// fullSpec is a representative decoration Hash exercising every field: a band
// with a centred caption + hairline, a rectangular close button with an ×
// glyph, a round traffic-light with an outline, a border, a shadow and a grip.
func fullSpec() map[string]any {
	return map[string]any{
		"title":        "Term",
		"title_ink":    "#f5f6fa",
		"title_color":  "#9b1c2e",
		"titlebar":     []any{0, 0, 40, 22},
		"title_center": true,
		"hairline":     "#bfbfbf",
		"border":       []any{0, 0, 40, 80},
		"border_color": "#ef4444",
		"shadow":       "#999999",
		"grip":         []any{26, 66, 14, 14},
		"show_grip":    true,
		"grip_color":   "#5b6072",
		"buttons": []any{
			map[string]any{
				"rect": []any{22, 4, 14, 14}, "shape": "rect",
				"face": "#e6e7ee", "glyph": "close", "glyph_ink": "#1a1a2e",
			},
			map[string]any{
				"rect": []any{4, 4, 12, 12}, "shape": "circle",
				"face": "#ff5f57", "outline": "#e0443e", "glyph": "none",
			},
			"not-a-hash", // skipped, like the Menu constructor
		},
	}
}

func TestDecorationConstruction(t *testing.T) {
	m := NewModule()
	id, err := m.Decoration(fullSpec())
	if err != nil {
		t.Fatalf("Decoration: %v", err)
	}
	d, ok := m.objs[id].(*toolkit.WindowDecoration)
	if !ok {
		t.Fatalf("handle %d is %T, want *toolkit.WindowDecoration", id, m.objs[id])
	}
	if d.Title != "Term" || !d.TitleCenter || !d.ShowGrip {
		t.Errorf("scalar fields wrong: title=%q center=%v grip=%v", d.Title, d.TitleCenter, d.ShowGrip)
	}
	if d.TitleColor != (painter.RGBA{R: 0x9b, G: 0x1c, B: 0x2e, A: 0xFF}) {
		t.Errorf("TitleColor = %v", d.TitleColor)
	}
	if d.Hairline.A == 0 || d.BorderColor.A == 0 || d.Shadow.A == 0 {
		t.Error("hairline/border/shadow colours should be opaque")
	}
	if d.Titlebar != (toolkit.Rect{X: 0, Y: 0, W: 40, H: 22}) {
		t.Errorf("Titlebar = %v", d.Titlebar)
	}
	if d.Grip != (toolkit.Rect{X: 26, Y: 66, W: 14, H: 14}) {
		t.Errorf("Grip = %v", d.Grip)
	}
	if len(d.Buttons) != 2 { // the "not-a-hash" element is skipped
		t.Fatalf("len(Buttons) = %d, want 2", len(d.Buttons))
	}
	if d.Buttons[0].Shape != toolkit.DecoButtonRect || d.Buttons[0].Glyph != toolkit.DecoGlyphClose {
		t.Errorf("button 0 = %+v", d.Buttons[0])
	}
	if d.Buttons[1].Shape != toolkit.DecoButtonCircle || d.Buttons[1].Outline.A == 0 {
		t.Errorf("button 1 = %+v", d.Buttons[1])
	}

	// It renders through the normal render path.
	img, err := m.Render(id, 40, 80)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if got := len(img["pixels"].([]byte)); got != 40*80*4 {
		t.Errorf("pixels len = %d, want %d", got, 40*80*4)
	}
}

func TestDecorationMinimalAndDefaults(t *testing.T) {
	// An empty spec builds a blank decoration (all zero) — no error.
	m := NewModule()
	id, err := m.Decoration(map[string]any{})
	if err != nil {
		t.Fatalf("empty Decoration: %v", err)
	}
	d := m.objs[id].(*toolkit.WindowDecoration)
	if d.Title != "" || d.TitleCenter || d.ShowGrip || len(d.Buttons) != 0 {
		t.Errorf("empty spec did not yield a blank decoration: %+v", d)
	}

	// A button hash omitting shape/glyph defaults to rect/none.
	id2, err := m.Decoration(map[string]any{
		"buttons": []any{map[string]any{"rect": []any{0, 0, 8, 8}, "face": "#ffffff"}},
	})
	if err != nil {
		t.Fatalf("default-button Decoration: %v", err)
	}
	b := m.objs[id2].(*toolkit.WindowDecoration).Buttons[0]
	if b.Shape != toolkit.DecoButtonRect || b.Glyph != toolkit.DecoGlyphNone {
		t.Errorf("defaults wrong: shape=%v glyph=%v", b.Shape, b.Glyph)
	}
}

func TestDecorationGlyphNames(t *testing.T) {
	m := NewModule()
	for name, want := range map[string]toolkit.DecoGlyph{
		"none":     toolkit.DecoGlyphNone,
		"close":    toolkit.DecoGlyphClose,
		"minimize": toolkit.DecoGlyphMinimize,
		"maximize": toolkit.DecoGlyphMaximize,
	} {
		id, err := m.Decoration(map[string]any{
			"buttons": []any{map[string]any{"glyph": name}},
		})
		if err != nil {
			t.Fatalf("glyph %q: %v", name, err)
		}
		if got := m.objs[id].(*toolkit.WindowDecoration).Buttons[0].Glyph; got != want {
			t.Errorf("glyph %q = %v, want %v", name, got, want)
		}
	}
	// The circle shape name maps too.
	id, err := m.Decoration(map[string]any{"buttons": []any{map[string]any{"shape": "circle"}}})
	if err != nil {
		t.Fatalf("circle shape: %v", err)
	}
	if got := m.objs[id].(*toolkit.WindowDecoration).Buttons[0].Shape; got != toolkit.DecoButtonCircle {
		t.Errorf("circle shape = %v", got)
	}
}

func TestDecorationErrors(t *testing.T) {
	m := NewModule()
	cases := []struct {
		name string
		spec map[string]any
	}{
		{"bad title_ink", map[string]any{"title_ink": "nope"}},
		{"bad title_color", map[string]any{"title_color": "zzzzzz"}},
		{"bad hairline", map[string]any{"hairline": "#12"}},
		{"bad border_color", map[string]any{"border_color": "xx"}},
		{"bad shadow", map[string]any{"shadow": "yy"}},
		{"bad grip_color", map[string]any{"grip_color": "qq"}},
		{"bad titlebar type", map[string]any{"titlebar": "nope"}},
		{"short border", map[string]any{"border": []any{1, 2, 3}}},
		{"non-number grip", map[string]any{"grip": []any{1, 2, "x", 4}}},
		{"buttons not array", map[string]any{"buttons": "nope"}},
		{"bad button rect", map[string]any{"buttons": []any{map[string]any{"rect": []any{1}}}}},
		{"bad button shape", map[string]any{"buttons": []any{map[string]any{"shape": "hexagon"}}}},
		{"bad button glyph", map[string]any{"buttons": []any{map[string]any{"glyph": "spiral"}}}},
		{"bad button face", map[string]any{"buttons": []any{map[string]any{"face": "zz"}}}},
		{"bad button outline", map[string]any{"buttons": []any{map[string]any{"outline": "zz"}}}},
		{"bad button glyph_ink", map[string]any{"buttons": []any{map[string]any{"glyph_ink": "zz"}}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := m.Decoration(c.spec); err == nil {
				t.Errorf("%s: expected error, got nil", c.name)
			}
		})
	}
}

func TestDecorationViaCall(t *testing.T) {
	m := NewModule()
	if !contains(Methods(m), "decoration") {
		t.Fatal("Methods should list decoration")
	}
	res, err := Call(m, "decoration", fullSpec())
	if err != nil {
		t.Fatalf("Call decoration: %v", err)
	}
	if _, ok := res.(int); !ok {
		t.Fatalf("decoration returned %T, want int handle", res)
	}
	// A malformed nested colour surfaces as a Call error.
	if _, err := Call(m, "decoration", map[string]any{"title_color": "zzzzzz"}); err == nil {
		t.Error("Call decoration with bad colour should error")
	}
}

func TestParseButtonsNil(t *testing.T) {
	// A nil "buttons" value yields no buttons and no error.
	bs, err := parseButtons(nil)
	if err != nil || bs != nil {
		t.Errorf("parseButtons(nil) = %v, %v; want nil, nil", bs, err)
	}
}

func TestSpecStrNilValue(t *testing.T) {
	// A present-but-nil value reads as the empty string.
	if got := specStr(map[string]any{"k": nil}, "k"); got != "" {
		t.Errorf("specStr(nil value) = %q, want empty", got)
	}
}
