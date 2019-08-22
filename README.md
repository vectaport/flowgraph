# flowgraph

### Getting Started

```
go get -u github.com/vectaport/flowgraph
go test
```

### Links

* [![GoDoc](https://godoc.org/github.com/vectaport/flowgraph?status.svg)](https://godoc.org/github.com/vectaport/flowgraph)
* [Wiki](https://github.com/vectaport/flowgraph/wiki)
* [Slides](https://www.youtube.com/watch?v=awAsZUBncG8) from [Minneapolis Golang Meetup, May 22nd 2019](https://www.meetup.com/Minneapolis-Golang/events/259276080/)

### Overview

Flowgraphs are built out of hubs interconnected by streams. The hubs are implemented with goroutines that use select to wait on incoming data or back-pressure handshakes. The data and handshakes travel on streams implemented with channels of empty interfaces for forward flow (interface{}) and channels of empty structs for back-pressure (struct{}).

The user of this package is completely isolated from the details of using goroutines, channels, and select, and only has to provide the empty interface functions that transform incoming data into outgoing data as needed for each hub of the flowgraph under construction. It includes the ability to log each data flow and transformation at the desired level of detail for debugging and monitoring purposes. 

The package allows for correct-by-construction dataflow systems that avoid deadlock and gridlock by using back-pressure to manage empty space.   It also supports looping constructs that can operate at the same efficiency as pipeline structures using channel buffering within the loop. 

All of this is made available with an API designed to directly underlie a future HDL for a flowgraph language.

