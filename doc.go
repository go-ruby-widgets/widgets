// Copyright (c) 2026 the go-ruby-widgets/widgets authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

// Package widgets is the pure-Go, Ruby-runtime-independent core of the Ruby
// `widgets` gem: a live widget UI toolkit — buttons, labels, text fields,
// lists, menus and the container/layout system that arranges them — shaped so
// that github.com/go-embedded-ruby/ruby (rbgo) can bind it as
// `require "widgets"` and build real pixel UIs.
//
// It is a thin adapter over github.com/go-widgets/toolkit (the widget set and
// its container/layout model) and github.com/go-widgets/painter (the pixel
// rasteriser). Unlike the stateless data adapters of the go-ruby-* family
// (opentype, regexp, erb, …), a Module here owns a LIVE object graph: every
// widget and container is stored under an integer handle the Ruby side keeps,
// and every operation — mutate, compose, lay out, render, dispatch an event —
// is addressed by that handle and returns a Ruby-shaped value: a Hash
// (map[string]any), an Array ([]any) or a scalar. A single dynamic entry point,
// Call, dispatches a Ruby-style snake_case method name to the matching Module
// method and coerces the arguments, which is exactly what an rbgo binding drives
// from method_missing. Nothing here imports the Ruby runtime, so the package is
// equally usable as a standalone Go library.
//
// # The object graph
//
//   - Constructors return an opaque integer handle: Button, Label, Entry,
//     TextView, CheckButton, DropDown, ListBox, Menu, MenuBar for leaves;
//     Container (a config-driven fit/box/border/card layout), HBox, VBox, Grid,
//     Frame, Dock and Border for containers; and the overlay + chrome set a
//     compositor needs — Notification, Toast, Badge, Image, ContextMenu,
//     Popover, CommandPalette, IconButton, Tooltip, Avatar and LevelBar.
//   - Mutators address a handle: SetText/Text, SetChecked/Checked, Select,
//     SetStyle, SetSpacing, the package-wide SetTheme, and the overlay state
//     SetVisible/Visible, Popup, AnchorIn, SetLife, Tick, SetKind and SetValue.
//   - Composition wires the tree: AddWidget, Add (with a flex/size/region
//     Hash), AddFixed, AddFlex, Attach (grid), DockAt, SetRegion (border),
//     AddMenu, SetActive (card) and SetLayout.
//   - Layout + query: SetBounds, Layout (at the origin) and Bounds.
//   - Render paints a tree into an RGBA pixel buffer; Dispatch routes an input
//     event into it.
//
// # The render seam
//
// Render(root, w, h) lays the tree out to fill a w×h surface and paints it,
// returning {"pixels": <RGBA bytes>, "stride": w*4, "w": w, "h": h}. The pixels
// are 4 bytes per pixel, row-major, top-left origin — exactly what a host
// (wasmbox) blits into a canvas or SharedArrayBuffer.
//
// # The event seam
//
// Dispatch(root, {"kind" => "click", "x" => …, "y" => …}) routes the event into
// the tree by hit-testing container bounds, then reports
// {"fired": [callback ids…], "repaint": bool}. A widget is wired to a callback
// by passing a callback identifier to its constructor (Button, Entry,
// CheckButton, DropDown, ListBox and Menu items); when it fires, its identifier
// appears in the "fired" Array so the Ruby side can invoke the matching block.
//
// # Usage from Go
//
//	m := widgets.NewModule()
//	root := m.VBox()
//	title := m.Label("Hello")
//	ok := m.Button("OK", "on_ok")
//	_ = m.AddWidget(root, title)
//	_ = m.AddWidget(root, ok)
//	_ = m.Layout(root, 200, 80)
//	img, _ := m.Render(root, 200, 80) // {"pixels":…, "stride":800, "w":200, "h":80}
//	out, _ := m.Dispatch(ok, map[string]any{"kind": "click"})
//	// out["fired"] == []any{"on_ok"}
//
// # Usage from Ruby
//
// Under rbgo, `require "widgets"` gives a Widgets module whose snake_case
// methods are these operations, returning Ruby Hashes, Arrays and scalars:
//
//	require "widgets"
//
//	root  = Widgets.v_box
//	ok    = Widgets.button("OK", "on_ok")
//	Widgets.add_widget(root, ok)
//	Widgets.layout(root, 200, 80)
//	img   = Widgets.render(root, 200, 80)   # => { "pixels" => …, "stride" => 800, … }
//	fired = Widgets.dispatch(ok, { "kind" => "click" })   # => { "fired" => ["on_ok"], … }
//
// The `require "widgets"` binding lives in rbgo (a thin method_missing shim over
// Call); it is pending in that repo.
package widgets
