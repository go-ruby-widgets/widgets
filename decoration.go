// Copyright (c) 2026 the go-ruby-widgets/widgets authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package widgets

import (
	"fmt"

	"github.com/go-widgets/painter"
	"github.com/go-widgets/toolkit"
)

// Decoration constructs a toolkit.WindowDecoration — a window's frame chrome
// (title-bar band + caption + buttons + border + optional shadow + resize grip)
// painted with EXPLICIT colours and EXPLICIT frame-local geometry — from a Ruby
// spec Hash, and returns its handle.
//
// The host (a compositor) owns the window model and its hit-testing, so it
// passes the exact rects it hit-tests against plus the palette its style needs;
// the widget only paints. The body region is left transparent so the host
// composites the decoration over a live window body.
//
// Recognised spec keys (all optional; absent = zero/omitted):
//
//	"title"        String    the caption
//	"title_ink"    hex       caption colour
//	"title_color"  hex       title-bar band fill
//	"titlebar"     [x,y,w,h] band rect (frame-local)
//	"title_center" Bool      centre the caption (default: left)
//	"hairline"     hex       band bottom hairline ("" = none)
//	"border"       [x,y,w,h] full frame extent (frame-local)
//	"border_color" hex       border stroke ("" = none)
//	"shadow"       hex       faux drop shadow past the border ("" = none)
//	"grip"         [x,y,w,h] resize-grip rect (frame-local)
//	"show_grip"    Bool      draw the grip
//	"grip_color"   hex       grip diagonals colour
//	"buttons"      [Hash]    the title-bar button cluster, each:
//	                 "rect"      [x,y,w,h] (frame-local)
//	                 "shape"     "rect" | "circle"     (default "rect")
//	                 "face"      hex   face fill
//	                 "outline"   hex   circle outline ("" = none)
//	                 "glyph"     "none"|"close"|"minimize"|"maximize"
//	                 "glyph_ink" hex   glyph colour
//
// A malformed colour, rect or enum value is an error, surfaced by Call.
func (m *Module) Decoration(spec map[string]any) (int, error) {
	d := &toolkit.WindowDecoration{}
	d.Title = specStr(spec, "title")
	d.TitleCenter = truthy(spec["title_center"])
	d.ShowGrip = truthy(spec["show_grip"])

	var err error
	if d.TitleInk, err = specColor(spec, "title_ink"); err != nil {
		return 0, err
	}
	if d.TitleColor, err = specColor(spec, "title_color"); err != nil {
		return 0, err
	}
	if d.Hairline, err = specColor(spec, "hairline"); err != nil {
		return 0, err
	}
	if d.BorderColor, err = specColor(spec, "border_color"); err != nil {
		return 0, err
	}
	if d.Shadow, err = specColor(spec, "shadow"); err != nil {
		return 0, err
	}
	if d.GripColor, err = specColor(spec, "grip_color"); err != nil {
		return 0, err
	}
	if d.Titlebar, err = specRect(spec, "titlebar"); err != nil {
		return 0, err
	}
	if d.Border, err = specRect(spec, "border"); err != nil {
		return 0, err
	}
	if d.Grip, err = specRect(spec, "grip"); err != nil {
		return 0, err
	}
	if d.Buttons, err = parseButtons(spec["buttons"]); err != nil {
		return 0, err
	}
	return m.reg(d), nil
}

// parseButtons turns the "buttons" value (a Ruby Array of Hashes) into the
// toolkit button slice. A nil value yields no buttons; a non-Array is an error;
// non-Hash elements are skipped (matching the Menu constructor's tolerance).
func parseButtons(v any) ([]toolkit.DecoButton, error) {
	if v == nil {
		return nil, nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("widgets: decoration: \"buttons\" must be an Array, got %T", v)
	}
	out := make([]toolkit.DecoButton, 0, len(arr))
	for i, raw := range arr {
		h, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		b := toolkit.DecoButton{}
		rect, err := specRect(h, "rect")
		if err != nil {
			return nil, fmt.Errorf("widgets: decoration: button %d: %w", i, err)
		}
		b.Rect = rect
		if b.Shape, err = parseShape(specStr(h, "shape")); err != nil {
			return nil, fmt.Errorf("widgets: decoration: button %d: %w", i, err)
		}
		if b.Glyph, err = parseGlyph(specStr(h, "glyph")); err != nil {
			return nil, fmt.Errorf("widgets: decoration: button %d: %w", i, err)
		}
		if b.Face, err = specColor(h, "face"); err != nil {
			return nil, fmt.Errorf("widgets: decoration: button %d: %w", i, err)
		}
		if b.Outline, err = specColor(h, "outline"); err != nil {
			return nil, fmt.Errorf("widgets: decoration: button %d: %w", i, err)
		}
		if b.GlyphInk, err = specColor(h, "glyph_ink"); err != nil {
			return nil, fmt.Errorf("widgets: decoration: button %d: %w", i, err)
		}
		out = append(out, b)
	}
	return out, nil
}

// parseShape maps a shape name to a toolkit.DecoButtonShape ("" = "rect").
func parseShape(name string) (toolkit.DecoButtonShape, error) {
	switch name {
	case "", "rect":
		return toolkit.DecoButtonRect, nil
	case "circle":
		return toolkit.DecoButtonCircle, nil
	default:
		return 0, fmt.Errorf("unknown button shape %q", name)
	}
}

// parseGlyph maps a glyph name to a toolkit.DecoGlyph ("" = "none").
func parseGlyph(name string) (toolkit.DecoGlyph, error) {
	switch name {
	case "", "none":
		return toolkit.DecoGlyphNone, nil
	case "close":
		return toolkit.DecoGlyphClose, nil
	case "minimize":
		return toolkit.DecoGlyphMinimize, nil
	case "maximize":
		return toolkit.DecoGlyphMaximize, nil
	default:
		return 0, fmt.Errorf("unknown button glyph %q", name)
	}
}

// specStr reads a string-valued key ("" when absent).
func specStr(spec map[string]any, key string) string {
	if v, ok := spec[key]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

// specColor reads a hex-colour key. An absent or empty value yields the zero
// RGBA (the toolkit's "absent/theme" sentinel).
func specColor(spec map[string]any, key string) (painter.RGBA, error) {
	return parseHexColor(specStr(spec, key))
}

// specRect reads a [x,y,w,h] rect key. An absent value yields the zero Rect
// (which the widget treats as "skip this element"). A present value must be an
// Array of at least four numbers.
func specRect(spec map[string]any, key string) (toolkit.Rect, error) {
	v, ok := spec[key]
	if !ok || v == nil {
		return toolkit.Rect{}, nil
	}
	arr, ok := v.([]any)
	if !ok || len(arr) < 4 {
		return toolkit.Rect{}, fmt.Errorf("widgets: decoration: %q must be a [x,y,w,h] Array, got %v", key, v)
	}
	nums := make([]int, 4)
	for i := 0; i < 4; i++ {
		n, ok := toInt(arr[i])
		if !ok {
			return toolkit.Rect{}, fmt.Errorf("widgets: decoration: %q element %d is not a number: %v", key, i, arr[i])
		}
		nums[i] = n
	}
	return toolkit.Rect{X: nums[0], Y: nums[1], W: nums[2], H: nums[3]}, nil
}
