// Copyright (c) 2026 the go-ruby-widgets/widgets authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package widgets

import (
	"fmt"
	"strings"

	"github.com/go-widgets/toolkit"
)

// This file binds the desktop-environment widgets a compositor (wasmdesk) needs
// beyond the overlay/chrome set in overlays.go: the StatusArea / StatusIcon tray
// (menu-bar extras + their badges), the full-bounds Wallpaper backdrop (image or
// gradient), and the Thumbnail window-preview tile (Exposé, Alt-Tab, dock peek).
// Each constructor returns an opaque integer handle exactly like the base set,
// and the DE-specific state (tray membership, selection, hover) is addressed by
// that handle through the shared add / set-state seams.

// --- tray (StatusArea / StatusIcon) ------------------------------------------

// StatusArea constructs an empty tray: a left-to-right row of StatusIcons (the
// menu-bar extras / notification-area slots), each in a square cell. Populate it
// by adding StatusIcon handles with AddWidget; the row re-flows on every add and
// on SetBounds.
func (m *Module) StatusArea() int { return m.reg(toolkit.NewStatusArea()) }

// StatusIcon constructs a tray indicator painting a stock vector glyph named by
// icon: "new", "open", "save", "cut", "copy", "paste", "undo", "redo", "search"
// or "settings" (an empty string draws no glyph — a badge-only slot). tooltip is
// the hover text the host surfaces. onClick fires on a primary click and
// onRightClick on a secondary (menu) click; each is wired only when non-empty.
// An unknown icon name is an error, reported by Call.
func (m *Module) StatusIcon(icon, tooltip, onClick, onRightClick string) (int, error) {
	fn, err := parseStockIcon(icon)
	if err != nil {
		return 0, err
	}
	s := toolkit.NewStatusIcon(fn)
	s.Tooltip = tooltip
	if onClick != "" {
		s.OnClick = func() { m.fire(onClick) }
	}
	if onRightClick != "" {
		s.OnRightClick = func() { m.fire(onRightClick) }
	}
	return m.reg(s), nil
}

// StatusIconImage is StatusIcon with a caller-supplied RGBA image instead of a
// stock glyph: pixels is raw RGBA bytes (a Ruby binary String) or a base64
// String and must hold at least w*h*4 bytes; w and h are the source dimensions
// and must be positive. The image is drawn aspect-preserved and centred in the
// tray cell. tooltip, onClick and onRightClick behave as in StatusIcon. A bad
// base64 string, non-positive size or short buffer is an error.
func (m *Module) StatusIconImage(pixels any, w, h int, tooltip, onClick, onRightClick string) (int, error) {
	b, err := toBytes(pixels)
	if err != nil {
		return 0, err
	}
	if w <= 0 || h <= 0 {
		return 0, fmt.Errorf("widgets: StatusIconImage: size must be positive, got %dx%d", w, h)
	}
	if len(b) < w*h*4 {
		return 0, fmt.Errorf("widgets: StatusIconImage: got %d pixel bytes, need %d (w*h*4)", len(b), w*h*4)
	}
	s := toolkit.NewStatusIconImage(b, w, h)
	s.Tooltip = tooltip
	if onClick != "" {
		s.OnClick = func() { m.fire(onClick) }
	}
	if onRightClick != "" {
		s.OnRightClick = func() { m.fire(onRightClick) }
	}
	return m.reg(s), nil
}

// --- wallpaper ---------------------------------------------------------------

// Wallpaper constructs a full-bounds desktop backdrop that paints a caller-
// supplied RGBA image scaled by mode: "fill" (the default, cover — aspect-
// preserved, cropped to fill the screen), "fit" (contain — the whole image
// centred inside the bounds), "center" (1:1, centred) or "tile" (repeated 1:1).
// pixels is raw RGBA bytes or a base64 String and must hold at least w*h*4
// bytes; w and h are the source dimensions and must be positive. The wallpaper
// is event-transparent (clicks pass through to the composited scene). A bad
// base64 string, non-positive size, short buffer or unknown mode is an error.
func (m *Module) Wallpaper(pixels any, w, h int, mode string) (int, error) {
	b, err := toBytes(pixels)
	if err != nil {
		return 0, err
	}
	if w <= 0 || h <= 0 {
		return 0, fmt.Errorf("widgets: Wallpaper: size must be positive, got %dx%d", w, h)
	}
	if len(b) < w*h*4 {
		return 0, fmt.Errorf("widgets: Wallpaper: got %d pixel bytes, need %d (w*h*4)", len(b), w*h*4)
	}
	md, err := parseWallpaperMode(mode)
	if err != nil {
		return 0, err
	}
	return m.reg(toolkit.NewWallpaper(b, w, h, md)), nil
}

// WallpaperGradient constructs an image-less Wallpaper painting a vertical
// gradient from topHex down to bottomHex — both "#rrggbb"/"#rrggbbaa" hex. An
// empty top selects the theme Background; an empty (zero-alpha) bottom makes the
// fill a solid top colour (no gradient). Like Wallpaper it is event-transparent.
// A malformed colour is an error.
func (m *Module) WallpaperGradient(topHex, bottomHex string) (int, error) {
	top, err := parseHexColor(topHex)
	if err != nil {
		return 0, err
	}
	bottom, err := parseHexColor(bottomHex)
	if err != nil {
		return 0, err
	}
	return m.reg(toolkit.NewWallpaperGradient(top, bottom)), nil
}

// --- thumbnail ---------------------------------------------------------------

// Thumbnail constructs a window-preview tile that renders a caller-supplied RGBA
// buffer scaled down (aspect-preserved, centred) into its bounds, with an
// optional caption strip carrying label and a selected/hover border. It is the
// tile an Exposé grid, an Alt-Tab switcher or a dock-hover peek is built from.
// pixels is raw RGBA bytes or a base64 String and must hold at least w*h*4
// bytes; w and h are the source dimensions and must be positive. onClick fires
// (reported by Dispatch) on a click, so a grid can select the tile; it is wired
// only when non-empty. A bad base64 string, non-positive size or short buffer is
// an error.
func (m *Module) Thumbnail(pixels any, w, h int, label, onClick string) (int, error) {
	b, err := toBytes(pixels)
	if err != nil {
		return 0, err
	}
	if w <= 0 || h <= 0 {
		return 0, fmt.Errorf("widgets: Thumbnail: size must be positive, got %dx%d", w, h)
	}
	if len(b) < w*h*4 {
		return 0, fmt.Errorf("widgets: Thumbnail: got %d pixel bytes, need %d (w*h*4)", len(b), w*h*4)
	}
	t := toolkit.NewThumbnail(b, w, h)
	t.Label = label
	if onClick != "" {
		t.OnClick = func() { m.fire(onClick) }
	}
	return m.reg(t), nil
}

// SetSelected sets a widget's selection state, dispatching on the handle's type:
// a Thumbnail's selected-border flag (v is truthy — the Alt-Tab / Exposé current
// choice), a CommandPalette's selection index (v is an integer, clamped into the
// filtered list) or a Calendar's selected day (v is an integer day, re-clamped
// into the current month). A handle that is none of these is an error, as is a
// non-integer v where an index/day is expected.
func (m *Module) SetSelected(id int, v any) error {
	o, err := m.get(id)
	if err != nil {
		return err
	}
	switch w := o.(type) {
	case *toolkit.Thumbnail:
		w.Selected = truthy(v)
	case *toolkit.CommandPalette:
		n, ok := toInt(v)
		if !ok {
			return fmt.Errorf("widgets: SetSelected: index must be an integer, got %T", v)
		}
		w.SetSelected(n)
	case *toolkit.Calendar:
		n, ok := toInt(v)
		if !ok {
			return fmt.Errorf("widgets: SetSelected: day must be an integer, got %T", v)
		}
		w.SetDate(w.Year, w.Month, n)
	default:
		return fmt.Errorf("widgets: SetSelected: handle %d (%T) has no selection state", id, o)
	}
	return nil
}

// SetHover sets the hover-border state of a Thumbnail (the pointer is over the
// tile). A handle that is not a Thumbnail is an error.
func (m *Module) SetHover(id int, v bool) error {
	o, err := m.get(id)
	if err != nil {
		return err
	}
	t, ok := o.(*toolkit.Thumbnail)
	if !ok {
		return fmt.Errorf("widgets: SetHover: handle %d (%T) is not a thumbnail", id, o)
	}
	t.Hover = v
	return nil
}

// --- small parsers -----------------------------------------------------------

// parseStockIcon maps a stock-icon name to the matching toolkit DrawIcon vector
// painter. An empty name yields a nil IconFunc (a badge-only slot).
func parseStockIcon(name string) (toolkit.IconFunc, error) {
	switch strings.ToLower(name) {
	case "":
		return nil, nil
	case "new":
		return toolkit.DrawIconNew, nil
	case "open":
		return toolkit.DrawIconOpen, nil
	case "save":
		return toolkit.DrawIconSave, nil
	case "cut":
		return toolkit.DrawIconCut, nil
	case "copy":
		return toolkit.DrawIconCopy, nil
	case "paste":
		return toolkit.DrawIconPaste, nil
	case "undo":
		return toolkit.DrawIconUndo, nil
	case "redo":
		return toolkit.DrawIconRedo, nil
	case "search":
		return toolkit.DrawIconSearch, nil
	case "settings":
		return toolkit.DrawIconSettings, nil
	default:
		return nil, fmt.Errorf("widgets: unknown status icon %q", name)
	}
}

// parseWallpaperMode maps a mode name to a toolkit.WallpaperMode ("" == fill).
func parseWallpaperMode(name string) (toolkit.WallpaperMode, error) {
	switch strings.ToLower(name) {
	case "", "fill":
		return toolkit.WallpaperFill, nil
	case "fit":
		return toolkit.WallpaperFit, nil
	case "center", "centre":
		return toolkit.WallpaperCenter, nil
	case "tile":
		return toolkit.WallpaperTile, nil
	default:
		return 0, fmt.Errorf("widgets: unknown wallpaper mode %q", name)
	}
}
