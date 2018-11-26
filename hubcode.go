package flowgraph

import ()

// HubCode is code arg to NewHub()
type HubCode int

// HubCode constants for NewHub() code arg.
// Comment fields are init arg for NewHub, number of source and number of results,
// and description.  If n or m the number of sources and results are set with
// SetNumSource and SetNumResult (or SetSource and SetResult)
const (
	Nop HubCode = iota

	Retrieve //	Retriever	0,1	retrieve one value with Retrieve method
	Transmit //	Transmitter	1,0	transmit one value with Transmit method
	AllOf    //	Transformer	n,m	waiting for all sources for Transform method
	OneOf    //	Transformer	n,m	waiting for one source for Transform method

	Wait   // 	nil		n+1,1 	wait for last source to pass rest
	Select // 	nil		1+n,1   select from rest by first source
	Steer  // 	nil		2|1,2	steer last source by first source
	Cross  // 	nil         1+2*n,2*n   steer left or right rank by first source

	Array    //	[]interface{}	0,1	produce array of values then EOF
	Constant //	interface{}	0,1	produce constant values forever
	Pass     //	nil		1,1	pass value
	Split    //	nil		1,n     split slice into values
	Join     //	nil		n,1     join values into slice
	Sink     //	[Sinker]	1,0	consume values forever

	Graph  // 	nil		n,m     hub with general purpose internals
	While  // 	nil		n,n	hub with while loop around internals
	During // 	nil		n,n	hub with while loop with continuous results

	Add      //	[Transformer]	2,1	add numbers, concat strings
	Subtract //	[Transformer]	2,1	subtract numbers
	Multiply //	[Transformer]	2,1	multiply numbers
	Divide   //	[Transformer]	2,1	divide numbers
	Modulo   //	[Transformer]	2,1	modulate numbers
	And      //	[Transformer]	2,1	AND bool or bit-wise AND integers
	Or       //	[Transformer]	2,1	OR bool or bit-wise OR integers
	Not      //	[Transformer]	1,1	negate bool, invert integers
	Shift    //	ShiftCode|Transformer	2,1	shift first by second, Arith,Barrel,Signed
)

// String method for HubCode
func (c HubCode) String() string {
	return []string{
		"Nop",

		"Retrieve",
		"Transmit",
		"AllOf",
		"OneOf",

		"Wait",
		"Select",
		"Steer",
		"Cross",

		"Array",
		"Constant",
		"Pass",
		"Split",
		"Join",
		"Sink",

		"Graph",
		"While",
		"During	",

		"Add",
		"Subtract",
		"Multiply",
		"Divide",
		"Modulo",
		"And",
		"Or",
		"Not",
		"Shift",
	}[c]
}

// ShiftCode is the subcode for the "Shift" HubCode, provided as the init arg to NewHub.
type ShiftCode int

const (
	Arith ShiftCode = iota
	Barrel
	Signed
)
