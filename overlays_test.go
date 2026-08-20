// Copyright (c) 2026 the go-ruby-widgets/widgets authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package widgets

import (
	"encoding/base64"
	"testing"

	"github.com/go-widgets/toolkit"
)

// --- overlay + chrome constructors -------------------------------------------

func TestOverlayConstructors(t *testing.T) {
	m := NewModule()

	if id := m.Notification("saved"); id == 0 {
		t.Fatal("Notification handle should be non-zero")
	}

	// Toast: bare (info), an explicit kind, and one with a wired action button;
	// then a bad kind.
	if _, err := m.Toast("copied", "", "", ""); err != nil {
		t.Fatalf("Toast bare: %v", err)
	}
	if _, err := m.Toast("done", "success", "Undo", "on_undo"); err != nil {
		t.Fatalf("Toast with action: %v", err)
	}
	if _, err := m.Toast("act", "info", "Undo", ""); err != nil { // action label, no callback
		t.Fatalf("Toast action no cb: %v", err)
	}
	if _, err := m.Toast("x", "chartreuse", "", ""); err == nil {
		t.Error("Toast bad kind should error")
	}

	// Badge: plain, coloured, and a malformed colour on each of fill/ink.
	if _, err := m.Badge("12", "", ""); err != nil {
		t.Fatalf("Badge plain: %v", err)
	}
	if _, err := m.Badge("NEW", "#C03030", "#FFFFFF"); err != nil {
		t.Fatalf("Badge coloured: %v", err)
	}
	if _, err := m.Badge("x", "nothex", ""); err == nil {
		t.Error("Badge bad fill should error")
	}
	if _, err := m.Badge("x", "", "nothex"); err == nil {
		t.Error("Badge bad ink should error")
	}

	// IconButton wired + bare.
	if id := m.IconButton("+", "on_add"); id == 0 {
		t.Fatal("IconButton handle should be non-zero")
	}
	m.IconButton("-", "")

	// Tooltip every placement + a bad one.
	for _, pl := range []string{"", "below", "above", "left", "right"} {
		if _, err := m.Tooltip("tip", pl); err != nil {
			t.Errorf("Tooltip(%q): %v", pl, err)
		}
	}
	if _, err := m.Tooltip("tip", "sideways"); err == nil {
		t.Error("Tooltip bad placement should error")
	}

	// Avatar plain, coloured, bad colour.
	if _, err := m.Avatar("DD", ""); err != nil {
		t.Fatalf("Avatar plain: %v", err)
	}
	if _, err := m.Avatar("DD", "#2E8B57"); err != nil {
		t.Fatalf("Avatar coloured: %v", err)
	}
	if _, err := m.Avatar("DD", "zzz"); err == nil {
		t.Error("Avatar bad colour should error")
	}

	// LevelBar (max floored to 1 by the toolkit for a non-positive arg).
	if id, err := m.LevelBar(4, "", nil); err != nil || id == 0 {
		t.Fatalf("LevelBar handle should be non-zero: %d %v", id, err)
	}
	if _, err := m.LevelBar(0, "", nil); err != nil {
		t.Fatalf("LevelBar(0): %v", err)
	}
}

func TestImageConstructor(t *testing.T) {
	m := NewModule()
	const w, h = 2, 2
	raw := make([]byte, w*h*4)
	for i := range raw {
		raw[i] = byte(i)
	}

	// Raw []byte, stretch (default) + fit.
	if _, err := m.Image(raw, w, h, ""); err != nil {
		t.Fatalf("Image []byte stretch: %v", err)
	}
	if _, err := m.Image(raw, w, h, "fit"); err != nil {
		t.Fatalf("Image []byte fit: %v", err)
	}

	// base64 string.
	b64 := base64.StdEncoding.EncodeToString(raw)
	if _, err := m.Image(b64, w, h, "stretch"); err != nil {
		t.Fatalf("Image base64: %v", err)
	}

	// nil pixels -> nil slice -> too-short error branch.
	if _, err := m.Image(nil, w, h, ""); err == nil {
		t.Error("Image nil pixels should error (short buffer)")
	}
	// Error branches: bad base64, non-positive size, short buffer, bad scale,
	// unsupported pixel type.
	if _, err := m.Image("not*base64", w, h, ""); err == nil {
		t.Error("Image bad base64 should error")
	}
	if _, err := m.Image(raw, 0, h, ""); err == nil {
		t.Error("Image zero width should error")
	}
	if _, err := m.Image(raw[:4], w, h, ""); err == nil {
		t.Error("Image short buffer should error")
	}
	if _, err := m.Image(raw, w, h, "warp"); err == nil {
		t.Error("Image bad scale should error")
	}
	if _, err := m.Image(42, w, h, ""); err == nil {
		t.Error("Image non-bytes pixels should error")
	}
}

func TestContextMenuAndPopover(t *testing.T) {
	m := NewModule()

	// ContextMenu wraps a Menu; bad handle + non-menu error.
	menu := m.Menu([]any{map[string]any{"label": "Open", "action": "on_open"}})
	cm, err := m.ContextMenu(menu)
	if err != nil {
		t.Fatalf("ContextMenu: %v", err)
	}
	if _, err := m.ContextMenu(999); err == nil {
		t.Error("ContextMenu bad handle should error")
	}
	if _, err := m.ContextMenu(m.HBox()); err == nil {
		t.Error("ContextMenu non-menu should error")
	}

	// Popup opens the menu at a point + errors on the wrong type / handle.
	if err := m.Popup(cm, 10, 20); err != nil {
		t.Fatalf("Popup: %v", err)
	}
	if err := m.Popup(999, 0, 0); err == nil {
		t.Error("Popup bad handle should error")
	}
	if err := m.Popup(m.HBox(), 0, 0); err == nil {
		t.Error("Popup non-context-menu should error")
	}

	// Popover: with a child, an empty (0) panel, and a bad child handle.
	if _, err := m.Popover(m.Label("body"), "Details"); err != nil {
		t.Fatalf("Popover with child: %v", err)
	}
	if _, err := m.Popover(0, ""); err != nil {
		t.Fatalf("Popover empty: %v", err)
	}
	if _, err := m.Popover(999, ""); err == nil {
		t.Error("Popover bad child should error")
	}
}

func TestCommandPalette(t *testing.T) {
	m := NewModule()
	id := m.CommandPalette([]any{
		map[string]any{"label": "New File", "action": "on_new"},
		map[string]any{"label": "Noop", "action": ""}, // action-less
		7, // not a Hash -> skipped
	})
	cp := m.objs[id].(*toolkit.CommandPalette)
	if len(cp.Commands) != 2 {
		t.Fatalf("want 2 commands, got %d", len(cp.Commands))
	}
	if cp.Commands[1].Action != nil {
		t.Error("empty action id should leave Action nil")
	}
}

// --- overlay state -----------------------------------------------------------

func TestSetVisibleAndVisible(t *testing.T) {
	m := NewModule()
	menu := m.Menu([]any{map[string]any{"label": "A", "action": "on_a"}})
	cm, _ := m.ContextMenu(menu)
	toggles := []int{
		m.Notification("n"),
		func() int { id, _ := m.Toast("t", "", "", ""); return id }(),
		func() int { id, _ := m.Popover(0, ""); return id }(),
		func() int { id, _ := m.Tooltip("tip", ""); return id }(),
		cm,
		m.CommandPalette(nil),
	}
	for _, id := range toggles {
		if err := m.SetVisible(id, true); err != nil {
			t.Fatalf("SetVisible(%d,true): %v", id, err)
		}
		if v, err := m.Visible(id); err != nil || !v {
			t.Errorf("Visible(%d) after show = %v,%v", id, v, err)
		}
		if err := m.SetVisible(id, false); err != nil {
			t.Fatalf("SetVisible(%d,false): %v", id, err)
		}
		if v, _ := m.Visible(id); v {
			t.Errorf("Visible(%d) after hide should be false", id)
		}
	}
	// error branches.
	if err := m.SetVisible(999, true); err == nil {
		t.Error("SetVisible bad handle")
	}
	if err := m.SetVisible(m.Label("x"), true); err == nil {
		t.Error("SetVisible non-overlay")
	}
	if _, err := m.Visible(999); err == nil {
		t.Error("Visible bad handle")
	}
	if _, err := m.Visible(m.Label("x")); err == nil {
		t.Error("Visible non-overlay")
	}
}

func TestAnchorIn(t *testing.T) {
	m := NewModule()
	notif := m.Notification("hi")
	toast, _ := m.Toast("yo", "", "", "")
	host := []int{0, 0, 400, 300}
	for _, corner := range []string{"", "top_left", "top_right", "bottom_left", "bottom_right", "top_center", "bottom_center"} {
		if err := m.AnchorIn(notif, host[0], host[1], host[2], host[3], corner); err != nil {
			t.Errorf("AnchorIn notification %q: %v", corner, err)
		}
		if err := m.AnchorIn(toast, host[0], host[1], host[2], host[3], corner); err != nil {
			t.Errorf("AnchorIn toast %q: %v", corner, err)
		}
	}
	// A corner-anchored notification gets a positive, on-host rect.
	b, _ := m.Bounds(notif)
	if b["w"].(int) <= 0 || b["h"].(int) <= 0 {
		t.Errorf("AnchorIn should size the notification, got %v", b)
	}
	// error branches: bad handle, bad corner, non-anchorable type.
	if err := m.AnchorIn(999, 0, 0, 10, 10, ""); err == nil {
		t.Error("AnchorIn bad handle")
	}
	if err := m.AnchorIn(notif, 0, 0, 10, 10, "sideways"); err == nil {
		t.Error("AnchorIn bad corner")
	}
	if err := m.AnchorIn(m.Label("x"), 0, 0, 10, 10, ""); err == nil {
		t.Error("AnchorIn non-anchorable")
	}
}

func TestLifeTickAndKind(t *testing.T) {
	m := NewModule()
	notif := m.Notification("n")
	toast, _ := m.Toast("t", "", "", "")

	for _, id := range []int{notif, toast} {
		if err := m.SetLife(id, 3); err != nil {
			t.Fatalf("SetLife(%d): %v", id, err)
		}
	}
	// Show + tick down a toast to the auto-hide edge.
	m.SetVisible(toast, true)
	m.SetLife(toast, 1)
	if err := m.Tick(toast); err != nil {
		t.Fatalf("Tick toast: %v", err)
	}
	if v, _ := m.Visible(toast); v {
		t.Error("toast should auto-hide when life reaches 0")
	}
	m.SetVisible(notif, true)
	if err := m.Tick(notif); err != nil {
		t.Fatalf("Tick notification: %v", err)
	}

	// SetLife / Tick error branches.
	if err := m.SetLife(999, 1); err == nil {
		t.Error("SetLife bad handle")
	}
	if err := m.SetLife(m.Label("x"), 1); err == nil {
		t.Error("SetLife non-lifed")
	}
	if err := m.Tick(999); err == nil {
		t.Error("Tick bad handle")
	}
	if err := m.Tick(m.Label("x")); err == nil {
		t.Error("Tick non-tickable")
	}

	// SetKind: every kind + errors.
	for _, k := range []string{"", "info", "success", "warning", "error"} {
		if err := m.SetKind(toast, k); err != nil {
			t.Errorf("SetKind(%q): %v", k, err)
		}
	}
	if err := m.SetKind(toast, "puce"); err == nil {
		t.Error("SetKind bad kind")
	}
	if err := m.SetKind(999, "info"); err == nil {
		t.Error("SetKind bad handle")
	}
	if err := m.SetKind(m.Label("x"), "info"); err == nil {
		t.Error("SetKind non-toast")
	}
}

func TestSetValue(t *testing.T) {
	m := NewModule()
	lb, _ := m.LevelBar(5, "", nil)
	if err := m.SetValue(lb, 3); err != nil {
		t.Fatalf("SetValue: %v", err)
	}
	if m.objs[lb].(*toolkit.LevelBar).Value().Get() != 3 {
		t.Error("SetValue did not stick")
	}
	if err := m.SetValue(999, 1); err == nil {
		t.Error("SetValue bad handle")
	}
	if err := m.SetValue(m.Label("x"), 1); err == nil {
		t.Error("SetValue non-level-bar")
	}
}

// --- SetText / Text over the new text-bearing overlays -----------------------

func TestOverlayTextAccessors(t *testing.T) {
	m := NewModule()
	notif := m.Notification("a")
	toast, _ := m.Toast("a", "", "", "")
	tip, _ := m.Tooltip("a", "")
	badge, _ := m.Badge("a", "", "")
	icon := m.IconButton("a", "")
	av, _ := m.Avatar("a", "")
	for _, id := range []int{notif, toast, tip, badge, icon, av} {
		if err := m.SetText(id, "Z"); err != nil {
			t.Fatalf("SetText(%d): %v", id, err)
		}
		got, err := m.Text(id)
		if err != nil {
			t.Fatalf("Text(%d): %v", id, err)
		}
		if got != "Z" {
			t.Errorf("handle %d: want Z, got %q", id, got)
		}
	}
}

// --- reflective surface: every new method reachable through Call / Methods ----

func TestOverlayCallSurface(t *testing.T) {
	m := NewModule()

	// Constructors via Call return int handles.
	nv, err := Call(m, "notification", "hi")
	if err != nil {
		t.Fatalf("Call notification: %v", err)
	}
	notif := nv.(int)

	// Omitted-trailing-args ergonomics: toast(text) alone (kind/action default "").
	tv, err := Call(m, "toast", "copied")
	if err != nil {
		t.Fatalf("Call toast: %v", err)
	}
	toast := tv.(int)

	// badge(text) alone; command_palette over an Array; level_bar(max).
	if _, err := Call(m, "badge", "12"); err != nil {
		t.Fatalf("Call badge: %v", err)
	}
	if _, err := Call(m, "command_palette", []any{map[string]any{"label": "Go", "action": "on_go"}}); err != nil {
		t.Fatalf("Call command_palette: %v", err)
	}
	lv, err := Call(m, "level_bar", 4)
	if err != nil {
		t.Fatalf("Call level_bar: %v", err)
	}

	// State methods via Call.
	if _, err := Call(m, "set_visible", notif, true); err != nil {
		t.Fatalf("Call set_visible: %v", err)
	}
	if vis, err := Call(m, "visible", notif); err != nil || vis.(bool) != true {
		t.Fatalf("Call visible: %v %v", vis, err)
	}
	if _, err := Call(m, "anchor_in", notif, 0, 0, 200, 100, "top_right"); err != nil {
		t.Fatalf("Call anchor_in: %v", err)
	}
	if _, err := Call(m, "set_life", toast, 2); err != nil {
		t.Fatalf("Call set_life: %v", err)
	}
	if _, err := Call(m, "tick", toast); err != nil {
		t.Fatalf("Call tick: %v", err)
	}
	if _, err := Call(m, "set_kind", toast, "warning"); err != nil {
		t.Fatalf("Call set_kind: %v", err)
	}
	if _, err := Call(m, "set_value", lv.(int), 2); err != nil {
		t.Fatalf("Call set_value: %v", err)
	}

	// image via Call with a base64 String argument.
	raw := make([]byte, 2*2*4)
	if _, err := Call(m, "image", base64.StdEncoding.EncodeToString(raw), 2, 2, "fit"); err != nil {
		t.Fatalf("Call image: %v", err)
	}

	// context_menu + popup + popover + icon_button + tooltip + avatar via Call.
	menu, _ := Call(m, "menu", []any{map[string]any{"label": "X", "action": "on_x"}})
	cm, err := Call(m, "context_menu", menu.(int))
	if err != nil {
		t.Fatalf("Call context_menu: %v", err)
	}
	if _, err := Call(m, "popup", cm.(int), 5, 5); err != nil {
		t.Fatalf("Call popup: %v", err)
	}
	if _, err := Call(m, "popover", 0, "Title"); err != nil {
		t.Fatalf("Call popover: %v", err)
	}
	if _, err := Call(m, "icon_button", "+", "on_add"); err != nil {
		t.Fatalf("Call icon_button: %v", err)
	}
	if _, err := Call(m, "tooltip", "tip", "above"); err != nil {
		t.Fatalf("Call tooltip: %v", err)
	}
	if _, err := Call(m, "avatar", "DD", ""); err != nil {
		t.Fatalf("Call avatar: %v", err)
	}

	// All new names present in the reflective listing.
	names := Methods(m)
	for _, want := range []string{
		"notification", "toast", "badge", "image", "context_menu", "popover",
		"command_palette", "icon_button", "tooltip", "avatar", "level_bar",
		"set_visible", "visible", "popup", "anchor_in", "set_life", "tick",
		"set_kind", "set_value",
	} {
		if !sortedContains(names, want) {
			t.Errorf("method %q missing from reflective listing", want)
		}
	}
}

// A wired Toast/CommandPalette/IconButton callback fires through Dispatch, the
// same seam the base widgets use.
func TestOverlayCallbacksFire(t *testing.T) {
	m := NewModule()

	// IconButton click fires its callback.
	ib := m.IconButton("+", "on_add")
	out, _ := m.Dispatch(ib, map[string]any{"kind": "click"})
	if fired := out["fired"].([]any); len(fired) != 1 || fired[0] != "on_add" {
		t.Errorf("icon button dispatch: %v", out)
	}

	// Toast action-button click fires its action (right edge of the pill).
	toast, _ := m.Toast("done", "success", "Undo", "on_undo")
	m.SetVisible(toast, true)
	m.SetBounds(toast, 0, 0, 200, 24)
	out, _ = m.Dispatch(toast, map[string]any{"kind": "click", "x": 195, "y": 12})
	if fired := out["fired"].([]any); len(fired) != 1 || fired[0] != "on_undo" {
		t.Errorf("toast action dispatch: %v", out)
	}

	// CommandPalette: opening it, then activating the sole command with Enter,
	// fires that command's wired action.
	cp := m.CommandPalette([]any{map[string]any{"label": "Go", "action": "on_go"}})
	m.SetVisible(cp, true)
	m.SetBounds(cp, 0, 0, 400, 300)
	out, _ = m.Dispatch(cp, map[string]any{"kind": "keydown", "code": "Enter"})
	if fired := out["fired"].([]any); len(fired) != 1 || fired[0] != "on_go" {
		t.Errorf("command palette activation: %v", out)
	}
}
