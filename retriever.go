package flowgraph

import (
)

// Retriever retrieves one value using the Retrieve method.
// Provide as init arg to NewHub with Retrieve HubCode.
// Use Hub.Tracef for tracing.
type Retriever interface {
	Retrieve(h Hub) (result interface{}, err error)
}

