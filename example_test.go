// Copyright (c) 2026 the go-ruby-widgets/widgets authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package widgets_test

import (
	"fmt"

	"github.com/go-ruby-widgets/widgets"
)

// Example builds a two-widget column, lays it out, renders it to a pixel buffer
// and routes a click into the button — every result a Ruby-shaped value.
func Example() {
	m := widgets.NewModule()

	root := m.VBox()
	title := m.Label("Hello")
	ok := m.Button("OK", "on_ok")
	_ = m.AddWidget(root, title)
	_ = m.AddWidget(root, ok)
	_ = m.Layout(root, 200, 80)

	img, _ := m.Render(root, 200, 80)
	fmt.Println("stride:", img["stride"], "w:", img["w"], "h:", img["h"])

	out, _ := m.Dispatch(ok, map[string]any{"kind": "click"})
	fmt.Println("fired:", out["fired"], "repaint:", out["repaint"])

	// Output:
	// stride: 800 w: 200 h: 80
	// fired: [on_ok] repaint: true
}
