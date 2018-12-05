package flowgraph

import ()

// Breaker returns true to break out of a loop.
type Breaker interface {
	Break() bool
}
