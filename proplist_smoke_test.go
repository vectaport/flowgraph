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
}
