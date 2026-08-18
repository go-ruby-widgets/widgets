// Copyright (c) 2026 the go-ruby-widgets/widgets authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package widgets

import (
	"encoding/base64"
	"testing"

	"github.com/go-widgets/toolkit"
)

// --- Toast refinements: icon / multi-line / multi-action ----------------------

func TestSetToastIconGlyph(t *testing.T) {
	m := NewModule()
	toast, _ := m.Toast("saved", "success", "", "")
	// Every stock glyph name resolves to a vector Icon and clears any pixels.
	for _, name := range []string{
		"new", "open", "save", "cut", "copy", "paste",
		"undo", "redo", "search", "settings",
	} {
		if err := m.SetToastIcon(toast, name, 0, 0); err != nil {
			t.Fatalf("SetToastIcon(%q): %v", name, err)
		}
		tk := m.objs[toast].(*toolkit.Toast)
		if tk.Icon == nil || tk.Pixels != nil {
			t.Errorf("glyph %q: want Icon set, pixels cleared", name)
		}
	}
	// A pixel icon supersedes a prior glyph and clears Icon.
	raw := make([]byte, 2*2*4)
	if err := m.SetToastIcon(toast, raw, 2, 2); err != nil {
		t.Fatalf("SetToastIcon pixels: %v", err)
	}
	tk := m.objs[toast].(*toolkit.Toast)
	if tk.Icon != nil || tk.IW != 2 || tk.IH != 2 || len(tk.Pixels) != 16 {
		t.Errorf("pixel icon: want pixels set, Icon cleared, got %+v", tk)
	}
	// A base64 String pixel icon also works.
	if err := m.SetToastIcon(toast, base64.StdEncoding.EncodeToString(raw), 2, 2); err != nil {
		t.Fatalf("SetToastIcon base64: %v", err)
	}
}

func TestSetToastIconErrors(t *testing.T) {
	m := NewModule()
	toast, _ := m.Toast("x", "", "", "")
	// Non-glyph string that is not valid base64.
	if err := m.SetToastIcon(toast, "!!!not-base64!!!", 4, 4); err == nil {
		t.Error("SetToastIcon: bad base64 should error")
	}
	// Non-positive size for a pixel buffer.
	if err := m.SetToastIcon(toast, make([]byte, 16), 0, 2); err == nil {
		t.Error("SetToastIcon: non-positive size should error")
	}
	// Short buffer.
	if err := m.SetToastIcon(toast, make([]byte, 4), 2, 2); err == nil {
		t.Error("SetToastIcon: short buffer should error")
	}
	// Unknown handle.
	if err := m.SetToastIcon(999, "save", 0, 0); err == nil {
		t.Error("SetToastIcon: unknown handle should error")
	}
	// Non-toast handle.
	if err := m.SetToastIcon(m.Label("x"), "save", 0, 0); err == nil {
		t.Error("SetToastIcon: non-toast should error")
	}
	// A non-string, non-bytes argument cannot be pixel data.
	if err := m.SetToastIcon(toast, 42, 2, 2); err == nil {
		t.Error("SetToastIcon: int argument should error")
	}
}

func TestSetToastLines(t *testing.T) {
	m := NewModule()
	toast, _ := m.Toast("x", "", "", "")
	if err := m.SetToastLines(toast, []any{"Title", "Body line"}); err != nil {
		t.Fatalf("SetToastLines: %v", err)
	}
	if got := m.objs[toast].(*toolkit.Toast).Lines; len(got) != 2 || got[0] != "Title" {
		t.Errorf("Lines not set: %v", got)
	}
	// An empty array reverts to the single-Text look (nil Lines).
	if err := m.SetToastLines(toast, nil); err != nil {
		t.Fatalf("SetToastLines nil: %v", err)
	}
	if m.objs[toast].(*toolkit.Toast).Lines != nil {
		t.Error("empty array should clear Lines")
	}
	if err := m.SetToastLines(999, []any{"a"}); err == nil {
		t.Error("SetToastLines: unknown handle should error")
	}
	if err := m.SetToastLines(m.Label("x"), []any{"a"}); err == nil {
		t.Error("SetToastLines: non-toast should error")
	}
}

func TestSetToastActions(t *testing.T) {
	m := NewModule()
	toast, _ := m.Toast("x", "", "", "")
	actions := []any{
		map[string]any{"label": "Undo", "callback": "on_undo"},
		map[string]any{"label": "Retry", "callback": ""}, // empty callback -> no fire wiring
		"not-a-hash",                       // skipped
		map[string]any{"label": "Dismiss"}, // no callback key
	}
	if err := m.SetToastActions(toast, actions); err != nil {
		t.Fatalf("SetToastActions: %v", err)
	}
	tk := m.objs[toast].(*toolkit.Toast)
	if len(tk.Actions) != 3 {
		t.Fatalf("want 3 actions (non-hash skipped), got %d", len(tk.Actions))
	}
	if tk.Actions[0].Label != "Undo" || tk.Actions[0].Callback == nil {
		t.Error("first action mis-wired")
	}
	if tk.Actions[1].Callback != nil || tk.Actions[2].Callback != nil {
		t.Error("empty/absent callback should stay nil")
	}
	// The wired action fires through the same seam Dispatch drains.
	m.fired = nil
	tk.Actions[0].Callback()
	if len(m.fired) != 1 || m.fired[0] != "on_undo" {
		t.Errorf("action callback did not fire: %v", m.fired)
	}
	// An empty array reverts to the single ActionLabel/Action pair (nil Actions).
	if err := m.SetToastActions(toast, nil); err != nil {
		t.Fatalf("SetToastActions nil: %v", err)
	}
	if m.objs[toast].(*toolkit.Toast).Actions != nil {
		t.Error("empty array should clear Actions")
	}
	if err := m.SetToastActions(999, actions); err == nil {
		t.Error("SetToastActions: unknown handle should error")
	}
	if err := m.SetToastActions(m.Label("x"), actions); err == nil {
		t.Error("SetToastActions: non-toast should error")
	}
}

// --- Toast button rects (host-side per-button routing) ------------------------

func TestButtonRects(t *testing.T) {
	m := NewModule()
	toast, _ := m.Toast("Deleted", "info", "", "")
	if err := m.SetToastActions(toast, []any{
		map[string]any{"label": "Undo", "callback": "on_undo"},
		map[string]any{"label": "Dismiss", "callback": "on_dismiss"},
	}); err != nil {
		t.Fatalf("SetToastActions: %v", err)
	}
	if err := m.SetBounds(toast, 0, 0, 300, 44); err != nil {
		t.Fatalf("SetBounds: %v", err)
	}

	got, err := m.ButtonRects(toast)
	if err != nil {
		t.Fatalf("ButtonRects: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 button rects, got %d", len(got))
	}

	// The binding is byte-faithful to the toolkit geometry, exactly.
	tk := m.objs[toast].(*toolkit.Toast)
	want := tk.ButtonRects()
	for i, g := range got {
		h, ok := g.(map[string]any)
		if !ok {
			t.Fatalf("rect %d is not a Hash: %T", i, g)
		}
		if h["x"] != want[i].X || h["y"] != want[i].Y || h["w"] != want[i].W || h["h"] != want[i].H {
			t.Fatalf("rect %d = %v, want %+v", i, h, want[i])
		}
		// Local-space contract: Y at the pill top, full pill height.
		if h["y"].(int) != 0 || h["h"].(int) != 44 {
			t.Fatalf("rect %d spans wrong vertical extent: %v", i, h)
		}
	}

	// Edge-to-edge tiling + the last button flush with the pill's right edge (300).
	r0, r1 := got[0].(map[string]any), got[1].(map[string]any)
	if r1["x"].(int) != r0["x"].(int)+r0["w"].(int) {
		t.Fatalf("buttons not contiguous: %v then %v", r0, r1)
	}
	if end := r1["x"].(int) + r1["w"].(int); end != 300 {
		t.Fatalf("last button ends at %d, want 300", end)
	}

	// Routing: a click at each rect's CENTRE, dispatched through the same OnEvent
	// seam a host would drive, fires exactly that button's callback (index-precise).
	names := []string{"on_undo", "on_dismiss"}
	for i, g := range got {
		h := g.(map[string]any)
		cx := h["x"].(int) + h["w"].(int)/2
		cy := h["y"].(int) + h["h"].(int)/2
		if err := m.SetVisible(toast, true); err != nil {
			t.Fatalf("SetVisible: %v", err)
		}
		out, err := m.Dispatch(toast, map[string]any{"kind": "click", "x": cx, "y": cy})
		if err != nil {
			t.Fatalf("Dispatch button %d: %v", i, err)
		}
		fired := out["fired"].([]any)
		if len(fired) != 1 || fired[0] != names[i] {
			t.Fatalf("click centre of button %d fired %v, want [%s]", i, fired, names[i])
		}
	}

	// An action-less toast yields an empty Array (not nil-crashing).
	plain, _ := m.Toast("hi", "", "", "")
	if err := m.SetBounds(plain, 0, 0, 120, 30); err != nil {
		t.Fatalf("SetBounds plain: %v", err)
	}
	pr, err := m.ButtonRects(plain)
	if err != nil {
		t.Fatalf("ButtonRects plain: %v", err)
	}
	if len(pr) != 0 {
		t.Fatalf("action-less toast rects = %v, want empty", pr)
	}

	// The snake_case dispatch name reaches the accessor.
	cv, err := Call(m, "button_rects", toast)
	if err != nil {
		t.Fatalf("Call button_rects: %v", err)
	}
	if len(cv.([]any)) != 2 {
		t.Fatalf("Call button_rects len = %d, want 2", len(cv.([]any)))
	}

	// Error paths: unknown handle + non-toast handle.
	if _, err := m.ButtonRects(999); err == nil {
		t.Error("ButtonRects: unknown handle should error")
	}
	if _, err := m.ButtonRects(m.Label("x")); err == nil {
		t.Error("ButtonRects: non-toast should error")
	}
}

// --- Label font size ----------------------------------------------------------

func TestSetFontSize(t *testing.T) {
	m := NewModule()
	lbl := m.Label("12:00")
	if err := m.SetFontSize(lbl, 48); err != nil {
		t.Fatalf("SetFontSize: %v", err)
	}
	if m.objs[lbl].(*toolkit.Label).FontSize != 48 {
		t.Error("FontSize did not stick")
	}
	if err := m.SetFontSize(999, 20); err == nil {
		t.Error("SetFontSize: unknown handle should error")
	}
	if err := m.SetFontSize(m.Button("b", ""), 20); err == nil {
		t.Error("SetFontSize: non-label should error")
	}
}

// --- Calendar -----------------------------------------------------------------

func TestCalendarControls(t *testing.T) {
	m := NewModule()
	cal := m.Calendar(2026, 8, 3)
	c := m.objs[cal].(*toolkit.Calendar)
	if c.Year().Get() != 2026 || c.Month().Get() != 8 || c.Day().Get() != 3 {
		t.Fatalf("Calendar ctor: %+v", c)
	}

	// Month-change callback fires on Prev/Next.
	if err := m.OnMonthChange(cal, "on_month"); err != nil {
		t.Fatalf("OnMonthChange: %v", err)
	}
	// Direct Month() and Year() changes each fire (Prev/Next below moves only the
	// Month, so a direct Year() Set covers the year subscriber). Restore afterwards.
	m.fired = nil
	c.Month().Set(12)
	if len(m.fired) != 1 || m.fired[0] != "on_month" {
		t.Errorf("month-set: fired=%v", m.fired)
	}
	m.fired = nil
	c.Year().Set(2027)
	if len(m.fired) != 1 || m.fired[0] != "on_month" {
		t.Errorf("year-set: fired=%v", m.fired)
	}
	c.Year().Set(2026)
	c.Month().Set(8)
	m.fired = nil
	if err := m.NextMonth(cal); err != nil {
		t.Fatalf("NextMonth: %v", err)
	}
	if c.Month().Get() != 9 || len(m.fired) != 1 || m.fired[0] != "on_month" {
		t.Errorf("NextMonth: month=%d fired=%v", c.Month().Get(), m.fired)
	}
	m.fired = nil
	if err := m.PrevMonth(cal); err != nil {
		t.Fatalf("PrevMonth: %v", err)
	}
	if c.Month().Get() != 8 || len(m.fired) != 1 {
		t.Errorf("PrevMonth: month=%d fired=%v", c.Month().Get(), m.fired)
	}
	// Clearing the wiring silences the callback.
	if err := m.OnMonthChange(cal, ""); err != nil {
		t.Fatalf("OnMonthChange clear: %v", err)
	}
	m.fired = nil
	m.NextMonth(cal)
	if len(m.fired) != 0 {
		t.Errorf("cleared OnMonthChange still fired: %v", m.fired)
	}

	// Selection: SetSelected sets the day, Selected reads it back.
	if err := m.SetSelected(cal, 15); err != nil {
		t.Fatalf("SetSelected(calendar): %v", err)
	}
	if got, err := m.Selected(cal); err != nil || got != 15 {
		t.Fatalf("Selected(calendar): %d %v", got, err)
	}
	// A non-integer day errors.
	if err := m.SetSelected(cal, "x"); err == nil {
		t.Error("SetSelected(calendar): non-int day should error")
	}

	// Select callback fires on a day click routed through Dispatch.
	if err := m.OnSelect(cal, "on_day"); err != nil {
		t.Fatalf("OnSelect: %v", err)
	}
	m.SetBounds(cal, 0, 0, 210, 200)
	// Header-arrow click also drives Prev/Next (exercise the header path harmlessly).
	m.Dispatch(cal, map[string]any{"kind": "click", "x": 5, "y": 5})
	// A grid click: sweep cells until one lands on a valid day and fires OnSelect.
	fired := false
	for y := 60; y < 190 && !fired; y += 20 {
		for x := 5; x < 205 && !fired; x += 30 {
			out, _ := m.Dispatch(cal, map[string]any{"kind": "click", "x": x, "y": y})
			for _, f := range out["fired"].([]any) {
				if f == "on_day" {
					fired = true
				}
			}
		}
	}
	if !fired {
		t.Error("OnSelect never fired on any grid click")
	}
	// Clearing OnSelect detaches the subscription: a day change no longer fires.
	if err := m.OnSelect(cal, ""); err != nil {
		t.Fatalf("OnSelect clear: %v", err)
	}
	m.fired = nil
	c.Day().Set(c.Day().Get() + 1)
	if len(m.fired) != 0 {
		t.Errorf("OnSelect not cleared: fired=%v", m.fired)
	}
}

func TestCalendarErrors(t *testing.T) {
	m := NewModule()
	lbl := m.Label("x")
	for _, tc := range []struct {
		name string
		fn   func() error
	}{
		{"PrevMonth", func() error { return m.PrevMonth(lbl) }},
		{"NextMonth", func() error { return m.NextMonth(lbl) }},
		{"OnSelect", func() error { return m.OnSelect(lbl, "cb") }},
		{"OnMonthChange", func() error { return m.OnMonthChange(lbl, "cb") }},
	} {
		if err := tc.fn(); err == nil {
			t.Errorf("%s on non-calendar should error", tc.name)
		}
	}
	// Unknown handle through the shared resolver.
	if err := m.PrevMonth(999); err == nil {
		t.Error("PrevMonth unknown handle should error")
	}
}

// --- LevelBar caption + thresholds -------------------------------------------

func TestLevelBarLabelAndThresholds(t *testing.T) {
	m := NewModule()
	lb, err := m.LevelBar(10, "Battery", []any{
		map[string]any{"min": 0, "color_hex": "#e00000"},
		map[string]any{"min": 4, "color_hex": "#e0a000"},
		map[string]any{"min": 8, "color_hex": "#00a000"},
		"skip-me", // non-hash skipped
	})
	if err != nil {
		t.Fatalf("LevelBar: %v", err)
	}
	l := m.objs[lb].(*toolkit.LevelBar)
	if l.Label != "Battery" || len(l.Thresholds) != 3 {
		t.Fatalf("label/thresholds: %q %v", l.Label, l.Thresholds)
	}
	if l.Thresholds[2].Min != 8 || l.Thresholds[2].Color.A == 0 {
		t.Errorf("threshold not parsed: %+v", l.Thresholds[2])
	}
	// A malformed colour_hex is an error.
	if _, err := m.LevelBar(4, "", []any{map[string]any{"min": 1, "color_hex": "zzz"}}); err == nil {
		t.Error("LevelBar: bad color_hex should error")
	}
	// Plain bar via omitted trailing args (empty label, nil thresholds).
	if _, err := m.LevelBar(3, "", nil); err != nil {
		t.Fatalf("plain LevelBar: %v", err)
	}
}

// --- CommandPalette accessors (host-drivable Spotlight) ----------------------

func TestCommandPaletteAccessors(t *testing.T) {
	m := NewModule()
	cp := m.CommandPalette([]any{
		map[string]any{"label": "Open File", "action": "on_open"},
		map[string]any{"label": "Save File", "action": "on_save"},
		map[string]any{"label": "Close", "action": "on_close"},
	})

	// Query round-trips and re-filters.
	if err := m.SetQuery(cp, "file"); err != nil {
		t.Fatalf("SetQuery: %v", err)
	}
	if q, err := m.Query(cp); err != nil || q != "file" {
		t.Fatalf("Query: %q %v", q, err)
	}
	got, err := m.FilteredCommands(cp)
	if err != nil {
		t.Fatalf("FilteredCommands: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 filtered ('file'), got %d: %v", len(got), got)
	}
	if got[0].(map[string]any)["label"] != "Open File" {
		t.Errorf("filtered order: %v", got)
	}

	// Selection: move + set + read.
	if err := m.SetSelected(cp, 1); err != nil {
		t.Fatalf("SetSelected(palette): %v", err)
	}
	if s, err := m.Selected(cp); err != nil || s != 1 {
		t.Fatalf("Selected(palette): %d %v", s, err)
	}
	if err := m.MoveSelection(cp, -1); err != nil {
		t.Fatalf("MoveSelection: %v", err)
	}
	if s, _ := m.Selected(cp); s != 0 {
		t.Errorf("MoveSelection: want 0, got %d", s)
	}
	// A non-integer selection index errors.
	if err := m.SetSelected(cp, "nope"); err == nil {
		t.Error("SetSelected(palette): non-int index should error")
	}

	// HandleKey: type into the query, then Enter fires the selected command.
	m.SetQuery(cp, "")
	out, err := m.HandleKey(cp, map[string]any{"kind": "char", "code": "c"})
	if err != nil {
		t.Fatalf("HandleKey char: %v", err)
	}
	if len(out["fired"].([]any)) != 0 {
		t.Error("typing a char should fire nothing")
	}
	if q, _ := m.Query(cp); q != "c" {
		t.Errorf("HandleKey char did not extend query: %q", q)
	}
	// "c" matches only "Close" -> Enter activates it.
	out, err = m.HandleKey(cp, map[string]any{"kind": "keydown", "code": "Enter"})
	if err != nil {
		t.Fatalf("HandleKey Enter: %v", err)
	}
	if fired := out["fired"].([]any); len(fired) != 1 || fired[0] != "on_close" {
		t.Errorf("HandleKey Enter: %v", fired)
	}
	// A malformed event Hash errors.
	if _, err := m.HandleKey(cp, map[string]any{"kind": "bogus"}); err == nil {
		t.Error("HandleKey: bad event kind should error")
	}
}

func TestCommandPaletteAccessorErrors(t *testing.T) {
	m := NewModule()
	lbl := m.Label("x")
	if err := m.SetQuery(lbl, "q"); err == nil {
		t.Error("SetQuery on non-palette should error")
	}
	if _, err := m.Query(lbl); err == nil {
		t.Error("Query on non-palette should error")
	}
	if err := m.MoveSelection(lbl, 1); err == nil {
		t.Error("MoveSelection on non-palette should error")
	}
	if _, err := m.FilteredCommands(lbl); err == nil {
		t.Error("FilteredCommands on non-palette should error")
	}
	if _, err := m.HandleKey(lbl, map[string]any{"kind": "char", "code": "a"}); err == nil {
		t.Error("HandleKey on non-palette should error")
	}
	// Unknown handles.
	if _, err := m.Query(999); err == nil {
		t.Error("Query unknown handle should error")
	}
	// Selected / SetSelected type coverage.
	if _, err := m.Selected(lbl); err == nil {
		t.Error("Selected on unselectable widget should error")
	}
	if _, err := m.Selected(999); err == nil {
		t.Error("Selected unknown handle should error")
	}
	if err := m.SetSelected(lbl, 1); err == nil {
		t.Error("SetSelected on unselectable widget should error")
	}
	if err := m.SetSelected(999, 1); err == nil {
		t.Error("SetSelected unknown handle should error")
	}
}

// --- reflective surface: every new method reachable through Call / Methods ----

func TestRefinementCallSurface(t *testing.T) {
	m := NewModule()

	toast, _ := Call(m, "toast", "saved")
	if _, err := Call(m, "set_toast_icon", toast.(int), "save", 0, 0); err != nil {
		t.Fatalf("Call set_toast_icon: %v", err)
	}
	if _, err := Call(m, "set_toast_lines", toast.(int), []any{"Title", "Body"}); err != nil {
		t.Fatalf("Call set_toast_lines: %v", err)
	}
	if _, err := Call(m, "set_toast_actions", toast.(int),
		[]any{map[string]any{"label": "Undo", "callback": "on_undo"}}); err != nil {
		t.Fatalf("Call set_toast_actions: %v", err)
	}

	lbl := m.Label("clock")
	if _, err := Call(m, "set_font_size", lbl, 40); err != nil {
		t.Fatalf("Call set_font_size: %v", err)
	}

	cal, err := Call(m, "calendar", 2026, 8, 3)
	if err != nil {
		t.Fatalf("Call calendar: %v", err)
	}
	for _, name := range []string{"prev_month", "next_month"} {
		if _, err := Call(m, name, cal.(int)); err != nil {
			t.Fatalf("Call %s: %v", name, err)
		}
	}
	if _, err := Call(m, "on_select", cal.(int), "on_day"); err != nil {
		t.Fatalf("Call on_select: %v", err)
	}
	if _, err := Call(m, "on_month_change", cal.(int), "on_month"); err != nil {
		t.Fatalf("Call on_month_change: %v", err)
	}

	// level_bar with a caption + thresholds via Call.
	if _, err := Call(m, "level_bar", 8, "Signal",
		[]any{map[string]any{"min": 4, "color_hex": "#00a000"}}); err != nil {
		t.Fatalf("Call level_bar: %v", err)
	}

	cp, err := Call(m, "command_palette",
		[]any{map[string]any{"label": "Go", "action": "on_go"}})
	if err != nil {
		t.Fatalf("Call command_palette: %v", err)
	}
	cpID := cp.(int)
	if _, err := Call(m, "set_query", cpID, "g"); err != nil {
		t.Fatalf("Call set_query: %v", err)
	}
	if q, err := Call(m, "query", cpID); err != nil || q.(string) != "g" {
		t.Fatalf("Call query: %v %v", q, err)
	}
	if _, err := Call(m, "set_selected", cpID, 0); err != nil {
		t.Fatalf("Call set_selected: %v", err)
	}
	if s, err := Call(m, "selected", cpID); err != nil || s.(int) != 0 {
		t.Fatalf("Call selected: %v %v", s, err)
	}
	if _, err := Call(m, "move_selection", cpID, 1); err != nil {
		t.Fatalf("Call move_selection: %v", err)
	}
	if fc, err := Call(m, "filtered_commands", cpID); err != nil || len(fc.([]any)) != 1 {
		t.Fatalf("Call filtered_commands: %v %v", fc, err)
	}
	if _, err := Call(m, "handle_key", cpID, map[string]any{"kind": "keydown", "code": "Escape"}); err != nil {
		t.Fatalf("Call handle_key: %v", err)
	}

	names := Methods(m)
	for _, want := range []string{
		"set_toast_icon", "set_toast_lines", "set_toast_actions",
		"set_font_size", "calendar", "prev_month", "next_month",
		"on_select", "on_month_change", "level_bar", "set_query", "query",
		"selected", "set_selected", "move_selection", "filtered_commands",
		"handle_key",
	} {
		if !sortedContains(names, want) {
			t.Errorf("method %q missing from reflective listing", want)
		}
	}
}
