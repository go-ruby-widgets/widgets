// Copyright (c) 2026 the go-ruby-widgets/widgets authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package widgets

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/go-widgets/toolkit"
)

// This file binds the overlay + chrome widgets a compositor (wasmdesk) and its
// applets need on top of the base leaf/container set in widgets.go: the transient
// Notification / Toast banners, the Badge count, the Image blitter (wallpaper +
// icons), the ContextMenu / Popover / CommandPalette pop-ups, and the compact
// IconButton / Tooltip / Avatar / LevelBar affordances. Each constructor returns
// an opaque integer handle exactly like the base set, and the overlay-specific
// state (visibility, anchoring, life, kind, value) is addressed by that handle.

// --- overlay + chrome constructors -------------------------------------------

// Notification constructs a transient, auto-dismissing banner carrying text. It
// starts hidden with the default life budget pre-armed; a host makes it visible
// (SetVisible / AnchorIn), positions it, then Ticks it down each animation frame.
func (m *Module) Notification(text string) int {
	return m.reg(toolkit.NewNotification(text))
}

// Toast constructs a short-lived severity pill: text rendered in a Kind-coloured
// body ("info"/"success"/"warning"/"error"; "" == info). An unknown kind is an
// error, reported by Call. When actionLabel is non-empty a small action button is
// rendered inside the pill's right edge and, when action is also non-empty, its
// callback identifier fires (reported by Dispatch) when that button is clicked.
func (m *Module) Toast(text, kind, actionLabel, action string) (int, error) {
	k, err := parseToastKind(kind)
	if err != nil {
		return 0, err
	}
	t := toolkit.NewToast(text, k)
	if actionLabel != "" {
		t.ActionLabel = actionLabel
		if action != "" {
			t.Action = func() { m.fire(action) }
		}
	}
	return m.reg(t), nil
}

// Badge constructs a small pill-shaped counter/indicator carrying text (the "12"
// on an inbox icon). fill overrides the pill body colour and ink the text colour;
// both are "#rrggbb"/"#rrggbbaa" hex, and an empty string selects the theme's
// Accent (fill) / Background (ink) at render time. A malformed colour is an error.
func (m *Module) Badge(text, fill, ink string) (int, error) {
	f, err := parseHexColor(fill)
	if err != nil {
		return 0, err
	}
	i, err := parseHexColor(ink)
	if err != nil {
		return 0, err
	}
	b := toolkit.NewBadge(text)
	b.Fill, b.Ink = f, i
	return m.reg(b), nil
}

// Image constructs a widget that blits a caller-supplied RGBA pixel buffer (the
// desktop wallpaper, an app icon). pixels is either raw RGBA bytes (a Ruby
// binary String surfaces as []byte) or a base64-encoded String; it must hold at
// least w*h*4 bytes. w and h are the source dimensions and must be positive.
// scale selects how the source maps onto the widget bounds: "stretch" (the
// default, fill ignoring aspect) or "fit" (aspect-preserving, centred). A bad
// base64 string, non-positive size, short buffer or unknown scale is an error.
func (m *Module) Image(pixels any, w, h int, scale string) (int, error) {
	b, err := toBytes(pixels)
	if err != nil {
		return 0, err
	}
	if w <= 0 || h <= 0 {
		return 0, fmt.Errorf("widgets: Image: size must be positive, got %dx%d", w, h)
	}
	if len(b) < w*h*4 {
		return 0, fmt.Errorf("widgets: Image: got %d pixel bytes, need %d (w*h*4)", len(b), w*h*4)
	}
	sm, err := parseScaleMode(scale)
	if err != nil {
		return 0, err
	}
	img := toolkit.NewImage(b, w, h)
	img.Scale = sm
	return m.reg(img), nil
}

// ContextMenu wraps an existing Menu handle (built with Menu) as a right-click
// pop-up: it appears at a point (Popup), clamps itself inside the surface, and
// dismisses on an outside click. The menu handle must reference a Menu.
func (m *Module) ContextMenu(menu int) (int, error) {
	o, err := m.get(menu)
	if err != nil {
		return 0, err
	}
	mn, ok := o.(*toolkit.Menu)
	if !ok {
		return 0, fmt.Errorf("widgets: ContextMenu: handle %d (%T) is not a menu", menu, o)
	}
	return m.reg(toolkit.NewContextMenu(mn)), nil
}

// Popover constructs a hidden floating panel wrapping child (pass 0 for an empty
// framed panel), with an optional title header. Make it visible with SetVisible
// and position it with SetBounds; while hidden it draws and dispatches nothing.
func (m *Module) Popover(child int, title string) (int, error) {
	var c toolkit.Widget
	if child != 0 {
		w, err := m.getWidget(child)
		if err != nil {
			return 0, err
		}
		c = w
	}
	p := toolkit.NewPopover(c)
	p.Title = title
	return m.reg(p), nil
}

// CommandPalette constructs a hidden centred spotlight overlay over commands — a
// Ruby Array of Hashes, each with a "label" (the searchable text) and an "action"
// (a callback identifier fired, and reported by Dispatch, when the command is
// chosen). Non-Hash elements are skipped, mirroring the Menu constructor. Open it
// with SetVisible(id, true), which clears any prior query + selection.
func (m *Module) CommandPalette(commands []any) int {
	cmds := make([]toolkit.PaletteCommand, 0, len(commands))
	for _, raw := range commands {
		h, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		c := toolkit.PaletteCommand{}
		if v, ok := h["label"]; ok {
			c.Label = fmt.Sprint(v)
		}
		if v, ok := h["action"]; ok {
			if cb := fmt.Sprint(v); cb != "" {
				c.Action = func() { m.fire(cb) }
			}
		}
		cmds = append(cmds, c)
	}
	return m.reg(toolkit.NewCommandPalette(cmds))
}

// IconButton constructs a compact toolbar tile whose whole face is the short
// glyph string icon ("+", "OK", "×"). When callback is non-empty it fires
// (reported by Dispatch) on every click.
func (m *Module) IconButton(icon, callback string) int {
	var onClick func()
	if callback != "" {
		onClick = func() { m.fire(callback) }
	}
	return m.reg(toolkit.NewIconButton(icon, onClick))
}

// Tooltip constructs a hidden text bubble. placement picks which side of its
// anchor the bubble sits on: "below" (the default), "above", "left" or "right".
// An unknown placement is an error. A host toggles it with SetVisible and
// positions it with SetBounds.
func (m *Module) Tooltip(text, placement string) (int, error) {
	pl, err := parsePlacement(placement)
	if err != nil {
		return 0, err
	}
	t := toolkit.NewTooltip(text)
	t.Placement = pl
	return m.reg(t), nil
}

// Avatar constructs a user-identity chip showing initials centred in a
// rounded-square body. color is the body fill as "#rrggbb"/"#rrggbbaa" hex; an
// empty string tracks the theme's Accent. A malformed colour is an error.
func (m *Module) Avatar(initials, color string) (int, error) {
	c, err := parseHexColor(color)
	if err != nil {
		return 0, err
	}
	a := toolkit.NewAvatar(initials)
	a.Color = c
	return m.reg(a), nil
}

// LevelBar constructs a discrete segmented indicator (battery / signal / VU
// meter) of max equal cells; the first Value cells fill in Accent (set with
// SetValue). max is floored at 1 by the toolkit.
func (m *Module) LevelBar(max int) int {
	return m.reg(toolkit.NewLevelBar(max))
}

// --- overlay state -----------------------------------------------------------

// SetVisible shows or hides a transient overlay: it toggles the Visible flag of
// a Notification, Toast, Popover or Tooltip, the Open flag of a ContextMenu, and
// Opens (v true, clearing query + selection) or Dismisses (v false) a
// CommandPalette. A handle that is not one of these is an error.
func (m *Module) SetVisible(id int, v bool) error {
	o, err := m.get(id)
	if err != nil {
		return err
	}
	switch w := o.(type) {
	case *toolkit.Notification:
		w.Visible = v
	case *toolkit.Toast:
		w.Visible = v
	case *toolkit.Popover:
		w.Visible = v
	case *toolkit.Tooltip:
		w.Visible = v
	case *toolkit.ContextMenu:
		w.Open = v
	case *toolkit.CommandPalette:
		if v {
			w.Open()
		} else {
			w.Dismiss()
		}
	default:
		return fmt.Errorf("widgets: SetVisible: handle %d (%T) is not a toggleable overlay", id, o)
	}
	return nil
}

// Visible reports whether a transient overlay (Notification, Toast, Popover,
// Tooltip, ContextMenu or CommandPalette) is currently shown. A handle that is
// not one of these is an error.
func (m *Module) Visible(id int) (bool, error) {
	o, err := m.get(id)
	if err != nil {
		return false, err
	}
	switch w := o.(type) {
	case *toolkit.Notification:
		return w.Visible, nil
	case *toolkit.Toast:
		return w.Visible, nil
	case *toolkit.Popover:
		return w.Visible, nil
	case *toolkit.Tooltip:
		return w.Visible, nil
	case *toolkit.ContextMenu:
		return w.Open, nil
	case *toolkit.CommandPalette:
		return w.Visible, nil
	default:
		return false, fmt.Errorf("widgets: Visible: handle %d (%T) is not a toggleable overlay", id, o)
	}
}

// Popup opens a ContextMenu anchored at (x, y) — the cursor point — so its next
// render draws the menu clamped inside the surface. The handle must be a
// ContextMenu.
func (m *Module) Popup(id, x, y int) error {
	o, err := m.get(id)
	if err != nil {
		return err
	}
	c, ok := o.(*toolkit.ContextMenu)
	if !ok {
		return fmt.Errorf("widgets: Popup: handle %d (%T) is not a context menu", id, o)
	}
	c.Popup(x, y)
	return nil
}

// AnchorIn sizes a Notification or Toast to its text and positions it at a corner
// of the host rect (x, y, w, h), inset by the widget's own margin: "top_left"
// (the default), "top_right", "bottom_left", "bottom_right", "top_center" or
// "bottom_center". An unknown corner, or a handle that is not a Notification or
// Toast, is an error.
func (m *Module) AnchorIn(id, x, y, w, h int, corner string) error {
	o, err := m.get(id)
	if err != nil {
		return err
	}
	cn, err := parseCorner(corner)
	if err != nil {
		return err
	}
	host := toolkit.Rect{X: x, Y: y, W: w, H: h}
	switch t := o.(type) {
	case *toolkit.Notification:
		t.AnchorIn(host, cn)
	case *toolkit.Toast:
		t.AnchorIn(host, cn, 0)
	default:
		return fmt.Errorf("widgets: AnchorIn: handle %d (%T) is not an anchorable overlay", id, o)
	}
	return nil
}

// SetLife sets the auto-dismiss countdown of a Notification or Toast: the number
// of Tick calls before it hides. For a Toast, 0 is the "sticky" sentinel (never
// auto-hide). A handle that is neither is an error.
func (m *Module) SetLife(id, n int) error {
	o, err := m.get(id)
	if err != nil {
		return err
	}
	switch w := o.(type) {
	case *toolkit.Notification:
		w.Life = n
	case *toolkit.Toast:
		w.Life = n
	default:
		return fmt.Errorf("widgets: SetLife: handle %d (%T) has no life budget", id, o)
	}
	return nil
}

// Tick advances a Notification or Toast one animation frame, decrementing its
// life and auto-hiding it when the countdown reaches 0. A host calls it once per
// frame from its render loop. A handle that is neither is an error.
func (m *Module) Tick(id int) error {
	o, err := m.get(id)
	if err != nil {
		return err
	}
	switch w := o.(type) {
	case *toolkit.Notification:
		w.Tick()
	case *toolkit.Toast:
		w.Tick()
	default:
		return fmt.Errorf("widgets: Tick: handle %d (%T) is not tickable", id, o)
	}
	return nil
}

// SetKind changes a Toast's severity ("info"/"success"/"warning"/"error"),
// re-tinting its pill. An unknown kind, or a handle that is not a Toast, is an
// error.
func (m *Module) SetKind(id int, kind string) error {
	o, err := m.get(id)
	if err != nil {
		return err
	}
	t, ok := o.(*toolkit.Toast)
	if !ok {
		return fmt.Errorf("widgets: SetKind: handle %d (%T) is not a toast", id, o)
	}
	k, err := parseToastKind(kind)
	if err != nil {
		return err
	}
	t.Kind = k
	return nil
}

// SetValue sets a LevelBar's filled-cell count (clamped by the widget to its
// cell range at draw time). A handle that is not a LevelBar is an error.
func (m *Module) SetValue(id, v int) error {
	o, err := m.get(id)
	if err != nil {
		return err
	}
	l, ok := o.(*toolkit.LevelBar)
	if !ok {
		return fmt.Errorf("widgets: SetValue: handle %d (%T) is not a level bar", id, o)
	}
	l.Value = v
	return nil
}

// --- small parsers / coercions -----------------------------------------------

// parseToastKind maps a kind name to a toolkit.ToastKind ("" == info).
func parseToastKind(name string) (toolkit.ToastKind, error) {
	switch strings.ToLower(name) {
	case "", "info":
		return toolkit.ToastInfo, nil
	case "success":
		return toolkit.ToastSuccess, nil
	case "warning":
		return toolkit.ToastWarning, nil
	case "error":
		return toolkit.ToastError, nil
	default:
		return 0, fmt.Errorf("widgets: unknown toast kind %q", name)
	}
}

// parsePlacement maps a placement name to a toolkit.TooltipPlacement ("" ==
// below).
func parsePlacement(name string) (toolkit.TooltipPlacement, error) {
	switch strings.ToLower(name) {
	case "", "below":
		return toolkit.PlaceBelow, nil
	case "above":
		return toolkit.PlaceAbove, nil
	case "left":
		return toolkit.PlaceLeft, nil
	case "right":
		return toolkit.PlaceRight, nil
	default:
		return 0, fmt.Errorf("widgets: unknown tooltip placement %q", name)
	}
}

// parseScaleMode maps a scale name to a toolkit.ScaleMode ("" == stretch).
func parseScaleMode(name string) (toolkit.ScaleMode, error) {
	switch strings.ToLower(name) {
	case "", "stretch":
		return toolkit.ScaleStretch, nil
	case "fit":
		return toolkit.ScaleFit, nil
	default:
		return 0, fmt.Errorf("widgets: unknown scale mode %q", name)
	}
}

// parseCorner maps a corner name to a toolkit.Corner ("" == top_left).
func parseCorner(name string) (toolkit.Corner, error) {
	switch strings.ToLower(name) {
	case "", "top_left":
		return toolkit.TopLeft, nil
	case "top_right":
		return toolkit.TopRight, nil
	case "bottom_left":
		return toolkit.BottomLeft, nil
	case "bottom_right":
		return toolkit.BottomRight, nil
	case "top_center":
		return toolkit.TopCenter, nil
	case "bottom_center":
		return toolkit.BottomCenter, nil
	default:
		return 0, fmt.Errorf("widgets: unknown corner %q", name)
	}
}

// toBytes coerces a Ruby-supplied pixel argument to a byte slice: a []byte (a
// binary String surfaces as one) passes through, a string is base64-decoded, and
// nil yields a nil slice. Any other type — or an undecodable base64 string — is
// an error.
func toBytes(v any) ([]byte, error) {
	switch x := v.(type) {
	case nil:
		return nil, nil
	case []byte:
		return x, nil
	case string:
		b, err := base64.StdEncoding.DecodeString(x)
		if err != nil {
			return nil, fmt.Errorf("widgets: invalid base64 pixel data: %w", err)
		}
		return b, nil
	default:
		return nil, fmt.Errorf("widgets: cannot use %T as pixel bytes", v)
	}
}
