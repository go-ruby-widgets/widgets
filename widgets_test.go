// Copyright (c) 2026 the go-ruby-widgets/widgets authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package widgets

import (
	"reflect"
	"testing"

	"github.com/go-widgets/toolkit"
)

// --- constructors ------------------------------------------------------------

func TestConstructors(t *testing.T) {
	m := NewModule()
	// Every constructor with a wired callback, then again bare (no callback) so
	// both the wired and unwired branches run.
	m.Button("OK", "cb")
	m.Button("OK", "")
	m.Label("hi")
	m.Entry("x", "cb")
	m.Entry("x", "")
	m.TextView("multi\nline")
	m.CheckButton("c", true, "cb")
	m.CheckButton("c", false, "")
	m.DropDown([]any{"a", "b"}, 0, "cb")
	m.DropDown([]any{"a"}, 0, "")
	m.ListBox([]any{"r0", "r1"}, "cb")
	m.ListBox([]any{"r0"}, "")
	m.MenuBar()
	m.HBox()
	m.VBox()
	m.Grid(2, 2)
	m.Border()
	if id := m.Label(""); id == 0 {
		t.Fatal("handle should count from 1")
	}
}

func TestMenuConstruction(t *testing.T) {
	m := NewModule()
	// A rich item, a separator item, an action-less item, and a non-Hash element
	// (skipped via continue). separator given as a truthy non-bool to exercise
	// truthy's default branch.
	id := m.Menu([]any{
		map[string]any{"label": "Open", "shortcut": "Ctrl+O", "action": "on_open"},
		map[string]any{"separator": 1},
		map[string]any{"label": "Noop", "action": ""},
		42, // not a Hash -> skipped
	})
	mn := m.objs[id].(*toolkit.Menu)
	if len(mn.Items) != 3 {
		t.Fatalf("want 3 items, got %d", len(mn.Items))
	}
	if !mn.Items[1].Separator {
		t.Error("item 1 should be a separator")
	}
	if mn.Items[2].Action != nil {
		t.Error("empty action id should leave Action nil")
	}
}

func TestContainerLayouts(t *testing.T) {
	m := NewModule()
	for _, name := range []string{"", "fit", "box", "hbox", "vbox", "border", "card"} {
		if _, err := m.Container(name); err != nil {
			t.Errorf("Container(%q): %v", name, err)
		}
	}
	if _, err := m.Container("nope"); err == nil {
		t.Error("Container(nope) should error")
	}
}

func TestFrameAndDock(t *testing.T) {
	m := NewModule()
	child := m.Label("x")
	if _, err := m.Frame(0); err != nil { // empty frame
		t.Error(err)
	}
	if _, err := m.Frame(child); err != nil {
		t.Error(err)
	}
	if _, err := m.Frame(999); err == nil {
		t.Error("Frame(bad child) should error")
	}
	if _, err := m.Dock(0); err != nil { // bars-only
		t.Error(err)
	}
	if _, err := m.Dock(child); err != nil {
		t.Error(err)
	}
	if _, err := m.Dock(999); err == nil {
		t.Error("Dock(bad body) should error")
	}
}

// --- mutation ----------------------------------------------------------------

func TestSetTextAndText(t *testing.T) {
	m := NewModule()
	lbl := m.Label("a")
	btn := m.Button("b", "")
	ent := m.Entry("c", "")
	tv := m.TextView("d")
	chk := m.CheckButton("e", false, "")
	for _, id := range []int{lbl, btn, ent, tv, chk} {
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
	// error branches: unknown handle + a handle with no text (a container).
	box := m.HBox()
	if err := m.SetText(999, "x"); err == nil {
		t.Error("SetText(bad handle)")
	}
	if err := m.SetText(box, "x"); err == nil {
		t.Error("SetText(hbox) should error")
	}
	if _, err := m.Text(999); err == nil {
		t.Error("Text(bad handle)")
	}
	if _, err := m.Text(box); err == nil {
		t.Error("Text(hbox) should error")
	}
}

func TestCheckedAndSelectAndStyleSpacing(t *testing.T) {
	m := NewModule()
	chk := m.CheckButton("c", false, "")
	if err := m.SetChecked(chk, true); err != nil {
		t.Fatal(err)
	}
	if v, _ := m.Checked(chk); !v {
		t.Error("want checked")
	}
	if err := m.SetChecked(999, true); err == nil {
		t.Error("SetChecked bad handle")
	}
	lbl := m.Label("x")
	if err := m.SetChecked(lbl, true); err == nil {
		t.Error("SetChecked non-check")
	}
	if _, err := m.Checked(999); err == nil {
		t.Error("Checked bad handle")
	}
	if _, err := m.Checked(lbl); err == nil {
		t.Error("Checked non-check")
	}

	// Select on DropDown (fires its callback) and ListBox.
	dd := m.DropDown([]any{"a", "b"}, 0, "onsel")
	if err := m.Select(dd, 1); err != nil {
		t.Fatal(err)
	}
	if len(m.fired) != 1 || m.fired[0] != "onsel" {
		t.Errorf("dropdown callback not fired: %v", m.fired)
	}
	lb := m.ListBox([]any{"r0", "r1"}, "")
	if err := m.Select(lb, 1); err != nil {
		t.Fatal(err)
	}
	if err := m.Select(999, 0); err == nil {
		t.Error("Select bad handle")
	}
	if err := m.Select(lbl, 0); err == nil {
		t.Error("Select non-selectable")
	}

	// SetStyle every case + errors.
	btn := m.Button("b", "")
	for _, s := range []string{"prominent", "secondary", "default", ""} {
		if err := m.SetStyle(btn, s); err != nil {
			t.Errorf("SetStyle(%q): %v", s, err)
		}
	}
	if err := m.SetStyle(btn, "weird"); err == nil {
		t.Error("SetStyle(weird)")
	}
	if err := m.SetStyle(999, "default"); err == nil {
		t.Error("SetStyle bad handle")
	}
	if err := m.SetStyle(lbl, "default"); err == nil {
		t.Error("SetStyle non-button")
	}

	// SetSpacing HBox/VBox/Grid + errors.
	for _, id := range []int{m.HBox(), m.VBox(), m.Grid(1, 1)} {
		if err := m.SetSpacing(id, 8); err != nil {
			t.Errorf("SetSpacing(%d): %v", id, err)
		}
	}
	if err := m.SetSpacing(999, 8); err == nil {
		t.Error("SetSpacing bad handle")
	}
	if err := m.SetSpacing(lbl, 8); err == nil {
		t.Error("SetSpacing non-box")
	}
}

func TestSetTheme(t *testing.T) {
	m := NewModule()
	for _, n := range []string{"", "light", "dark"} {
		if err := m.SetTheme(n); err != nil {
			t.Errorf("SetTheme(%q): %v", n, err)
		}
	}
	if err := m.SetTheme("chartreuse"); err == nil {
		t.Error("SetTheme(bad)")
	}
}

func TestUseOpentypeText(t *testing.T) {
	// The toolkit's active font is a process-global; restore the bitmap default
	// so this test does not perturb the pixel/metric geometry other tests read.
	defer toolkit.SetFont(nil)

	m := NewModule()
	if err := m.UseOpentypeText(); err != nil {
		t.Fatalf("UseOpentypeText: %v", err)
	}
	if err := m.UseOpentypeTextSize(20); err != nil {
		t.Fatalf("UseOpentypeTextSize: %v", err)
	}

	// Reachable through the reflective surface under the snake_case names the
	// Ruby binding drives, both directions (Call in, Methods out).
	if _, err := Call(m, "use_opentype_text"); err != nil {
		t.Fatalf("Call use_opentype_text: %v", err)
	}
	if _, err := Call(m, "use_opentype_text_size", 18); err != nil {
		t.Fatalf("Call use_opentype_text_size: %v", err)
	}
	names := Methods(m)
	if !sortedContains(names, "use_opentype_text") || !sortedContains(names, "use_opentype_text_size") {
		t.Errorf("use_opentype_text[_size] missing from %v", names)
	}
}

// --- composition -------------------------------------------------------------

func TestAddWidget(t *testing.T) {
	m := NewModule()
	c, _ := m.Container("vbox")
	hb := m.HBox()
	vb := m.VBox()
	child := func() int { return m.Label("x") }
	for _, p := range []int{c, hb, vb} {
		if err := m.AddWidget(p, child()); err != nil {
			t.Errorf("AddWidget into %d: %v", p, err)
		}
	}
	if err := m.AddWidget(999, child()); err == nil {
		t.Error("AddWidget bad parent")
	}
	if err := m.AddWidget(c, 999); err == nil {
		t.Error("AddWidget bad child")
	}
	if err := m.AddWidget(m.Label("l"), child()); err == nil {
		t.Error("AddWidget into non-container")
	}
}

func TestAddWithConfig(t *testing.T) {
	m := NewModule()
	c, _ := m.Container("border")
	if err := m.Add(c, m.Label("n"), map[string]any{"region": "north", "size": 20}); err != nil {
		t.Fatal(err)
	}
	if err := m.Add(c, m.Label("c"), map[string]any{"flex": 2}); err != nil {
		t.Fatal(err)
	}
	if err := m.Add(c, m.Label("z"), nil); err != nil { // nil opts
		t.Fatal(err)
	}
	// Every region name flows through parseRegion.
	for _, r := range []string{"south", "east", "west", "center", ""} {
		if err := m.Add(c, m.Label(r), map[string]any{"region": r}); err != nil {
			t.Errorf("Add region %q: %v", r, err)
		}
	}
	if err := m.Add(c, m.Label("b"), map[string]any{"region": "bogus"}); err == nil {
		t.Error("Add bad region")
	}
	if err := m.Add(999, m.Label("x"), nil); err == nil {
		t.Error("Add bad parent")
	}
	if err := m.Add(c, 999, nil); err == nil {
		t.Error("Add bad child")
	}
	if err := m.Add(m.HBox(), m.Label("x"), nil); err == nil {
		t.Error("Add into non-container")
	}
}

func TestBoxAddFixedFlexAndGridAttach(t *testing.T) {
	m := NewModule()
	hb := m.HBox()
	vb := m.VBox()
	if err := m.AddFixed(hb, m.Label("x"), 30); err != nil {
		t.Fatal(err)
	}
	if err := m.AddFixed(vb, m.Label("x"), 30); err != nil {
		t.Fatal(err)
	}
	if err := m.AddFlex(hb, m.Label("x"), 2); err != nil {
		t.Fatal(err)
	}
	if err := m.AddFlex(vb, m.Label("x"), 2); err != nil {
		t.Fatal(err)
	}
	if err := m.AddFixed(999, m.Label("x"), 1); err == nil {
		t.Error("AddFixed bad parent")
	}
	if err := m.AddFixed(hb, 999, 1); err == nil {
		t.Error("AddFixed bad child")
	}
	if err := m.AddFixed(m.Label("l"), m.Label("x"), 1); err == nil {
		t.Error("AddFixed non-box")
	}
	if err := m.AddFlex(999, m.Label("x"), 1); err == nil {
		t.Error("AddFlex bad parent")
	}
	if err := m.AddFlex(m.Label("l"), m.Label("x"), 1); err == nil {
		t.Error("AddFlex non-box")
	}

	g := m.Grid(2, 2)
	if err := m.Attach(g, m.Label("x"), 1, 1); err != nil {
		t.Fatal(err)
	}
	if err := m.Attach(999, m.Label("x"), 0, 0); err == nil {
		t.Error("Attach bad parent")
	}
	if err := m.Attach(m.HBox(), m.Label("x"), 0, 0); err == nil {
		t.Error("Attach non-grid")
	}
}

func TestDockAt(t *testing.T) {
	m := NewModule()
	d, _ := m.Dock(m.Label("body"))
	for _, side := range []string{"top", "bottom", "left", "right"} {
		if err := m.DockAt(d, m.Label(side), side, 24); err != nil {
			t.Errorf("DockAt(%q): %v", side, err)
		}
	}
	if err := m.DockAt(d, m.Label("x"), "nowhere", 10); err == nil {
		t.Error("DockAt bad side")
	}
	if err := m.DockAt(999, m.Label("x"), "top", 10); err == nil {
		t.Error("DockAt bad parent")
	}
	if err := m.DockAt(m.HBox(), m.Label("x"), "top", 10); err == nil {
		t.Error("DockAt non-dock")
	}
}

func TestSetRegion(t *testing.T) {
	m := NewModule()
	b := m.Border()
	for _, r := range []string{"north", "south", "east", "west", "center", ""} {
		if err := m.SetRegion(b, m.Label(r), r, 16); err != nil {
			t.Errorf("SetRegion(%q): %v", r, err)
		}
	}
	if err := m.SetRegion(b, m.Label("x"), "middle", 1); err == nil {
		t.Error("SetRegion bad region")
	}
	if err := m.SetRegion(999, m.Label("x"), "north", 1); err == nil {
		t.Error("SetRegion bad parent")
	}
	if err := m.SetRegion(m.HBox(), m.Label("x"), "north", 1); err == nil {
		t.Error("SetRegion non-border")
	}
}

func TestAddMenu(t *testing.T) {
	m := NewModule()
	bar := m.MenuBar()
	menu := m.Menu([]any{map[string]any{"label": "New", "action": "on_new"}})
	if err := m.AddMenu(bar, "File", menu); err != nil {
		t.Fatal(err)
	}
	if err := m.AddMenu(999, "File", menu); err == nil {
		t.Error("AddMenu bad bar handle")
	}
	if err := m.AddMenu(m.HBox(), "File", menu); err == nil {
		t.Error("AddMenu non-bar")
	}
	if err := m.AddMenu(bar, "File", 999); err == nil {
		t.Error("AddMenu bad menu handle")
	}
	if err := m.AddMenu(bar, "File", m.HBox()); err == nil {
		t.Error("AddMenu non-menu")
	}
}

func TestSetActiveAndSetLayout(t *testing.T) {
	m := NewModule()
	c, _ := m.Container("card")
	m.AddWidget(c, m.Label("card0"))
	m.AddWidget(c, m.Label("card1"))
	if err := m.SetActive(c, 1); err != nil {
		t.Fatal(err)
	}
	if err := m.SetActive(999, 0); err == nil {
		t.Error("SetActive bad handle")
	}
	if err := m.SetActive(m.HBox(), 0); err == nil {
		t.Error("SetActive non-container")
	}
	fit, _ := m.Container("fit")
	if err := m.SetActive(fit, 0); err == nil {
		t.Error("SetActive on non-card layout")
	}

	if err := m.SetLayout(c, "vbox"); err != nil {
		t.Fatal(err)
	}
	if err := m.SetLayout(999, "fit"); err == nil {
		t.Error("SetLayout bad handle")
	}
	if err := m.SetLayout(m.HBox(), "fit"); err == nil {
		t.Error("SetLayout non-container")
	}
	if err := m.SetLayout(c, "bogus"); err == nil {
		t.Error("SetLayout bad layout")
	}
}

// --- layout / query / render -------------------------------------------------

func TestSetBoundsLayoutBounds(t *testing.T) {
	m := NewModule()
	lbl := m.Label("x")
	if err := m.SetBounds(lbl, 1, 2, 3, 4); err != nil {
		t.Fatal(err)
	}
	got, err := m.Bounds(lbl)
	if err != nil {
		t.Fatal(err)
	}
	if got["x"] != 1 || got["y"] != 2 || got["w"] != 3 || got["h"] != 4 {
		t.Errorf("bounds mismatch: %v", got)
	}
	if err := m.Layout(lbl, 50, 60); err != nil {
		t.Fatal(err)
	}
	if got, _ := m.Bounds(lbl); got["w"] != 50 || got["h"] != 60 || got["x"] != 0 {
		t.Errorf("layout mismatch: %v", got)
	}
	if err := m.SetBounds(999, 0, 0, 1, 1); err == nil {
		t.Error("SetBounds bad handle")
	}
	if _, err := m.Bounds(999); err == nil {
		t.Error("Bounds bad handle")
	}
}

func TestRender(t *testing.T) {
	m := NewModule()
	btn := m.Button("", "") // no label so the fill is a uniform Surface colour
	const w, h = 40, 20
	img, err := m.Render(btn, w, h)
	if err != nil {
		t.Fatal(err)
	}
	if img["stride"] != w*4 || img["w"] != w || img["h"] != h {
		t.Errorf("render header wrong: %v", img)
	}
	buf := img["pixels"].([]byte)
	if len(buf) != 4*w*h {
		t.Fatalf("buffer len %d, want %d", len(buf), 4*w*h)
	}
	// Known pixel: the centre of a default-light button is the Surface colour.
	surface := toolkit.DefaultLight().Surface
	off := (h/2*w + w/2) * 4
	if buf[off] != surface.R || buf[off+1] != surface.G || buf[off+2] != surface.B || buf[off+3] != 0xFF {
		t.Errorf("centre pixel = %v,%v,%v,%v; want %v,%v,%v,255",
			buf[off], buf[off+1], buf[off+2], buf[off+3], surface.R, surface.G, surface.B)
	}
	// Non-empty: at least one opaque pixel was written.
	opaque := false
	for i := 3; i < len(buf); i += 4 {
		if buf[i] == 0xFF {
			opaque = true
			break
		}
	}
	if !opaque {
		t.Error("render produced no opaque pixels")
	}
	// error branches.
	if _, err := m.Render(999, w, h); err == nil {
		t.Error("Render bad handle")
	}
	if _, err := m.Render(btn, 0, h); err == nil {
		t.Error("Render zero width")
	}
	// A dark theme renders a different centre pixel.
	m.SetTheme("dark")
	dimg, _ := m.Render(btn, w, h)
	dbuf := dimg["pixels"].([]byte)
	if dbuf[off] == buf[off] && dbuf[off+1] == buf[off+1] && dbuf[off+2] == buf[off+2] {
		t.Error("dark theme should change the rendered pixel")
	}
}

// --- events ------------------------------------------------------------------

func TestDispatch(t *testing.T) {
	m := NewModule()

	// Button click -> callback fires, repaint true.
	btn := m.Button("OK", "on_ok")
	out, err := m.Dispatch(btn, map[string]any{"kind": "click"})
	if err != nil {
		t.Fatal(err)
	}
	fired := out["fired"].([]any)
	if len(fired) != 1 || fired[0] != "on_ok" || out["repaint"] != true {
		t.Errorf("button dispatch: %v", out)
	}

	// Label click -> no callback, repaint false.
	lbl := m.Label("x")
	out, _ = m.Dispatch(lbl, map[string]any{"kind": "click"})
	if len(out["fired"].([]any)) != 0 || out["repaint"] != false {
		t.Errorf("label dispatch should not repaint: %v", out)
	}

	// CheckButton click -> toggle callback.
	chk := m.CheckButton("c", false, "on_tog")
	out, _ = m.Dispatch(chk, map[string]any{"kind": "click"})
	if out["fired"].([]any)[0] != "on_tog" {
		t.Errorf("check dispatch: %v", out)
	}

	// Entry char (OnChange) and Enter (OnSubmit); exercises x/y/code/ctrl/shift.
	ent := m.Entry("", "on_ent")
	out, _ = m.Dispatch(ent, map[string]any{"kind": "char", "code": "a", "x": 1, "y": 0, "ctrl": true, "shift": 1})
	if len(out["fired"].([]any)) == 0 {
		t.Errorf("entry char should fire OnChange: %v", out)
	}
	m.SetText(ent, "hello")
	out, _ = m.Dispatch(ent, map[string]any{"kind": "keydown", "code": "Enter"})
	if out["fired"].([]any)[0] != "on_ent" {
		t.Errorf("entry Enter should fire OnSubmit: %v", out)
	}

	// ListBox click -> activate (needs bounds so the row hit-tests).
	lb := m.ListBox([]any{"r0", "r1"}, "on_row")
	m.SetBounds(lb, 0, 0, 100, 100)
	out, _ = m.Dispatch(lb, map[string]any{"kind": "click", "x": 5, "y": 5})
	if out["fired"].([]any)[0] != "on_row" {
		t.Errorf("listbox dispatch: %v", out)
	}

	// Menu click on the action row.
	menu := m.Menu([]any{map[string]any{"label": "Open", "action": "on_open"}})
	out, _ = m.Dispatch(menu, map[string]any{"kind": "click", "x": 10, "y": 10})
	if out["fired"].([]any)[0] != "on_open" {
		t.Errorf("menu dispatch: %v", out)
	}

	// Every event-kind name parses; keyup/mousedrag/mouseup for coverage.
	for _, k := range []string{"keyup", "mousedrag", "mouseup"} {
		if _, err := m.Dispatch(btn, map[string]any{"kind": k}); err != nil {
			t.Errorf("Dispatch kind %q: %v", k, err)
		}
	}
	// error branches.
	if _, err := m.Dispatch(999, map[string]any{"kind": "click"}); err == nil {
		t.Error("Dispatch bad handle")
	}
	if _, err := m.Dispatch(btn, map[string]any{"kind": "levitate"}); err == nil {
		t.Error("Dispatch bad kind")
	}
	if _, err := m.Dispatch(btn, nil); err == nil {
		t.Error("Dispatch nil ev (missing kind)")
	}
}

// --- reflective surface: Call / Methods / coerce -----------------------------

func TestCall(t *testing.T) {
	m := NewModule()

	// Returns a value (int handle).
	v, err := Call(m, "button", "OK", "cb")
	if err != nil {
		t.Fatal(err)
	}
	id := v.(int)

	// Returns (value, error) with nil error -> the value.
	got, err := Call(m, "text", id)
	if err != nil || got.(string) != "OK" {
		t.Fatalf("call text: %v %v", got, err)
	}

	// Returns error-only, nil -> nil result.
	if r, err := Call(m, "set_text", id, "Hi"); err != nil || r != nil {
		t.Fatalf("call set_text: %v %v", r, err)
	}

	// Returns error-only, non-nil -> surfaced.
	if _, err := Call(m, "set_text", 999, "x"); err == nil {
		t.Error("call set_text bad handle should error")
	}

	// A no-return method dispatched via Call (SetBounds -> error only, nil).
	if _, err := Call(m, "set_bounds", id, 0, 0, 10, 10); err != nil {
		t.Fatal(err)
	}

	// Assignable Array/Hash arguments flow straight through.
	if _, err := Call(m, "drop_down", []any{"a", "b"}, 0, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Call(m, "dispatch", id, map[string]any{"kind": "click"}); err != nil {
		t.Fatal(err)
	}

	// Numeric coercion: a float64/int64 handle coerces to the int parameter.
	if _, err := Call(m, "bounds", float64(id)); err != nil {
		t.Fatalf("float handle: %v", err)
	}
	if _, err := Call(m, "bounds", int64(id)); err != nil {
		t.Fatalf("int64 handle: %v", err)
	}

	// Omitted trailing args default (a bare label -> Label("")).
	if _, err := Call(m, "label"); err != nil {
		t.Fatal(err)
	}

	// A non-string value coerces to a String parameter via fmt.Sprint.
	if _, err := Call(m, "label", 42); err != nil {
		t.Fatalf("int->string label: %v", err)
	}

	// A non-bool value coerces to a Bool parameter via truthiness.
	if _, err := Call(m, "check_button", "c", 1, ""); err != nil {
		t.Fatalf("int->bool check_button: %v", err)
	}

	// An omitted Hash argument coerces to a nil map (add's opts left off).
	cc, _ := m.Container("vbox")
	if _, err := Call(m, "add", cc, m.Label("x")); err != nil {
		t.Fatalf("add with omitted opts: %v", err)
	}

	// A leading-underscore name exercises camelize's empty-segment skip.
	if _, err := Call(m, "_border"); err != nil {
		t.Fatalf("_border: %v", err)
	}

	// error surfaces.
	if _, err := Call(nil, "x"); err == nil {
		t.Error("nil receiver")
	}
	if _, err := Call(m, "no_such_method"); err == nil {
		t.Error("unknown method")
	}
	if _, err := Call(m, "label", "a", "b"); err == nil {
		t.Error("too many args")
	}
	if _, err := Call(m, "select", "not-an-int", 0); err == nil {
		t.Error("int coercion failure")
	}
	if _, err := Call(m, "drop_down", 5, 0, ""); err == nil {
		t.Error("non-array where Array expected")
	}
	if _, err := Call(m, "add", 1, 2, 5); err == nil {
		t.Error("non-hash where Hash expected")
	}
}

func TestMethods(t *testing.T) {
	m := NewModule()
	names := Methods(m)
	if !sortedContains(names, "add_widget") || !sortedContains(names, "drop_down") ||
		!sortedContains(names, "set_text") || !sortedContains(names, "menu_bar") {
		t.Errorf("expected snake_case names missing from %v", names)
	}
	// Sorted + non-empty.
	if len(names) == 0 || !reflect.DeepEqual(names, sortedCopy(names)) {
		t.Error("Methods should be sorted and non-empty")
	}
}

// --- low-level coercions (branches Call cannot reach on its own) -------------

func TestCoerceAndHelpers(t *testing.T) {
	// coerce's default branch: a parameter kind none of the methods use.
	if _, err := coerce(nil, reflect.TypeOf(float64(0))); err == nil {
		t.Error("coerce to float should be unsupported")
	}
	// String from nil -> "".
	v, err := coerce(nil, reflect.TypeOf(""))
	if err != nil || v.String() != "" {
		t.Errorf("coerce nil->string: %v %v", v, err)
	}
	// Slice(non-byte) from nil -> zero slice; from a bad value -> error.
	sl, err := coerce(nil, reflect.TypeOf([]any(nil)))
	if err != nil || !sl.IsNil() {
		t.Errorf("coerce nil->slice: %v %v", sl, err)
	}
	if _, err := coerce(5, reflect.TypeOf([]any(nil))); err == nil {
		t.Error("coerce int->slice should error")
	}
	// Map from a bad value -> error (nil path is covered via Call add).
	if _, err := coerce(5, reflect.TypeOf(map[string]any(nil))); err == nil {
		t.Error("coerce int->map should error")
	}
	// toInt over every numeric kind + failure.
	if n, ok := toInt(int64(7)); !ok || n != 7 {
		t.Error("toInt int64")
	}
	if n, ok := toInt(3.9); !ok || n != 3 {
		t.Error("toInt float64")
	}
	if _, ok := toInt("x"); ok {
		t.Error("toInt string should fail")
	}
	// truthy: nil, bool, and a non-bool value.
	if truthy(nil) || !truthy(true) || truthy(false) || !truthy(7) {
		t.Error("truthy semantics wrong")
	}
	// toStrings over a scalar element.
	if got := toStrings([]any{1, "b"}); got[0] != "1" || got[1] != "b" {
		t.Errorf("toStrings: %v", got)
	}
}

// --- small test helpers ------------------------------------------------------

func sortedContains(s []string, want string) bool {
	for _, v := range s {
		if v == want {
			return true
		}
	}
	return false
}

func sortedCopy(s []string) []string {
	out := append([]string(nil), s...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}
