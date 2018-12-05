package flowgraph

import ()

// Transformer transforms a slice of source values into a slice
// of result values with the Transform method. Provide as init arg to
// NewHub with AllOf or OneOf HubCode or optionally with a logic or
// math HubCode (access HubCode from a Transform method to customize
// these transforms). Use Hub.Tracef for tracing.
type Transformer interface {
	Transform(h Hub, source []interface{}) (result []interface{}, err error)
}
