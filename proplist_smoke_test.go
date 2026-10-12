package flowgraph

import "testing"

func TestPropListSmoke(t *testing.T) {
	p := NewPropList()
	p.Set("a", 1).Set("b", 2).Set("c", 3)

	if got := p.Keys(); len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("Keys() = %v, want [a b c]", got)
	}

	if !p.Touch("a") {
		t.Fatalf("Touch(a) = false, want true")
	}
	if got := p.Keys(); got[len(got)-1] != "a" {
		t.Fatalf("Keys() after Touch(a) = %v, want a last", got)
	}

	if k, v, ok := p.At(0); !ok || k != "b" || v != 2 {
		t.Fatalf("At(0) = %v,%v,%v, want b,2,true", k, v, ok)
	}

	if ok := p.SetField("b", "hits", 1); !ok {
		t.Fatalf("SetField(b,hits) = false, want true")
	}
	if v, ok := p.GetField("b", "hits"); !ok || v != 1 {
		t.Fatalf("GetField(b,hits) = %v,%v, want 1,true", v, ok)
	}
	if _, ok := p.GetField("missing", "hits"); ok {
		t.Fatalf("GetField(missing,hits) ok = true, want false")
	}

	if v, ok := p.Delete("c"); !ok || v != 3 {
		t.Fatalf("Delete(c) = %v,%v, want 3,true", v, ok)
	}
	if p.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", p.Len())
	}
	if p.Touch("nope") {
		t.Fatalf("Touch(nope) = true, want false")
	}

	p.Set("b", 22)
	if v, ok := p.Get("b"); !ok || v != 22 {
		t.Fatalf("Get(b) after update = %v,%v, want 22,true", v, ok)
	}
	if p.Len() != 2 {
		t.Fatalf("Len() after updating existing key = %d, want 2", p.Len())
	}
	if v, ok := p.GetField("b", "hits"); !ok || v != 1 {
		t.Fatalf("GetField(b,hits) after Set update = %v,%v, want 1,true (update must not reset fields)", v, ok)
	}

	if ok := p.SetField("b", "key", "hijacked"); ok {
		t.Fatalf("SetField(b,key,...) = true, want false (key/val are reserved)")
	}
	if ok := p.SetField("b", "val", "hijacked"); ok {
		t.Fatalf("SetField(b,val,...) = true, want false (key/val are reserved)")
	}
}

func TestDupPropList(t *testing.T) {
	src := NewPropList()
	src.Set("a", 1).Set("b", 2).Set("c", 3)
	src.Touch("a")

	clone := DupPropList(src)
	if got := clone.Keys(); len(got) != 3 || got[0] != "b" || got[1] != "c" || got[2] != "a" {
		t.Fatalf("Keys() = %v, want [b c a]", got)
	}
	for _, want := range []struct {
		key string
		val any
	}{{"a", 1}, {"b", 2}, {"c", 3}} {
		if v, ok := clone.Get(want.key); !ok || v != want.val {
			t.Fatalf("Get(%s) = %v,%v, want %v,true", want.key, v, ok, want.val)
		}
	}

	// clone is independent of src
	clone.Set("a", 99)
	if v, _ := src.Get("a"); v != 1 {
		t.Fatalf("src.Get(a) = %v after cloning, want unchanged 1", v)
	}
}

// slicePropList is a second PropList implementation, distinct from
// proplist's own storage, so DupPropList's test proves it clones
// through the interface (At) rather than happening to work only
// because both sides share one concrete type.
type slicePropList struct {
	keys []string
	vals []any
}

func (p *slicePropList) Get(key string) (any, bool) {
	for i, k := range p.keys {
		if k == key {
			return p.vals[i], true
		}
	}
	return nil, false
}
func (p *slicePropList) Set(key string, v any) PropList {
	p.keys = append(p.keys, key)
	p.vals = append(p.vals, v)
	return p
}
func (p *slicePropList) Delete(string) (any, bool) { return nil, false }
func (p *slicePropList) Has(string) bool           { return false }
func (p *slicePropList) Len() int                  { return len(p.keys) }
func (p *slicePropList) Keys() []string            { return p.keys }
func (p *slicePropList) Touch(string) bool         { return false }
func (p *slicePropList) At(i int) (string, any, bool) {
	if i < 0 || i >= len(p.keys) {
		return "", nil, false
	}
	return p.keys[i], p.vals[i], true
}
func (p *slicePropList) GetField(string, string) (any, bool) { return nil, false }
func (p *slicePropList) SetField(string, string, any) bool   { return false }
func (p *slicePropList) Empty() bool                         { return p == nil }
func (p *slicePropList) Base() any                           { return p }

func TestDupPropListFromAlternateSource(t *testing.T) {
	src := &slicePropList{}
	src.Set("x", 10).Set("y", 20).Set("z", 30)

	clone := DupPropList(src)
	for _, want := range []struct {
		key string
		val any
	}{{"x", 10}, {"y", 20}, {"z", 30}} {
		if v, ok := clone.Get(want.key); !ok || v != want.val {
			t.Fatalf("Get(%s) = %v,%v, want %v,true", want.key, v, ok, want.val)
		}
	}
	if got := clone.Keys(); len(got) != 3 || got[0] != "x" || got[1] != "y" || got[2] != "z" {
		t.Fatalf("Keys() = %v, want [x y z]", got)
	}
}
