package flowgraph

// PropList interface for ordered, deletable property lists -- shared by a
// native Go-backed implementation and one wrapping an external owner (e.g. C++).
type PropList interface {

	// Get returns a key's value, and whether it was found
	Get(key string) (any, bool)

	// Set adds or updates a key's value, appending new keys in order
	Set(key string, v any) PropList

	// Delete removes a key, returning its value and whether it was found
	Delete(key string) (any, bool)

	// Has returns true if key is present
	Has(key string) bool

	// Len returns the number of entries
	Len() int

	// Keys returns all keys in insertion order
	Keys() []string

	// Touch moves a key to the most-recently-used end without
	// changing its value. Returns false if key is absent.
	Touch(key string) bool

	// At returns the key and value at slice position i, for random
	// or positional access into the underlying ordered storage.
	At(i int) (key string, v any, ok bool)

	// GetField reads an auxiliary field stored in key's entry, such
	// as a usage stat a wrapping hub maintains.
	GetField(key, field string) (any, bool)

	// SetField writes an auxiliary field into key's entry.
	SetField(key, field string, v any) bool

	// Empty returns true if the underlying implementation is nil
	Empty() bool

	// Base returns value of underlying implementation
	Base() any
}

// PropList implementation backed entirely by Go, no external owner. Each
// entry is a map so callers can attach fields beyond key/val -- usage
// stats, say -- without this type changing shape.
type proplist struct {
	entries []map[string]any
}

// NewPropList returns an empty native PropList
func NewPropList() PropList {
	return &proplist{}
}

func (p *proplist) indexOf(key string) int {
	for i, e := range p.entries {
		if e["key"] == key {
			return i
		}
	}
	return -1
}

// Get returns a key's value, and whether it was found
func (p *proplist) Get(key string) (any, bool) {
	i := p.indexOf(key)
	if i < 0 {
		return nil, false
	}
	return p.entries[i]["val"], true
}

// Set adds or updates a key's value, appending new keys in order
func (p *proplist) Set(key string, v any) PropList {
	if i := p.indexOf(key); i >= 0 {
		p.entries[i]["val"] = v
		return p
	}
	p.entries = append(p.entries, map[string]any{"key": key, "val": v})
	return p
}

// Delete removes a key, returning its value and whether it was found
func (p *proplist) Delete(key string) (any, bool) {
	i := p.indexOf(key)
	if i < 0 {
		return nil, false
	}
	v := p.entries[i]["val"]
	p.entries = append(p.entries[:i], p.entries[i+1:]...)
	return v, true
}

// Has returns true if key is present
func (p *proplist) Has(key string) bool {
	return p.indexOf(key) >= 0
}

// Len returns the number of entries
func (p *proplist) Len() int {
	return len(p.entries)
}

// Keys returns all keys in insertion order
func (p *proplist) Keys() []string {
	out := make([]string, len(p.entries))
	for i, e := range p.entries {
		out[i] = e["key"].(string)
	}
	return out
}

// Touch moves key to the most-recently-used end
func (p *proplist) Touch(key string) bool {
	i := p.indexOf(key)
	if i < 0 {
		return false
	}
	e := p.entries[i]
	p.entries = append(p.entries[:i], p.entries[i+1:]...)
	p.entries = append(p.entries, e)
	return true
}

// At returns the key/value at slice position i
func (p *proplist) At(i int) (string, any, bool) {
	if i < 0 || i >= len(p.entries) {
		return "", nil, false
	}
	return p.entries[i]["key"].(string), p.entries[i]["val"], true
}

// GetField reads an auxiliary field from key's entry
func (p *proplist) GetField(key, field string) (any, bool) {
	i := p.indexOf(key)
	if i < 0 {
		return nil, false
	}
	v, ok := p.entries[i][field]
	return v, ok
}

// SetField writes an auxiliary field into key's entry
func (p *proplist) SetField(key, field string, v any) bool {
	i := p.indexOf(key)
	if i < 0 {
		return false
	}
	p.entries[i][field] = v
	return true
}

// Empty returns true if the underlying implementation is nil
func (p *proplist) Empty() bool {
	return p == nil
}

// Base returns value of underlying implementation
func (p *proplist) Base() any {
	return p
}
