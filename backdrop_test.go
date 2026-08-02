// Copyright (c) 2026 the go-ruby-widgets/widgets authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package widgets

import (
	"testing"

	"github.com/go-widgets/painter"
	"github.com/go-widgets/toolkit"
)

func TestBackdropConstruction(t *testing.T) {
	m := NewModule()
	id, err := m.Backdrop("#11131a", "#171a24", 40)
	if err != nil {
		t.Fatalf("Backdrop: %v", err)
	}
	bd, ok := m.objs[id].(*toolkit.Backdrop)
	if !ok {
		t.Fatalf("handle %d is %T, want *toolkit.Backdrop", id, m.objs[id])
	}
	if bd.Fill != (painter.RGBA{R: 0x11, G: 0x13, B: 0x1a, A: 0xFF}) {
		t.Errorf("Fill = %v", bd.Fill)
	}
	if bd.Grid != (painter.RGBA{R: 0x17, G: 0x1a, B: 0x24, A: 0xFF}) {
		t.Errorf("Grid = %v", bd.Grid)
	}
	if bd.Step != 40 {
		t.Errorf("Step = %d, want 40", bd.Step)
	}

	// It renders to an RGBA buffer through the normal render path.
	img, err := m.Render(id, 80, 60)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if got := len(img["pixels"].([]byte)); got != 80*60*4 {
		t.Errorf("pixels len = %d, want %d", got, 80*60*4)
	}
}

func TestBackdropThemeDefaults(t *testing.T) {
	// Empty colour strings leave the zero RGBA, the "use theme" sentinel.
	m := NewModule()
	id, err := m.Backdrop("", "", 0)
	if err != nil {
		t.Fatalf("Backdrop: %v", err)
	}
	bd := m.objs[id].(*toolkit.Backdrop)
	if bd.Fill != (painter.RGBA{}) || bd.Grid != (painter.RGBA{}) {
		t.Errorf("empty strings should yield zero RGBA, got fill=%v grid=%v", bd.Fill, bd.Grid)
	}
}

func TestBackdropBadColours(t *testing.T) {
	m := NewModule()
	if _, err := m.Backdrop("nope", "#171a24", 40); err == nil {
		t.Error("bad fill should error")
	}
	if _, err := m.Backdrop("#11131a", "nope", 40); err == nil {
		t.Error("bad grid should error")
	}
}

func TestBackdropViaCall(t *testing.T) {
	// The Ruby-facing dispatch path: backdrop is enumerated by Methods and
	// invoked by name through Call.
	m := NewModule()
	if !contains(Methods(m), "backdrop") {
		t.Fatal("Methods should list backdrop")
	}
	res, err := Call(m, "backdrop", "#11131a", "#171a24", 40)
	if err != nil {
		t.Fatalf("Call backdrop: %v", err)
	}
	if _, ok := res.(int); !ok {
		t.Fatalf("backdrop returned %T, want int handle", res)
	}
	// A malformed colour surfaces as a Call error (unwrapped trailing error).
	if _, err := Call(m, "backdrop", "zzzzzz", "", 10); err == nil {
		t.Error("Call backdrop with bad colour should error")
	}
}

func TestParseHexColor(t *testing.T) {
	cases := []struct {
		in      string
		want    painter.RGBA
		wantErr bool
	}{
		{"#11131a", painter.RGBA{R: 0x11, G: 0x13, B: 0x1a, A: 0xFF}, false},
		{"171a24", painter.RGBA{R: 0x17, G: 0x1a, B: 0x24, A: 0xFF}, false}, // no '#'
		{"#8090a0c0", painter.RGBA{R: 0x80, G: 0x90, B: 0xa0, A: 0xc0}, false},
		{"  #11131a  ", painter.RGBA{R: 0x11, G: 0x13, B: 0x1a, A: 0xFF}, false}, // trimmed
		{"", painter.RGBA{}, false},      // empty -> theme sentinel
		{"#123", painter.RGBA{}, true},   // wrong length
		{"gggggg", painter.RGBA{}, true}, // right length, non-hex digit
	}
	for _, c := range cases {
		got, err := parseHexColor(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("parseHexColor(%q) err = %v, wantErr %v", c.in, err, c.wantErr)
			continue
		}
		if !c.wantErr && got != c.want {
			t.Errorf("parseHexColor(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// contains reports whether s is in xs.
func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
