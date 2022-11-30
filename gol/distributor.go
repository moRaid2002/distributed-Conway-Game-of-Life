package gol

import (
	"flag"
	"fmt"
	"net/rpc"
	"strconv"
	"sync"
	"time"
	"uk.ac.bris.cs/gameoflife/gol/stubs"
	"uk.ac.bris.cs/gameoflife/gol/subParams"
	"uk.ac.bris.cs/gameoflife/util"
)

//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   	Hiro

// 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  	MO
type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
	keyPresses <-chan rune
}

var Lastturn = 0
var Lastturnx = 0
var LastState [][]byte
var diff = 0

func makeCall(client *rpc.Client, channel chan *rpc.Call, req stubs.Request, res *stubs.Response) {

	client.Go(stubs.BrokerClient, req, res, channel)

}
func LiveView(client *rpc.Client, c distributorChannels, newWorld *[][]byte, p subParams.Params) {

	req := stubs.Request{*newWorld, p, 0, "", 0, p.ImageHeight, 0, ""}
	res := new(stubs.Response)
	client.Call(stubs.BrokerLiveView, req, res)
	if Lastturnx < res.Turn {
		for h := 0; h < p.ImageHeight; h++ {
			for w := 0; w < p.ImageWidth; w++ {
				if res.NewState[h][w] != LastState[h][w] {
					c.events <- CellFlipped{res.Turn, util.Cell{w, h}}
				}
			}
		}
		LastState = res.NewState
		Lastturnx = res.Turn
	}
}
func Alive(client *rpc.Client, c distributorChannels, flags *bool, newWorld *[][]byte, p subParams.Params) {

	req := stubs.Request{*newWorld, p, 0, "", 0, p.ImageHeight, 0, ""}
	res := new(stubs.Response)
	client.Call(stubs.BrokerAlive, req, res)
	if *flags && res.Turn > Lastturn {
		Lastturn = res.Turn
		c.events <- TurnComplete{res.Turn}
		c.events <- AliveCellsCount{res.Turn, res.Alive}
	}

}
func Press(client *rpc.Client, keypress string, newWorld *[][]byte, p subParams.Params, c distributorChannels) {

	req := stubs.Request{*newWorld, p, 0, keypress, 0, p.ImageHeight, 0, ""}
	res := new(stubs.Response)
	client.Call(stubs.BrokerKeyPress, req, res)
	if keypress == "s" {
		c.ioCommand <- ioOutput
		filename2 := "current-state-s"
		c.ioFilename <- filename2
		for h := 0; h < p.ImageHeight; h++ {
			for w := 0; w < p.ImageWidth; w++ {
				c.ioOutput <- res.NewState[h][w]
			}
		}

	}

}

func client(newWorld *[][]byte, p subParams.Params, server2 string, c distributorChannels, flags *bool) {
	server := flag.String(server2, "44.202.187.144:8030", "IP:port string to connect to as server")
	flag.Parse()
	client, _ := rpc.Dial("tcp", *server)
	defer client.Close()
	req := stubs.Request{*newWorld, p, 0, "", 0, p.ImageHeight, 0, ""}
	res := new(stubs.Response)
	channel := make(chan *rpc.Call, 10)
	makeCall(client, channel, req, res)

	ticker := time.NewTicker(time.Second * 2)
	mutex := sync.Mutex{}

	go func(flags *bool) {
		for {
			if *flags {
				LiveView(client, c, newWorld, p)
			}
		}
	}(flags)
	go func(mutex *sync.Mutex) {
		for {
			receivingKeyPress := <-c.keyPresses
			switch receivingKeyPress {
			case 'p':
				Press(client, "p", newWorld, p, c)
			case 's':
				Press(client, "s", newWorld, p, c)

			case 'q':
				mutex.Lock()
				fmt.Println("stopping client")
				Press(client, "q", newWorld, p, c)
				mutex.Unlock()
			case 'k':
				mutex.Lock()
				fmt.Println("stopping")
				Press(client, "k", newWorld, p, c)
				mutex.Unlock()
			}

		}

	}(&mutex)

	go func(mutex *sync.Mutex) {
		for {
			select {
			case <-ticker.C:
				mutex.Lock()
				Alive(client, c, flags, newWorld, p)
				mutex.Unlock()
			}
		}
	}(&mutex)

	select {
	case <-channel:
		*newWorld = res.NewState

	}

}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	// TODO: Create a 2D slice to store the world.
	newWorld := make([][]byte, p.ImageHeight) // creating the empty 2D slice for Height and Width
	for i := range newWorld {
		newWorld[i] = make([]byte, p.ImageWidth)
	}
	//----------------------------------- Input(Reading) of the PGM image --------------------------//

	filename := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight) // Convert number to String, using Itoa.
	c.ioCommand <- ioInput                                                     // readPgmImage opens a pgm file and sends its data as an array of bytes.

	c.ioFilename <- filename // send the converted file to ioFilename

	for h := 0; h < p.ImageHeight; h++ {
		for w := 0; w < p.ImageWidth; w++ {

			newWorld[h][w] = <-c.ioInput
			if newWorld[h][w] == 255 {
				// If the cell is alive then you need to notify Using CellFlipped event.
				c.events <- CellFlipped{0, util.Cell{w, h}}
			}
		}
	}

	// TODO: Execute all turns of the Game of Life.
	flag := true
	x := subParams.Params{p.Turns, p.Threads, p.ImageWidth, p.ImageHeight}
	LastState = newWorld
	diff++
	client(&newWorld, x, filename+"-"+strconv.Itoa(p.Turns)+"-"+strconv.Itoa(p.Threads)+"-"+strconv.Itoa(diff), c, &flag)

	// TODO: Report the final state using FinalTurnCompleteEvent.
	/*	c.ioCommand <- ioOutput
		filename = filename + "x" + strconv.Itoa(p.Turns)
		c.ioFilename <- filename

		for h := 0; h < p.ImageHeight; h++ {
			for w := 0; w < p.ImageWidth; w++ {

				c.ioOutput <- newWorld[h][w]

			}
		}

	*/

	// Make sure that the Io has finished any output before exiting.
	new := make([]util.Cell, 0)
	for h := 0; h < p.ImageHeight; h++ {
		for w := 0; w < p.ImageWidth; w++ {

			if newWorld[h][w] == 255 {
				cell := util.Cell{w, h}
				new = append(new, cell)
			}
		}
	}
	c.events <- FinalTurnComplete{
		p.Turns,
		new,
	}
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
	flag = false
}
