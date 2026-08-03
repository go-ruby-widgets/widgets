// Copyright (c) 2026 the go-ruby-widgets/widgets authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package widgets

import (
	"encoding/base64"
	"testing"
)

// rgba22 is a 2x2 RGBA pixel buffer for the image-backed constructors.
func rgba22() []byte {
	b := make([]byte, 2*2*4)
	for i := range b {
		b[i] = byte(i)
	}
	return b
}

// --- tray --------------------------------------------------------------------

func TestStatusIconEveryStockIcon(t *testing.T) {
	m := NewModule()
	// Every stock name plus the empty (nil-glyph) slot must construct.
	for _, name := range []string{
		"", "new", "open", "save", "cut", "copy", "paste", "undo", "redo",
		"search", "settings", "NEW", // case-insensitive
	} {
		if _, err := m.StatusIcon(name, "tip", "", ""); err != nil {
			t.Errorf("StatusIcon(%q): %v", name, err)
		}
	}
	// A wired variant (both callbacks) and an unknown name.
	if id, err := m.StatusIcon("settings", "Settings", "on_open", "on_menu"); err != nil || id == 0 {
		t.Fatalf("StatusIcon wired: id=%d err=%v", id, err)
	}
	if _, err := m.StatusIcon("frobnicate", "", "", ""); err == nil {
		t.Error("StatusIcon unknown name should error")
	}
}

func TestStatusIconImage(t *testing.T) {
	m := NewModule()
	raw := rgba22()

	// Raw []byte, then a base64 string, both wired with callbacks.
	if _, err := m.StatusIconImage(raw, 2, 2, "tip", "on_click", "on_right"); err != nil {
		t.Fatalf("StatusIconImage []byte: %v", err)
	}
	b64 := base64.StdEncoding.EncodeToString(raw)
	if _, err := m.StatusIconImage(b64, 2, 2, "", "", ""); err != nil {
		t.Fatalf("StatusIconImage base64: %v", err)
	}

	// Error branches: bad base64, non-positive size, short buffer.
	if _, err := m.StatusIconImage("not*base64", 2, 2, "", "", ""); err == nil {
		t.Error("StatusIconImage bad base64 should error")
	}
	if _, err := m.StatusIconImage(raw, 0, 2, "", "", ""); err == nil {
		t.Error("StatusIconImage zero width should error")
	}
	if _, err := m.StatusIconImage(raw[:4], 2, 2, "", "", ""); err == nil {
		t.Error("StatusIconImage short buffer should error")
	}
}

func TestStatusAreaAndAddSeam(t *testing.T) {
	m := NewModule()
	area := m.StatusArea()
	if area == 0 {
		t.Fatal("StatusArea handle should be non-zero")
	}
	icon, _ := m.StatusIcon("search", "Find", "", "")

	// StatusIcon joins the StatusArea via the shared AddWidget seam.
	if err := m.AddWidget(area, icon); err != nil {
		t.Fatalf("AddWidget area<-icon: %v", err)
	}
	// A non-icon child of a StatusArea is rejected.
	label := m.Label("x")
	if err := m.AddWidget(area, label); err == nil {
		t.Error("AddWidget area<-label should error")
	}

	// A Badge attaches to a StatusIcon via the same seam.
	badge, _ := m.Badge("3", "", "")
	if err := m.AddWidget(icon, badge); err != nil {
		t.Fatalf("AddWidget icon<-badge: %v", err)
	}
	// A non-badge child of a StatusIcon is rejected.
	if err := m.AddWidget(icon, label); err == nil {
		t.Error("AddWidget icon<-label should error")
	}
}

// --- wallpaper ---------------------------------------------------------------

func TestWallpaper(t *testing.T) {
	m := NewModule()
	raw := rgba22()

	for _, mode := range []string{"", "fill", "fit", "center", "centre", "tile"} {
		if _, err := m.Wallpaper(raw, 2, 2, mode); err != nil {
			t.Errorf("Wallpaper(%q): %v", mode, err)
		}
	}
	// base64 source.
	b64 := base64.StdEncoding.EncodeToString(raw)
	if _, err := m.Wallpaper(b64, 2, 2, "fill"); err != nil {
		t.Fatalf("Wallpaper base64: %v", err)
	}

	// Error branches: bad base64, non-positive size, short buffer, unknown mode.
	if _, err := m.Wallpaper("not*base64", 2, 2, ""); err == nil {
		t.Error("Wallpaper bad base64 should error")
	}
	if _, err := m.Wallpaper(raw, 2, 0, ""); err == nil {
		t.Error("Wallpaper zero height should error")
	}
	if _, err := m.Wallpaper(raw[:4], 2, 2, ""); err == nil {
		t.Error("Wallpaper short buffer should error")
	}
	if _, err := m.Wallpaper(raw, 2, 2, "sideways"); err == nil {
		t.Error("Wallpaper unknown mode should error")
	}
}

func TestWallpaperGradient(t *testing.T) {
	m := NewModule()
	if _, err := m.WallpaperGradient("", ""); err != nil {
		t.Fatalf("WallpaperGradient theme default: %v", err)
	}
	if _, err := m.WallpaperGradient("#101820", "#2E4057"); err != nil {
		t.Fatalf("WallpaperGradient colours: %v", err)
	}
	if _, err := m.WallpaperGradient("nothex", ""); err == nil {
		t.Error("WallpaperGradient bad top should error")
	}
	if _, err := m.WallpaperGradient("", "nothex"); err == nil {
		t.Error("WallpaperGradient bad bottom should error")
	}
}

// --- thumbnail ---------------------------------------------------------------

func TestThumbnail(t *testing.T) {
	m := NewModule()
	raw := rgba22()

	// Labelled + wired, then a bare (label-less, callback-less) base64 tile.
	if _, err := m.Thumbnail(raw, 2, 2, "term", "on_pick"); err != nil {
		t.Fatalf("Thumbnail labelled: %v", err)
	}
	b64 := base64.StdEncoding.EncodeToString(raw)
	if _, err := m.Thumbnail(b64, 2, 2, "", ""); err != nil {
		t.Fatalf("Thumbnail base64: %v", err)
	}

	// Error branches: bad base64, non-positive size, short buffer.
	if _, err := m.Thumbnail("not*base64", 2, 2, "", ""); err == nil {
		t.Error("Thumbnail bad base64 should error")
	}
	if _, err := m.Thumbnail(raw, 0, 2, "", ""); err == nil {
		t.Error("Thumbnail zero width should error")
	}
	if _, err := m.Thumbnail(raw[:4], 2, 2, "", ""); err == nil {
		t.Error("Thumbnail short buffer should error")
	}
}

func TestSetSelectedAndSetHover(t *testing.T) {
	m := NewModule()
	th, _ := m.Thumbnail(rgba22(), 2, 2, "win", "")

	if err := m.SetSelected(th, true); err != nil {
		t.Fatalf("SetSelected: %v", err)
	}
	if err := m.SetHover(th, true); err != nil {
		t.Fatalf("SetHover: %v", err)
	}

	// Wrong-type handle.
	lbl := m.Label("x")
	if err := m.SetSelected(lbl, true); err == nil {
		t.Error("SetSelected on a label should error")
	}
	if err := m.SetHover(lbl, true); err == nil {
		t.Error("SetHover on a label should error")
	}
	// Unknown handle.
	if err := m.SetSelected(9999, true); err == nil {
		t.Error("SetSelected on unknown handle should error")
	}
	if err := m.SetHover(9999, true); err == nil {
		t.Error("SetHover on unknown handle should error")
	}
}

// --- callbacks fire via Dispatch ---------------------------------------------

func TestDesktopCallbacksFire(t *testing.T) {
	m := NewModule()

	// StatusIcon: a primary click fires OnClick, a secondary (Code "right")
	// click fires OnRightClick.
	icon, _ := m.StatusIcon("settings", "S", "on_click", "on_right")
	out, _ := m.Dispatch(icon, map[string]any{"kind": "click"})
	if fired := out["fired"].([]any); len(fired) != 1 || fired[0] != "on_click" {
		t.Errorf("status icon primary click: %v", out)
	}
	out, _ = m.Dispatch(icon, map[string]any{"kind": "click", "code": "right"})
	if fired := out["fired"].([]any); len(fired) != 1 || fired[0] != "on_right" {
		t.Errorf("status icon right click: %v", out)
	}

	// The image-backed StatusIcon wires the same closures.
	img, _ := m.StatusIconImage(rgba22(), 2, 2, "", "on_iclick", "on_iright")
	out, _ = m.Dispatch(img, map[string]any{"kind": "click"})
	if fired := out["fired"].([]any); len(fired) != 1 || fired[0] != "on_iclick" {
		t.Errorf("image status icon primary click: %v", out)
	}
	out, _ = m.Dispatch(img, map[string]any{"kind": "click", "code": "right"})
	if fired := out["fired"].([]any); len(fired) != 1 || fired[0] != "on_iright" {
		t.Errorf("image status icon right click: %v", out)
	}

	// Thumbnail click fires OnClick.
	th, _ := m.Thumbnail(rgba22(), 2, 2, "win", "on_pick")
	out, _ = m.Dispatch(th, map[string]any{"kind": "click"})
	if fired := out["fired"].([]any); len(fired) != 1 || fired[0] != "on_pick" {
		t.Errorf("thumbnail click: %v", out)
	}
}

// --- reflective Call surface -------------------------------------------------

func TestDesktopCallSurface(t *testing.T) {
	m := NewModule()
	raw := rgba22()
	b64 := base64.StdEncoding.EncodeToString(raw)

	area, err := Call(m, "status_area")
	if err != nil {
		t.Fatalf("Call status_area: %v", err)
	}
	icon, err := Call(m, "status_icon", "search", "Find", "on_c", "on_r")
	if err != nil {
		t.Fatalf("Call status_icon: %v", err)
	}
	if _, err := Call(m, "status_icon_image", b64, 2, 2, "tip", "", ""); err != nil {
		t.Fatalf("Call status_icon_image: %v", err)
	}
	if _, err := Call(m, "add_widget", area.(int), icon.(int)); err != nil {
		t.Fatalf("Call add_widget area<-icon: %v", err)
	}

	if _, err := Call(m, "wallpaper", b64, 2, 2, "fit"); err != nil {
		t.Fatalf("Call wallpaper: %v", err)
	}
	if _, err := Call(m, "wallpaper_gradient", "#101820", "#2E4057"); err != nil {
		t.Fatalf("Call wallpaper_gradient: %v", err)
	}

	th, err := Call(m, "thumbnail", b64, 2, 2, "win", "on_pick")
	if err != nil {
		t.Fatalf("Call thumbnail: %v", err)
	}
	if _, err := Call(m, "set_selected", th.(int), true); err != nil {
		t.Fatalf("Call set_selected: %v", err)
	}
	if _, err := Call(m, "set_hover", th.(int), true); err != nil {
		t.Fatalf("Call set_hover: %v", err)
	}

	// The new names surface in the reflective method list.
	names := map[string]bool{}
	for _, n := range Methods(m) {
		names[n] = true
	}
	for _, want := range []string{
		"status_area", "status_icon", "status_icon_image", "wallpaper",
		"wallpaper_gradient", "thumbnail", "set_selected", "set_hover",
	} {
		if !names[want] {
			t.Errorf("Methods() missing %q", want)
		}
	}
}
