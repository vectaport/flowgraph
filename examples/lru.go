/* lru Flowgraph HDL

driver()(.X(e0))
cache(.A(e0))(.X(e1))
sink(.A(e1))()

An LRU cache kept in a flowgraph.PropList: list order IS recency order, so
a hit touches its entry to the end and a miss past capacity evicts index 0.
*/

package main

import (
	"fmt"
	"math/rand"

	"github.com/vectaport/fgbase"
	"github.com/vectaport/flowgraph"
)

const (
	keyspace = 200
	cap_     = 32
	nreq     = 5000
)

type lruCache struct {
	store                    flowgraph.PropList
	hitcnt, misscnt, evicted int
}

func (c *lruCache) Transform(h flowgraph.Hub, source []any) (result []any, err error) {
	// allOfFire calls Transform even on the Array hub's EOS marker (an
	// error value, not a key); its result is discarded once EOS fires,
	// so just avoid the panic.
	key, ok := source[0].(string)
	if !ok {
		return []any{nil}, nil
	}

	if _, ok := c.store.Get(key); ok {
		c.hitcnt++
		c.store.Touch(key)
		hits, _ := c.store.GetField(key, "hits")
		c.store.SetField(key, "hits", hits.(int)+1)
		return []any{true}, nil
	}

	c.misscnt++
	if c.store.Len() >= cap_ {
		victim, _, _ := c.store.At(0)
		c.store.Delete(victim)
		c.evicted++
	}
	c.store.Set(key, key)
	c.store.SetField(key, "hits", 0)
	return []any{false}, nil
}

type lruSink struct {
	cache *lruCache
}

func (s *lruSink) Sink(source []any) {}

func main() {
	fgbase.TraceStyle = fgbase.New
	fgbase.ConfigByFlag(map[string]any{"trace": "QQ"})

	fg := flowgraph.New("lru")

	rand.Seed(1)
	reqs := make([]any, nreq)
	for i := range reqs {
		idx := int(rand.Float64() * rand.Float64() * float64(keyspace))
		reqs[i] = fmt.Sprintf("k%d", idx)
	}

	driver := fg.NewHub("driver", flowgraph.Array, reqs).
		SetResultNames("X")

	cache := &lruCache{store: flowgraph.NewPropList()}
	cacheHub := fg.NewHub("cache", flowgraph.AllOf, cache).
		SetSourceNames("A").
		SetResultNames("X")

	sink := fg.NewHub("sink", flowgraph.Sink, &lruSink{cache}).
		SetSourceNames("A")

	fg.Connect(driver, "X", cacheHub, "A")
	fg.Connect(cacheHub, "X", sink, "A")

	fg.Run()

	fmt.Printf("\nkeyspace=%d cap=%d requests=%d\n", keyspace, cap_, nreq)
	fmt.Printf("hits=%d misses=%d evictions=%d\n", cache.hitcnt, cache.misscnt, cache.evicted)
	fmt.Printf("hit rate = %.1f%%\n", 100*float64(cache.hitcnt)/float64(nreq))
}
