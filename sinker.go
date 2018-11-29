package flowgraph

import (
)

// Sinker consumes wavefronts of values one at a time forever.
// Optionally provide as init arg to NewHub with Sink HubCode.
type Sinker interface {
	Sink(source []interface{})
}
