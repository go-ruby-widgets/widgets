# go-ruby-widgets

[![CI](https://github.com/go-ruby-widgets/widgets/actions/workflows/ci.yml/badge.svg)](https://github.com/go-ruby-widgets/widgets/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/go-ruby-widgets/widgets.svg)](https://pkg.go.dev/github.com/go-ruby-widgets/widgets)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-ruby-widgets/widgets)](https://goreportcard.com/report/github.com/go-ruby-widgets/widgets)

The pure-Go, Ruby-runtime-independent core of the Ruby **`widgets`** gem — a live
widget UI toolkit (buttons, labels, text fields, lists, menus and the
container/layout system that arranges them) — shaped so that
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby) (`rbgo`) can bind it
as `require "widgets"` and build real pixel UIs.

It is a thin adapter over the [go-widgets](https://github.com/go-widgets) stack:

| Library | Role |
| --- | --- |
| [`go-widgets/toolkit`](https://github.com/go-widgets/toolkit) | The pure-Go widget set + container/layout model. |
| [`go-widgets/painter`](https://github.com/go-widgets/painter) | The pixel rasteriser (RGBA buffer back-end). |

Unlike the stateless data adapters of the `go-ruby-*` family (opentype, regexp,
erb, …), a `Module` here owns a **live object graph**: every widget and container
is stored under an integer **handle** the Ruby side keeps, and every operation —
mutate, compose, lay out, render, dispatch an event — is addressed by that handle
and returns a Ruby-shaped value (a Hash `map[string]any`, an Array `[]any` or a
scalar). Nothing here imports the Ruby runtime, so it is equally usable as a
standalone Go library.

`CGO_ENABLED=0`, no display, no network — deterministic and cross-compiles to all
six 64-bit Go architectures **and `js/wasm`** (the target [wasmdesk](https://github.com/wasmdesk) runs).

## The Ruby-facing surface

### Constructors (return an opaque integer handle)

| Kind | Methods |
| --- | --- |
| Leaves | `button(label, cb)`, `label(text)`, `entry(initial, cb)`, `text_view(initial)`, `check_button(label, checked, cb)`, `drop_down(options, selected, cb)`, `list_box(items, cb)`, `menu(items)`, `menu_bar` |
| Containers | `container(layout)` (`fit`/`box`/`hbox`/`vbox`/`border`/`card`), `h_box`, `v_box`, `grid(cols, rows)`, `frame(child)`, `dock(body)`, `border`, `backdrop(fill, grid, step)` |
| Overlays & chrome | `notification(text)`, `toast(text, kind, action_label, action)`, `badge(text, fill, ink)`, `image(pixels, w, h, scale)`, `context_menu(menu)`, `popover(child, title)`, `command_palette(commands)`, `icon_button(icon, cb)`, `tooltip(text, placement)`, `avatar(initials, color)`, `level_bar(max)`, `decoration(...)` |
| Desktop | `status_area`, `status_icon(icon, tooltip, on_click, on_right_click)`, `status_icon_image(pixels, w, h, tooltip, on_click, on_right_click)`, `wallpaper(pixels, w, h, mode)` (`fill`/`fit`/`center`/`tile`), `wallpaper_gradient(top_hex, bottom_hex)`, `thumbnail(pixels, w, h, label, on_click)` |

### Mutation

`set_text` / `text`, `set_checked` / `checked`, `select(id, idx)`, `set_style`,
`set_spacing`, and the module-wide `set_theme("light"|"dark")`. Overlay state:
`set_visible` / `visible`, `popup`, `anchor_in`, `set_life`, `tick`, `set_kind`,
`set_value`. Thumbnail state: `set_selected(id, bool)`, `set_hover(id, bool)`.

The module-wide `use_opentype_text` (and `use_opentype_text_size(px)`) upgrades
the toolkit's active font from the built-in 5x7 bitmap to anti-aliased, shaped
OpenType text (the bundled Atkinson Hyperlegible face). Call it once before the
first render and every widget — window titles, menus, HUD, desktop, frame
decorations — repaints against the vector face.

### Composition

`add_widget(parent, child)` (also joins a `status_icon` to a `status_area` and a
`badge` to a `status_icon`), `add(parent, child, {flex:, size:, region:})`,
`add_fixed`, `add_flex`, `attach(grid, child, col, row)`,
`dock_at(dock, child, side, size)`,
`set_region(border, child, region, size)`, `add_menu(bar, name, menu)`,
`set_active(card, idx)`, `set_layout(container, layout)`.

### Layout / query

`set_bounds(id, x, y, w, h)`, `layout(id, w, h)` (at the origin), `bounds(id)`.

### The render seam

```ruby
img = Widgets.render(root, w, h)
# => { "pixels" => <RGBA bytes>, "stride" => w*4, "w" => w, "h" => h }
```

`pixels` is 4 bytes per pixel, row-major, top-left origin — exactly what a host
(wasmbox) blits into a `<canvas>` / `SharedArrayBuffer`.

### The event seam

```ruby
out = Widgets.dispatch(root, { "kind" => "click", "x" => 10, "y" => 4 })
# => { "fired" => ["on_ok"], "repaint" => true }
```

`kind` is one of `click` / `keydown` / `keyup` / `char` / `mousedrag` /
`mouseup`. A widget is wired to a callback by passing an identifier to its
constructor; when it fires, that identifier appears in `fired` so the Ruby side
can invoke the matching block.

### Reflective dispatch

`Call(recv, method, args...) (any, error)` dispatches a snake_case method name to
the matching `Module` method, coercing Ruby scalars / Arrays / Hashes to the Go
parameter types (a trailing `error` return is unwrapped) — the single entry point
an rbgo `method_missing` shim drives. `Methods(recv)` lists the accepted names.

## Usage from Go

```go
m := widgets.NewModule()
root := m.VBox()
title := m.Label("Hello")
ok := m.Button("OK", "on_ok")
_ = m.AddWidget(root, title)
_ = m.AddWidget(root, ok)
_ = m.Layout(root, 200, 80)

img, _ := m.Render(root, 200, 80)                       // {"pixels":…, "stride":800, …}
out, _ := m.Dispatch(ok, map[string]any{"kind": "click"}) // out["fired"] == []any{"on_ok"}
```

## Usage from Ruby

```ruby
require "widgets"

root  = Widgets.v_box
ok    = Widgets.button("OK", "on_ok")
Widgets.add_widget(root, ok)
Widgets.layout(root, 200, 80)
img   = Widgets.render(root, 200, 80)
fired = Widgets.dispatch(ok, { "kind" => "click" })   # => { "fired" => ["on_ok"], … }
```

The `require "widgets"` binding lives in `rbgo` (a thin `method_missing` shim over
`Call`); it is pending in that repo.

## License

BSD-3-Clause. Copyright (c) 2026, the go-ruby-widgets/widgets authors.
