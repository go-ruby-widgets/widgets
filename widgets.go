// Copyright (c) 2026 the go-ruby-widgets/widgets authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package widgets

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/go-widgets/painter"
	"github.com/go-widgets/toolkit"
)

// Module is the stateful Ruby receiver: the `Widgets` module under rbgo. Unlike
// the stateless data adapters (opentype, regexp, …) a Module owns a live widget
// tree — every constructed widget or container is stored under an integer
// handle the Ruby side references, and every mutation, layout, render and event
// dispatch is addressed by that handle. A Module is NOT safe for concurrent use
// (its handle table and pending-callback list are mutated in place).
type Module struct {
	objs  map[int]any
	next  int
	theme *toolkit.Theme
	fired []string
}

// NewModule returns a fresh Module with an empty handle table and the default
// light theme. Handles count up from 1, so a 0 handle always means "none"
// (used by Frame/Dock to mean a nil child/body).
func NewModule() *Module {
	return &Module{objs: map[int]any{}, next: 1, theme: toolkit.DefaultLight()}
}

// reg stores o under a fresh handle and returns it.
func (m *Module) reg(o any) int {
	id := m.next
	m.next++
	m.objs[id] = o
	return id
}

// get looks up the object behind a handle. A handle of 0 or an unknown id is an
// error.
func (m *Module) get(id int) (any, error) {
	o, ok := m.objs[id]
	if !ok {
		return nil, fmt.Errorf("widgets: unknown handle %d", id)
	}
	return o, nil
}

// getWidget looks up a handle as a toolkit.Widget. Every object a Module stores
// is a widget (every constructor registers a toolkit widget or container), so
// the only failure is an unknown handle, surfaced by get.
func (m *Module) getWidget(id int) (toolkit.Widget, error) {
	o, err := m.get(id)
	if err != nil {
		return nil, err
	}
	return o.(toolkit.Widget), nil
}

// fire records that a wired callback identifier ran; Dispatch drains the list.
func (m *Module) fire(cb string) { m.fired = append(m.fired, cb) }

// --- widget constructors (each returns an opaque integer handle) -------------

// Button constructs a push button labelled label. When callback is non-empty it
// is fired (reported by Dispatch) on every click.
func (m *Module) Button(label, callback string) int {
	b := toolkit.NewButton(label, nil)
	if callback != "" {
		b.OnClick = func() { m.fire(callback) }
	}
	return m.reg(b)
}

// Label constructs a passive text label.
func (m *Module) Label(text string) int { return m.reg(toolkit.NewLabel(text)) }

// Entry constructs a single-line text field seeded with initial. When callback
// is non-empty it fires on every value change and on submit (Enter).
func (m *Module) Entry(initial, callback string) int {
	e := toolkit.NewEntry(initial)
	if callback != "" {
		e.OnChange = func(string) { m.fire(callback) }
		e.OnSubmit = func(string) { m.fire(callback) }
	}
	return m.reg(e)
}

// TextView constructs a multi-line editable text area seeded with initial.
func (m *Module) TextView(initial string) int {
	return m.reg(toolkit.NewTextView(initial))
}

// CheckButton constructs a labelled checkbox. When callback is non-empty it
// fires on every toggle.
func (m *Module) CheckButton(label string, checked bool, callback string) int {
	c := toolkit.NewCheckButton(label, checked)
	if callback != "" {
		c.OnToggle = func(bool) { m.fire(callback) }
	}
	return m.reg(c)
}

// DropDown constructs a drop-down selector over options (a Ruby Array of
// strings) with the given initial selection. When callback is non-empty it
// fires on every Select.
func (m *Module) DropDown(options []any, selected int, callback string) int {
	d := toolkit.NewDropDown(toStrings(options), selected)
	if callback != "" {
		d.OnSelect = func(int) { m.fire(callback) }
	}
	return m.reg(d)
}

// ListBox constructs a scrollable single-column list over items (a Ruby Array
// of strings). When callback is non-empty it fires on every row activation.
func (m *Module) ListBox(items []any, callback string) int {
	l := toolkit.NewListBox(toStrings(items))
	if callback != "" {
		l.OnActivate = func(int) { m.fire(callback) }
	}
	return m.reg(l)
}

// Menu constructs a vertical popover menu from items — a Ruby Array of Hashes,
// each with "label", optional "shortcut", a Bool "separator" and an "action"
// (a callback identifier fired when the row is chosen). Non-Hash elements are
// skipped.
func (m *Module) Menu(items []any) int {
	mis := make([]toolkit.MenuItem, 0, len(items))
	for _, raw := range items {
		h, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		it := toolkit.MenuItem{}
		if v, ok := h["label"]; ok {
			it.Label = fmt.Sprint(v)
		}
		if v, ok := h["shortcut"]; ok {
			it.Shortcut = fmt.Sprint(v)
		}
		if v, ok := h["separator"]; ok {
			it.Separator = truthy(v)
		}
		if v, ok := h["action"]; ok {
			if cb := fmt.Sprint(v); cb != "" {
				it.Action = func() { m.fire(cb) }
			}
		}
		mis = append(mis, it)
	}
	return m.reg(toolkit.NewMenu(mis))
}

// MenuBar constructs an empty horizontal menu bar; attach menus with AddMenu.
func (m *Module) MenuBar() int { return m.reg(toolkit.NewMenuBar()) }

// --- container constructors --------------------------------------------------

// Container constructs a config-driven container whose children are placed by a
// named layout: "fit" (fill), "box"/"hbox" (a row), "vbox" (a column),
// "border" (edge regions + centre) or "card" (one visible child). An unknown
// name is an error, reported by Call.
func (m *Module) Container(layout string) (int, error) {
	l, err := makeLayout(layout)
	if err != nil {
		return 0, err
	}
	return m.reg(toolkit.NewContainer(l)), nil
}

// HBox constructs an imperative horizontal box (children added with AddWidget /
// AddFixed / AddFlex).
func (m *Module) HBox() int { return m.reg(toolkit.NewHBox()) }

// VBox constructs an imperative vertical box.
func (m *Module) VBox() int { return m.reg(toolkit.NewVBox()) }

// Grid constructs a cols×rows table (children placed with Attach).
func (m *Module) Grid(cols, rows int) int { return m.reg(toolkit.NewGrid(cols, rows)) }

// Frame constructs a 1-pixel-bordered panel around child; pass a 0 child for an
// empty frame.
func (m *Module) Frame(child int) (int, error) {
	var c toolkit.Widget
	if child != 0 {
		w, err := m.getWidget(child)
		if err != nil {
			return 0, err
		}
		c = w
	}
	return m.reg(toolkit.NewFrame(c)), nil
}

// Dock constructs an edge-docking container around body (pass 0 for a bars-only
// frame); attach bars with DockAt.
func (m *Module) Dock(body int) (int, error) {
	var b toolkit.Widget
	if body != 0 {
		w, err := m.getWidget(body)
		if err != nil {
			return 0, err
		}
		b = w
	}
	return m.reg(toolkit.NewDock(b)), nil
}

// Border constructs an edge-region container (set regions with SetRegion).
func (m *Module) Border() int { return m.reg(toolkit.NewBorder()) }

// --- mutation ----------------------------------------------------------------

// SetText sets the text of a Label, Button, Entry, TextView or CheckButton.
func (m *Module) SetText(id int, s string) error {
	o, err := m.get(id)
	if err != nil {
		return err
	}
	switch w := o.(type) {
	case *toolkit.Label:
		w.Text = s
	case *toolkit.Button:
		w.Label = s
	case *toolkit.Entry:
		w.Text = s
		w.Cursor = len([]rune(s))
	case *toolkit.TextView:
		w.SetText(s)
	case *toolkit.CheckButton:
		w.Label = s
	default:
		return fmt.Errorf("widgets: SetText: handle %d (%T) has no text", id, o)
	}
	return nil
}

// Text reads the text of a Label, Button, Entry, TextView or CheckButton.
func (m *Module) Text(id int) (string, error) {
	o, err := m.get(id)
	if err != nil {
		return "", err
	}
	switch w := o.(type) {
	case *toolkit.Label:
		return w.Text, nil
	case *toolkit.Button:
		return w.Label, nil
	case *toolkit.Entry:
		return w.Text, nil
	case *toolkit.TextView:
		return w.Text(), nil
	case *toolkit.CheckButton:
		return w.Label, nil
	default:
		return "", fmt.Errorf("widgets: Text: handle %d (%T) has no text", id, o)
	}
}

// SetChecked sets a CheckButton's state.
func (m *Module) SetChecked(id int, v bool) error {
	o, err := m.get(id)
	if err != nil {
		return err
	}
	c, ok := o.(*toolkit.CheckButton)
	if !ok {
		return fmt.Errorf("widgets: SetChecked: handle %d (%T) is not a check button", id, o)
	}
	c.Checked = v
	return nil
}

// Checked reads a CheckButton's state.
func (m *Module) Checked(id int) (bool, error) {
	o, err := m.get(id)
	if err != nil {
		return false, err
	}
	c, ok := o.(*toolkit.CheckButton)
	if !ok {
		return false, fmt.Errorf("widgets: Checked: handle %d (%T) is not a check button", id, o)
	}
	return c.Checked, nil
}

// Select changes the selection of a DropDown (which also fires its callback) or
// a ListBox.
func (m *Module) Select(id, idx int) error {
	o, err := m.get(id)
	if err != nil {
		return err
	}
	switch w := o.(type) {
	case *toolkit.DropDown:
		w.Select(idx)
	case *toolkit.ListBox:
		w.Selected = idx
	default:
		return fmt.Errorf("widgets: Select: handle %d (%T) is not selectable", id, o)
	}
	return nil
}

// SetStyle sets a Button's resting appearance: "default", "prominent" or
// "secondary".
func (m *Module) SetStyle(id int, style string) error {
	o, err := m.get(id)
	if err != nil {
		return err
	}
	b, ok := o.(*toolkit.Button)
	if !ok {
		return fmt.Errorf("widgets: SetStyle: handle %d (%T) is not a button", id, o)
	}
	switch strings.ToLower(style) {
	case "prominent":
		b.Style = toolkit.ButtonProminent
	case "secondary":
		b.Style = toolkit.ButtonSecondary
	case "", "default":
		b.Style = toolkit.ButtonDefault
	default:
		return fmt.Errorf("widgets: SetStyle: unknown style %q", style)
	}
	return nil
}

// SetSpacing sets the inter-child gap of an HBox, VBox or Grid.
func (m *Module) SetSpacing(id, n int) error {
	o, err := m.get(id)
	if err != nil {
		return err
	}
	switch w := o.(type) {
	case *toolkit.HBox:
		w.Spacing = n
	case *toolkit.VBox:
		w.Spacing = n
	case *toolkit.Grid:
		w.Spacing = n
	default:
		return fmt.Errorf("widgets: SetSpacing: handle %d (%T) has no spacing", id, o)
	}
	return nil
}

// SetTheme switches the render theme to "light" or "dark".
func (m *Module) SetTheme(name string) error {
	switch strings.ToLower(name) {
	case "", "light":
		m.theme = toolkit.DefaultLight()
	case "dark":
		m.theme = toolkit.DefaultDark()
	default:
		return fmt.Errorf("widgets: SetTheme: unknown theme %q", name)
	}
	return nil
}

// --- composition -------------------------------------------------------------

// AddWidget appends child to a Container, HBox or VBox with the default per-item
// config (an equal share in a box, the centre of a border layout).
func (m *Module) AddWidget(parent, child int) error {
	p, err := m.get(parent)
	if err != nil {
		return err
	}
	c, err := m.getWidget(child)
	if err != nil {
		return err
	}
	switch pc := p.(type) {
	case *toolkit.Container:
		pc.AddWidget(c)
	case *toolkit.HBox:
		pc.Append(c)
	case *toolkit.VBox:
		pc.Append(c)
	default:
		return fmt.Errorf("widgets: AddWidget: handle %d (%T) is not a box/container", parent, p)
	}
	return nil
}

// Add appends child to a Container with an explicit item config Hash: "flex"
// (proportional weight), "size" (fixed main-axis extent) and "region"
// ("north"/"south"/"east"/"west"/"center" for a border layout). A nil Hash is
// the zero config.
func (m *Module) Add(parent, child int, opts map[string]any) error {
	p, err := m.get(parent)
	if err != nil {
		return err
	}
	c, err := m.getWidget(child)
	if err != nil {
		return err
	}
	pc, ok := p.(*toolkit.Container)
	if !ok {
		return fmt.Errorf("widgets: Add: handle %d (%T) is not a container", parent, p)
	}
	it := toolkit.Item{Widget: c}
	if v, ok := opts["flex"]; ok {
		it.Flex, _ = toInt(v)
	}
	if v, ok := opts["size"]; ok {
		it.Size, _ = toInt(v)
	}
	if v, ok := opts["region"]; ok {
		r, err := parseRegion(fmt.Sprint(v))
		if err != nil {
			return err
		}
		it.Region = r
	}
	pc.Add(it)
	return nil
}

// AddFixed appends child to an HBox or VBox with a fixed main-axis size.
func (m *Module) AddFixed(parent, child, size int) error {
	p, c, err := m.parentChild(parent, child)
	if err != nil {
		return err
	}
	switch pc := p.(type) {
	case *toolkit.HBox:
		pc.AddFixed(c, size)
	case *toolkit.VBox:
		pc.AddFixed(c, size)
	default:
		return fmt.Errorf("widgets: AddFixed: handle %d (%T) is not a box", parent, p)
	}
	return nil
}

// AddFlex appends child to an HBox or VBox with a proportional flex weight.
func (m *Module) AddFlex(parent, child, flex int) error {
	p, c, err := m.parentChild(parent, child)
	if err != nil {
		return err
	}
	switch pc := p.(type) {
	case *toolkit.HBox:
		pc.AddFlex(c, flex)
	case *toolkit.VBox:
		pc.AddFlex(c, flex)
	default:
		return fmt.Errorf("widgets: AddFlex: handle %d (%T) is not a box", parent, p)
	}
	return nil
}

// Attach places child at (col, row) in a Grid.
func (m *Module) Attach(parent, child, col, row int) error {
	p, c, err := m.parentChild(parent, child)
	if err != nil {
		return err
	}
	g, ok := p.(*toolkit.Grid)
	if !ok {
		return fmt.Errorf("widgets: Attach: handle %d (%T) is not a grid", parent, p)
	}
	g.Attach(c, col, row)
	return nil
}

// DockAt attaches child to an edge ("top"/"bottom"/"left"/"right") of a Dock,
// reserving size pixels along that edge's axis.
func (m *Module) DockAt(parent, child int, side string, size int) error {
	p, c, err := m.parentChild(parent, child)
	if err != nil {
		return err
	}
	d, ok := p.(*toolkit.Dock)
	if !ok {
		return fmt.Errorf("widgets: DockAt: handle %d (%T) is not a dock", parent, p)
	}
	s, err := parseSide(side)
	if err != nil {
		return err
	}
	d.Dock(c, s, size)
	d.SetBounds(d.Bounds())
	return nil
}

// SetRegion assigns child to a Border region ("north"/"south"/"east"/"west"/
// "center"), with size the edge band's thickness (ignored for the centre).
func (m *Module) SetRegion(parent, child int, region string, size int) error {
	p, c, err := m.parentChild(parent, child)
	if err != nil {
		return err
	}
	b, ok := p.(*toolkit.Border)
	if !ok {
		return fmt.Errorf("widgets: SetRegion: handle %d (%T) is not a border", parent, p)
	}
	switch strings.ToLower(region) {
	case "north":
		b.North, b.NorthSize = c, size
	case "south":
		b.South, b.SouthSize = c, size
	case "east":
		b.East, b.EastSize = c, size
	case "west":
		b.West, b.WestSize = c, size
	case "", "center", "centre":
		b.Center = c
	default:
		return fmt.Errorf("widgets: SetRegion: unknown region %q", region)
	}
	b.SetBounds(b.Bounds())
	return nil
}

// AddMenu appends a named menu to a MenuBar.
func (m *Module) AddMenu(bar int, name string, menu int) error {
	p, err := m.get(bar)
	if err != nil {
		return err
	}
	mb, ok := p.(*toolkit.MenuBar)
	if !ok {
		return fmt.Errorf("widgets: AddMenu: handle %d (%T) is not a menu bar", bar, p)
	}
	mo, err := m.get(menu)
	if err != nil {
		return err
	}
	mn, ok := mo.(*toolkit.Menu)
	if !ok {
		return fmt.Errorf("widgets: AddMenu: handle %d (%T) is not a menu", menu, mo)
	}
	mb.AddMenu(name, mn)
	return nil
}

// SetActive selects the visible child of a Container backed by a card layout.
func (m *Module) SetActive(container, idx int) error {
	o, err := m.get(container)
	if err != nil {
		return err
	}
	c, ok := o.(*toolkit.Container)
	if !ok {
		return fmt.Errorf("widgets: SetActive: handle %d (%T) is not a container", container, o)
	}
	cl, ok := c.Layout.(*toolkit.CardLayout)
	if !ok {
		return fmt.Errorf("widgets: SetActive: container %d is not a card layout", container)
	}
	cl.Active = idx
	c.SetBounds(c.Bounds())
	return nil
}

// SetLayout swaps a Container's layout to a named one (see Container).
func (m *Module) SetLayout(container int, layout string) error {
	o, err := m.get(container)
	if err != nil {
		return err
	}
	c, ok := o.(*toolkit.Container)
	if !ok {
		return fmt.Errorf("widgets: SetLayout: handle %d (%T) is not a container", container, o)
	}
	l, err := makeLayout(layout)
	if err != nil {
		return err
	}
	c.Layout = l
	c.SetBounds(c.Bounds())
	return nil
}

// parentChild resolves a parent object and a child widget in one step, the
// common opening of the composition methods.
func (m *Module) parentChild(parent, child int) (any, toolkit.Widget, error) {
	p, err := m.get(parent)
	if err != nil {
		return nil, nil, err
	}
	c, err := m.getWidget(child)
	if err != nil {
		return nil, nil, err
	}
	return p, c, nil
}

// --- layout / query / render -------------------------------------------------

// SetBounds positions a widget (or the root of a tree) at (x, y) with size w×h,
// running the toolkit layout over any descendants.
func (m *Module) SetBounds(id, x, y, w, h int) error {
	wdg, err := m.getWidget(id)
	if err != nil {
		return err
	}
	wdg.SetBounds(toolkit.Rect{X: x, Y: y, W: w, H: h})
	return nil
}

// Layout is SetBounds at the origin — the common case for a top-level tree.
func (m *Module) Layout(id, w, h int) error { return m.SetBounds(id, 0, 0, w, h) }

// Bounds reports a widget's placement as a Hash with "x", "y", "w", "h".
func (m *Module) Bounds(id int) (map[string]any, error) {
	wdg, err := m.getWidget(id)
	if err != nil {
		return nil, err
	}
	r := wdg.Bounds()
	return map[string]any{"x": r.X, "y": r.Y, "w": r.W, "h": r.H}, nil
}

// Render lays a tree out to fill a w×h surface and paints it, returning a Hash
// with the RGBA "pixels" (4 bytes per pixel, row-major, top-left origin),
// "stride" (the byte offset between rows, i.e. w*4) and "w"/"h". A host
// (wasmbox) blits pixels straight into a canvas / SharedArrayBuffer.
func (m *Module) Render(id, w, h int) (map[string]any, error) {
	wdg, err := m.getWidget(id)
	if err != nil {
		return nil, err
	}
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("widgets: Render: size must be positive, got %dx%d", w, h)
	}
	buf := make([]byte, 4*w*h)
	p := painter.NewPixelPainter(buf, w, h)
	wdg.SetBounds(toolkit.Rect{X: 0, Y: 0, W: w, H: h})
	wdg.Draw(p, m.theme)
	return map[string]any{"pixels": buf, "stride": w * 4, "w": w, "h": h}, nil
}

// Dispatch routes an input event into a tree and reports the outcome. The ev
// Hash carries "kind" ("click"/"keydown"/"keyup"/"char"/"mousedrag"/"mouseup"),
// "x"/"y" (widget-local pixels), "code" (key/char text) and Bool "ctrl"/"shift".
// The result Hash is "fired" (an Array of the callback identifiers that ran,
// in order) and a Bool "repaint" (whether any callback ran, i.e. the tree may
// have changed).
func (m *Module) Dispatch(id int, ev map[string]any) (map[string]any, error) {
	wdg, err := m.getWidget(id)
	if err != nil {
		return nil, err
	}
	kind, err := parseEventKind(fmt.Sprint(ev["kind"]))
	if err != nil {
		return nil, err
	}
	e := toolkit.Event{Kind: kind}
	if v, ok := ev["x"]; ok {
		e.X, _ = toInt(v)
	}
	if v, ok := ev["y"]; ok {
		e.Y, _ = toInt(v)
	}
	if v, ok := ev["code"]; ok {
		e.Code = fmt.Sprint(v)
	}
	if v, ok := ev["ctrl"]; ok {
		e.Ctrl = truthy(v)
	}
	if v, ok := ev["shift"]; ok {
		e.Shift = truthy(v)
	}
	m.fired = nil
	wdg.OnEvent(e)
	fired := make([]any, len(m.fired))
	for i, f := range m.fired {
		fired[i] = f
	}
	return map[string]any{"fired": fired, "repaint": len(m.fired) > 0}, nil
}

// --- small parsers / coercions -----------------------------------------------

// makeLayout maps a layout name to a fresh toolkit.Layout.
func makeLayout(name string) (toolkit.Layout, error) {
	switch strings.ToLower(name) {
	case "", "fit":
		return toolkit.FitLayout{}, nil
	case "box", "hbox":
		return &toolkit.BoxLayout{}, nil
	case "vbox":
		return &toolkit.BoxLayout{Vertical: true}, nil
	case "border":
		return toolkit.BorderLayout{}, nil
	case "card":
		return &toolkit.CardLayout{}, nil
	default:
		return nil, fmt.Errorf("widgets: unknown layout %q", name)
	}
}

// parseRegion maps a region name to a toolkit.Region.
func parseRegion(name string) (toolkit.Region, error) {
	switch strings.ToLower(name) {
	case "", "center", "centre":
		return toolkit.RegionCenter, nil
	case "north":
		return toolkit.RegionNorth, nil
	case "south":
		return toolkit.RegionSouth, nil
	case "west":
		return toolkit.RegionWest, nil
	case "east":
		return toolkit.RegionEast, nil
	default:
		return 0, fmt.Errorf("widgets: unknown region %q", name)
	}
}

// parseSide maps a dock-edge name to a toolkit.DockSide.
func parseSide(name string) (toolkit.DockSide, error) {
	switch strings.ToLower(name) {
	case "top":
		return toolkit.DockTop, nil
	case "bottom":
		return toolkit.DockBottom, nil
	case "left":
		return toolkit.DockLeft, nil
	case "right":
		return toolkit.DockRight, nil
	default:
		return 0, fmt.Errorf("widgets: unknown side %q", name)
	}
}

// parseEventKind maps an event-kind name to a toolkit.EventKind.
func parseEventKind(name string) (toolkit.EventKind, error) {
	switch strings.ToLower(name) {
	case "click":
		return toolkit.EventClick, nil
	case "keydown":
		return toolkit.EventKeyDown, nil
	case "keyup":
		return toolkit.EventKeyUp, nil
	case "char":
		return toolkit.EventChar, nil
	case "mousedrag":
		return toolkit.EventMouseDrag, nil
	case "mouseup":
		return toolkit.EventMouseUp, nil
	default:
		return 0, fmt.Errorf("widgets: unknown event kind %q", name)
	}
}

// toStrings coerces a Ruby Array (or a single scalar) into a []string.
func toStrings(v []any) []string {
	out := make([]string, len(v))
	for i, e := range v {
		out[i] = fmt.Sprint(e)
	}
	return out
}

// --- dynamic dispatch (the uniform surface rbgo binds via method_missing) ----

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// Call dispatches a Ruby-style snake_case method name to the matching exported
// method of the Module, coercing each Ruby-supplied argument to the Go
// parameter type. Trailing arguments may be omitted (they default to nil/zero).
// The result is the method's Ruby-shaped return value (or nil for a method that
// returns nothing); a trailing error return is unwrapped into Call's own error.
// This is the single entry point an rbgo binding drives from method_missing.
func Call(recv any, method string, args ...any) (any, error) {
	rv := reflect.ValueOf(recv)
	if !rv.IsValid() {
		return nil, fmt.Errorf("widgets: nil receiver")
	}
	fn := rv.MethodByName(camelize(method))
	if !fn.IsValid() {
		return nil, fmt.Errorf("widgets: unknown method %q", method)
	}
	ft := fn.Type()
	if len(args) > ft.NumIn() {
		return nil, fmt.Errorf("widgets: %s takes at most %d argument(s), got %d",
			method, ft.NumIn(), len(args))
	}
	in := make([]reflect.Value, ft.NumIn())
	for i := 0; i < ft.NumIn(); i++ {
		var arg any
		if i < len(args) {
			arg = args[i]
		}
		v, err := coerce(arg, ft.In(i))
		if err != nil {
			return nil, fmt.Errorf("widgets: %s argument %d: %w", method, i+1, err)
		}
		in[i] = v
	}
	outs := fn.Call(in)
	if n := len(outs); n > 0 && ft.Out(n-1) == errorType {
		if e := outs[n-1]; !e.IsNil() {
			return nil, e.Interface().(error)
		}
		outs = outs[:n-1]
	}
	if len(outs) == 0 {
		return nil, nil
	}
	return outs[0].Interface(), nil
}

// Methods lists, sorted, the Ruby-style snake_case names Call accepts for recv.
func Methods(recv any) []string {
	t := reflect.TypeOf(recv)
	names := make([]string, 0, t.NumMethod())
	for i := 0; i < t.NumMethod(); i++ {
		names = append(names, snakeize(t.Method(i).Name))
	}
	sort.Strings(names)
	return names
}

// coerce converts a Ruby-supplied argument into the Go type target expects.
func coerce(val any, target reflect.Type) (reflect.Value, error) {
	if val != nil {
		if vv := reflect.ValueOf(val); vv.Type().AssignableTo(target) {
			return vv, nil
		}
	}
	switch target.Kind() {
	case reflect.Int, reflect.Int64:
		n, ok := toInt(val)
		if !ok {
			return reflect.Value{}, fmt.Errorf("cannot use %T as an integer", val)
		}
		return reflect.ValueOf(n).Convert(target), nil
	case reflect.String:
		if val == nil {
			return reflect.ValueOf("").Convert(target), nil
		}
		return reflect.ValueOf(fmt.Sprint(val)).Convert(target), nil
	case reflect.Bool:
		return reflect.ValueOf(truthy(val)), nil
	case reflect.Slice:
		if val == nil { // an omitted Array becomes a nil slice
			return reflect.Zero(target), nil
		}
		return reflect.Value{}, fmt.Errorf("cannot use %T as %s", val, target)
	case reflect.Map:
		if val == nil { // an omitted Hash becomes a nil map
			return reflect.Zero(target), nil
		}
		return reflect.Value{}, fmt.Errorf("cannot use %T as %s", val, target)
	default:
		return reflect.Value{}, fmt.Errorf("cannot use %T as %s", val, target)
	}
}

// toInt coerces a Ruby number to an int.
func toInt(val any) (int, bool) {
	switch x := val.(type) {
	case int:
		return x, true
	case int64:
		return int(x), true
	case float64:
		return int(x), true
	default:
		return 0, false
	}
}

// truthy applies Ruby truthiness: only false and nil are falsey.
func truthy(val any) bool {
	switch x := val.(type) {
	case nil:
		return false
	case bool:
		return x
	default:
		return true
	}
}

// camelize turns a snake_case method name into its exported Go method name.
func camelize(s string) string {
	var b strings.Builder
	for _, w := range strings.Split(s, "_") {
		if w == "" {
			continue
		}
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		b.WriteString(string(r))
	}
	return b.String()
}

// snakeize is the inverse: an exported Go method name to snake_case.
func snakeize(s string) string {
	var b []rune
	rs := []rune(s)
	for i, r := range rs {
		if unicode.IsUpper(r) {
			prevLower := i > 0 && unicode.IsLower(rs[i-1])
			nextLower := i+1 < len(rs) && unicode.IsLower(rs[i+1])
			if i > 0 && (prevLower || nextLower) {
				b = append(b, '_')
			}
			r = unicode.ToLower(r)
		}
		b = append(b, r)
	}
	return string(b)
}
