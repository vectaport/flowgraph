package main

import (
	"github.com/vectaport/fgbase"
	"github.com/vectaport/flowgraph"

	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync/atomic"
	"time"
)

/* DuckPondC Flowgraph HDL *

nestW()(duckImport00)
sinkW(duckSinkW)()

pond00(duckImport00)(duckExport00) {
        swim(duckImport00)(duckExport00)
}
steer00(duckExport00)(duckImport10, duckSinkW)

pond10(duckImport10)(duckExport10) {
        swim(duckImport10)(duckExport10)
}
steer10(duckExport10)(duckImport00, duckSinkE)

nestE()(duckExport10)
sinkE(duckSinkE)()

*/

var duckCnt int64 = -1

type duck struct {
	ID    int64
	Loops int
	Orig  string
	Curr  string
	Exit  bool
	Flew  bool
	Steer int
	Rando int
}

func (d *duck) Break() bool {
	if d.Exit {
		d.Flew = true
	}
	return d.Exit
}

func (d *duck) Clear() {
	d.Exit = false
}

func dialPort(hostPort string) (conn net.Conn) {
	conn, err := net.Dial("tcp", hostPort)
	if conn == nil {
		fgbase.StderrLog.Printf("Connection to %s refused\n", hostPort)
		os.Exit(1)
	}
	if err != nil {
		fgbase.StderrLog.Printf("%v\n", err)
		os.Exit(1)
	}
	return
}

func buffConn(hostPort string) (rw *bufio.ReadWriter) {
	if fgbase.DotOutput {
		return nil
	}
	conn := dialPort(hostPort)
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	return bufio.NewReadWriter(reader, writer)
}

func readNewLineAck(h flowgraph.Hub, rw *bufio.ReadWriter) {
	_, err := rw.ReadString('\n')
	if err != nil {
		h.LogError("%v", err)
	}
}

func writeCommand(h flowgraph.Hub, rw *bufio.ReadWriter, cmd string) {
	_, err := rw.WriteString(cmd)
	if err != nil {
		h.LogError("%v", err)
	}
	rw.Flush()
}

type nestC struct {
	rw         *bufio.ReadWriter
	loc        string
	wcnt, ecnt int
}

func (n *nestC) Retrieve(h flowgraph.Hub) (result interface{}, err error) {
	dc := atomic.AddInt64(&duckCnt, 1)
	d := duck{dc, 0, h.Name()[len(h.Name())-1:], h.Name()[len(h.Name())-1:], false, false, -1, rand.Intn(1000)}
	result = &d

	animstr := func() string {
		dx, dy := 0, 0
		ox, oy := 0, 0
		if n.loc == "W" {
			ox = -360
			dx = 4
			n.wcnt++
		} else {
			ox = 360
			dx = -4
			n.ecnt++
		}
		return fmt.Sprintf("dal.stepl=list();tal=list(:attr);dal.stepl,tal;tal.ID=dal.ID;tal.dx=%d;tal.dy=%d;tal.nsteps=30;dal.ox=%d;dal.oy=%d;", dx, dy, ox, oy)
	}()

	flipstr := func() string {
		if n.loc == "W" {
			return "fliph();"
		}
		return ""
	}()

	movestr := func() string {
		if n.loc == "W" {
			return fmt.Sprintf("move(-560 mod(%d %d)*16-32);", n.wcnt, fgbase.ChannelSize)
		}
		return fmt.Sprintf("move(530 mod(%d %d)*16-32);", n.ecnt, fgbase.ChannelSize)
	}()

	// write command
	s := fmt.Sprintf("global(duckcnt)=duckcnt+1;"+
		"dal=list(:attr);"+
		"dal.ID=%d;"+
		"dal.duck=import(\"MallardMale.drs\");"+
		flipstr+
		"g=text(dtos(gtod(dal.duck 0,0)) \"%02d\");"+
		"dal.duck=growgroup(dal.duck g);"+
		animstr+
		"at(ducks %d :set dal);"+
		movestr+
		"update\n",
		d.ID, d.ID, d.ID)
	writeCommand(h, n.rw, s)

	// handle ack by new-line
	readNewLineAck(h, n.rw)

	return
}

type swimC struct {
	rw    *bufio.ReadWriter
	Count int
}

func (s *swimC) Transform(h flowgraph.Hub, source []interface{}) (result []interface{}, err error) {
	d := source[0].(*duck)
	if d.Flew == true || d.Loops == 0 {
		s.Count++
		d.Flew = false
	}
	d.Exit = rand.Intn(1000) >= 996
	if d.Exit {
		s.Count--
	}
	d.Loops++
	result = []interface{}{d}

	// write command
	movstr := "tal.nsteps=10;" +
		"tal.dx=int(rand(0,3))-1;" +
		"tal.dy=int(rand(0,3))-1;"
	fly := false
	if d.Exit && d.Loops%2 == 1 {
		fly = true
		if d.Curr == "W" {
			movstr = "tal.nsteps=40;tal.dx=(360-dal.ox)/40;tal.dy=(300-dal.oy)/40;dal.ox=dal.ox-380;dal.curr=\"E\";"
			d.Curr = "E"
		} else {
			movstr = "tal.nsteps=40;tal.dx=-(360+dal.ox)/40;tal.dy=(300-dal.oy)/40;dal.ox=dal.ox+380;dal.curr=\"W\";"
			d.Curr = "W"
		}
	}
	cmd := fmt.Sprintf(""+
		"dal=at(ducks %d);"+
		"dal.swim=1;"+
		"tal=list(:attr);"+
		"tal.ID=dal.ID;"+
		movstr+
		"if(%t :then dal.swim=2);"+
		"if(size(dal.stepl)==0||%t :then dal.stepl,tal)\n",
		source[0].(*duck).ID, d.Exit, fly)
	writeCommand(h, s.rw, cmd)

	// handle ack by new-line
	readNewLineAck(h, s.rw)

	return
}

type steerDuck struct {
}

func (s *steerDuck) Transform(h flowgraph.Hub, source []interface{}) (result []interface{}, err error) {
	d := source[0].(*duck)
	if d.Loops%2 == 0 {
		d.Steer = 0
		result = []interface{}{d, nil}
	} else {
		d.Steer = 1
		result = []interface{}{nil, d}
	}
	return
}

type sinkC struct {
	rw *bufio.ReadWriter
}

func (k *sinkC) Sink(source []interface{}) {

	if source[0].(*duck).ID < 0 {
		return
	}
	fmt.Printf("// Duck %+v leaving pond\n", source[0])

	// write command
	id := source[0].(*duck).ID
	s := fmt.Sprintf("global(duckcnt)=duckcnt-1;dd=at(ducks %d);dd.swim=0;select(dd.duck);colors(12 11);if(false :then at(ducks %d :set nil);delete(dd.duck));update;select(:clear);colors(1 10)\n", id, id)
	k.rw.WriteString(s)
	k.rw.Flush()

	// handle ack by new-line
	k.rw.ReadString('\n')
}

func main() {
	fmt.Printf("BEGIN:  TestDuckPondC\n")
	var gridLock = false
	flag.BoolVar(&gridLock, "gridlock", false, "demo gridlock")
	flowgraph.ParseFlags()
	
	oldRunTime := fgbase.RunTime
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second * 1000
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestDuckPondC")

	duckImport00 := fg.NewStream("duckImport00")
	duckWait00 := fg.NewStream("duckWait00")
	duckSinkW := fg.NewStream("duckSinkW").Init(&duck{-2, 0, "W", "W", false, false, -1, rand.Intn(1000)})
	duckFakeW := duckSinkW
	if gridLock {
		duckFakeW = fg.NewStream("duckFakeW").Const(&duck{-4, 0, "W", "W", false, false, -1, rand.Intn(1000)})
	}
	duckExport00 := fg.NewStream("duckExport00")

	duckImport10 := fg.NewStream("duckImport10")
	duckWait10 := fg.NewStream("duckWait10")
	duckSinkE := fg.NewStream("duckSinkE").Init(&duck{-1, 0, "E", "E", false, false, -1, rand.Intn(1000)})
	duckFakeE := duckSinkE
	if gridLock {
		duckFakeE = fg.NewStream("duckFakeE").Const(&duck{-3, 0, "E", "E", false, false, -1, rand.Intn(1000)})
	}
	duckExport10 := fg.NewStream("duckExport10")

	fg.NewHub("nestW", flowgraph.Retrieve, &nestC{rw: buffConn("localhost:20002"), loc: "W"}).
		ConnectResults(duckWait00)
	fg.NewHub("shoreW", flowgraph.Wait, nil).
		ConnectSources(duckWait00, duckFakeW).ConnectResults(duckImport00)
	fg.NewHub("sinkW", flowgraph.Sink, &sinkC{rw: buffConn("localhost:20002")}).
		ConnectSources(duckSinkW)

	pond00 := fg.NewGraphHub("pond00", flowgraph.While)
	pond00.ConnectSources(duckImport00).ConnectResults(duckExport00)
	pond00.NewHub("swim00", flowgraph.AllOf, &swimC{rw: buffConn("localhost:20002")}).
		SetNumSource(1).SetNumResult(1)
	pond00.Loop()
	fg.NewHub("steer00", flowgraph.AllOf, &steerDuck{}).
		ConnectSources(duckExport00).ConnectResults(duckSinkW, duckImport10)

	pond10 := fg.NewGraphHub("pond10", flowgraph.While)
	pond10.ConnectSources(duckImport10).ConnectResults(duckExport10)
	pond10.NewHub("swim10", flowgraph.AllOf, &swimC{rw: buffConn("localhost:20002")}).
		SetNumSource(1).SetNumResult(1)
	pond10.Loop()
	fg.NewHub("steer10", flowgraph.AllOf, &steerDuck{}).
		ConnectSources(duckExport10).ConnectResults(duckSinkE, duckImport00)

	fg.NewHub("nestE", flowgraph.Retrieve, &nestC{rw: buffConn("localhost:20002"), loc: "E"}).
		ConnectResults(duckWait10)
	fg.NewHub("shoreE", flowgraph.Wait, nil).
		ConnectSources(duckWait10, duckFakeE).ConnectResults(duckImport10)
	fg.NewHub("sinkE", flowgraph.Sink, &sinkC{rw: buffConn("localhost:20002")}).
		ConnectSources(duckSinkE)

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestDuckPondC\n")
}
