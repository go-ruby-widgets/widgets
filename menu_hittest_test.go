// Copyright (c) 2026 the go-ruby-widgets/widgets authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package widgets

import (
	"testing"

	"github.com/go-widgets/toolkit"
)

// hitMenu builds the fixture menu shared by the hit-test cases: an enabled row,
// a separator, a second enabled row, and a disabled (action-less) row. Its
// geometry at the default theme (scale 1, compact density) is fixed:
//
//	row 0 enabled   top=2  h=22 -> band [2,24)
//	row 1 separator top=24 h=6  -> band [24,30)  (RowAt -> -1)
//	row 2 enabled   top=30 h=22 -> band [30,52)
//	row 3 disabled  top=52 h=22 -> band [52,74)  (RowAt -> -1)
func hitMenu(m *Module) (int, *toolkit.Menu) {
	id := m.Menu([]any{
		map[string]any{"label": "Open", "action": "on_open"},
		map[string]any{"separator": true},
		map[string]any{"label": "Save", "action": "on_save"},
		map[string]any{"label": "Disabled", "action": ""},
	})
	return id, m.objs[id].(*toolkit.Menu)
}

// TestMenuRowAtControlRun validates the adapter instrument against the
// known-good control (toolkit.Menu.RowAt) BEFORE the exact-index assertions
// lean on it: over a dense sweep of widget-local points on a laid-out menu, the
// Ruby-callable MenuRowAt must return EXACTLY what a direct toolkit RowAt does,
// so the two can never resolve a point to different rows.
func TestMenuRowAtControlRun(t *testing.T) {
	m := NewModule()
	id, mn := hitMenu(m)
	if err := m.Layout(id, 200, 200); err != nil {
		t.Fatalf("Layout: %v", err)
	}
	for _, x := range []int{-1, 0, 5, 199, 200, 500} {
		for y := -5; y <= 100; y++ {
			got, err := m.MenuRowAt(id, x, y)
			if err != nil {
				t.Fatalf("MenuRowAt(%d,%d): %v", x, y, err)
			}
			if want := mn.RowAt(x, y); got != want {
				t.Fatalf("MenuRowAt(%d,%d)=%d, toolkit RowAt=%d (control-run divergence)", x, y, got, want)
			}
		}
	}
}

// TestMenuRowAtExact pins the exact row index at exact coordinates: row tops,
// centres, last pixels, boundaries, a point on the separator, a point on the
// disabled row, and points outside the bounds on every side.
func TestMenuRowAtExact(t *testing.T) {
	m := NewModule()
	id, _ := hitMenu(m)
	if err := m.Layout(id, 200, 200); err != nil {
		t.Fatalf("Layout: %v", err)
	}
	cases := []struct {
		name string
		x, y int
		want int
	}{
		{"above top inset", 5, 1, -1},
		{"row0 top boundary", 5, 2, 0},
		{"row0 centre", 5, 13, 0},
		{"row0 last pixel", 5, 23, 0},
		{"separator top", 5, 24, -1},
		{"separator mid", 5, 27, -1},
		{"row2 top boundary", 5, 30, 2},
		{"row2 centre", 5, 40, 2},
		{"row2 last pixel", 5, 51, 2},
		{"disabled row top", 5, 52, -1},
		{"disabled row mid", 5, 60, -1},
		{"below last row", 5, 74, -1},
		{"x off left", -1, 13, -1},
		{"x off right", 500, 13, -1},
		{"y far below", 5, 999, -1},
	}
	for _, c := range cases {
		got, err := m.MenuRowAt(id, c.x, c.y)
		if err != nil {
			t.Fatalf("%s: MenuRowAt(%d,%d): %v", c.name, c.x, c.y, err)
		}
		if got != c.want {
			t.Errorf("%s: MenuRowAt(%d,%d)=%d, want %d", c.name, c.x, c.y, got, c.want)
		}
	}
}

// TestMenuRowAtUnlaidOut documents the contract: RowAt reads the menu's Bounds,
// so on a menu that was never laid out (zero bounds) every point is outside and
// the result is uniformly -1 — never a false row hit.
func TestMenuRowAtUnlaidOut(t *testing.T) {
	m := NewModule()
	id, _ := hitMenu(m)
	for _, p := range [][2]int{{5, 2}, {5, 13}, {5, 40}, {0, 0}} {
		got, err := m.MenuRowAt(id, p[0], p[1])
		if err != nil {
			t.Fatalf("MenuRowAt(%d,%d): %v", p[0], p[1], err)
		}
		if got != -1 {
			t.Errorf("unlaid MenuRowAt(%d,%d)=%d, want -1", p[0], p[1], got)
		}
	}
}

// TestMenuRowAtErrors exercises the error branches: an unknown handle, the 0
// "none" handle, and a handle that resolves to another widget.
func TestMenuRowAtErrors(t *testing.T) {
	m := NewModule()
	if _, err := m.MenuRowAt(999, 5, 5); err == nil {
		t.Error("MenuRowAt: unknown handle should error")
	}
	if _, err := m.MenuRowAt(0, 5, 5); err == nil {
		t.Error("MenuRowAt: zero handle should error")
	}
	if _, err := m.MenuRowAt(m.HBox(), 5, 5); err == nil {
		t.Error("MenuRowAt: non-menu handle should error")
	}
}

// TestMenuRowTop asserts exact row tops (with the scroll offset applied),
// panic-free out-of-range indices, and the error branches.
func TestMenuRowTop(t *testing.T) {
	m := NewModule()
	id, _ := hitMenu(m)
	if err := m.Layout(id, 200, 200); err != nil {
		t.Fatalf("Layout: %v", err)
	}
	for _, c := range []struct{ i, want int }{
		{0, 2}, {1, 24}, {2, 30}, {3, 52}, {4, 74}, {-1, 2},
	} {
		got, err := m.MenuRowTop(id, c.i)
		if err != nil {
			t.Fatalf("MenuRowTop(%d): %v", c.i, err)
		}
		if got != c.want {
			t.Errorf("MenuRowTop(%d)=%d, want %d", c.i, got, c.want)
		}
	}
	if _, err := m.MenuRowTop(999, 0); err == nil {
		t.Error("MenuRowTop: unknown handle should error")
	}
	if _, err := m.MenuRowTop(m.HBox(), 0); err == nil {
		t.Error("MenuRowTop: non-menu handle should error")
	}
}

// TestMenuRowHeight asserts exact row heights (normal, separator), the -1 for an
// out-of-range index, and the error branches.
func TestMenuRowHeight(t *testing.T) {
	m := NewModule()
	id, _ := hitMenu(m)
	if err := m.Layout(id, 200, 200); err != nil {
		t.Fatalf("Layout: %v", err)
	}
	for _, c := range []struct{ i, want int }{
		{0, 22}, {1, 6}, {2, 22}, {3, 22}, {-1, -1}, {4, -1}, {99, -1},
	} {
		got, err := m.MenuRowHeight(id, c.i)
		if err != nil {
			t.Fatalf("MenuRowHeight(%d): %v", c.i, err)
		}
		if got != c.want {
			t.Errorf("MenuRowHeight(%d)=%d, want %d", c.i, got, c.want)
		}
	}
	if _, err := m.MenuRowHeight(999, 0); err == nil {
		t.Error("MenuRowHeight: unknown handle should error")
	}
	if _, err := m.MenuRowHeight(m.HBox(), 0); err == nil {
		t.Error("MenuRowHeight: non-menu handle should error")
	}
}

// TestMenuHitTestRubySurface proves the methods reach Ruby through the same
// reflect-driven Call surface method_missing binds — snake_case names, integer
// args and results — exactly as menu/render/layout do.
func TestMenuHitTestRubySurface(t *testing.T) {
	m := NewModule()
	id, _ := hitMenu(m)
	if _, err := Call(m, "layout", id, 200, 200); err != nil {
		t.Fatalf("Call layout: %v", err)
	}
	got, err := Call(m, "menu_row_at", id, 5, 40)
	if err != nil {
		t.Fatalf("Call menu_row_at: %v", err)
	}
	if got != 2 {
		t.Errorf("Call menu_row_at = %v, want 2", got)
	}
	top, err := Call(m, "menu_row_top", id, 2)
	if err != nil || top != 30 {
		t.Errorf("Call menu_row_top = %v, %v; want 30", top, err)
	}
	h, err := Call(m, "menu_row_height", id, 1)
	if err != nil || h != 6 {
		t.Errorf("Call menu_row_height = %v, %v; want 6", h, err)
	}
	names := Methods(m)
	for _, want := range []string{"menu_row_at", "menu_row_top", "menu_row_height"} {
		if !sortedContains(names, want) {
			t.Errorf("Methods missing %q", want)
		}
	}
}
