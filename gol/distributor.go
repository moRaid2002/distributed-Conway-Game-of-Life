package gol

import (
	"flag"
	"fmt"
	"net/rpc"
	"strconv"
	"uk.ac.bris.cs/gameoflife/gol/stubs"
	"uk.ac.bris.cs/gameoflife/gol/subParams"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
	keyPresses <-chan rune
}

func makeCall(client *rpc.Client, message *[][]byte, p subParams.Params, out chan<- int) {
	request := stubs.Request{message, p, out}
	response := new(stubs.Response)
	client.Call(stubs.GameOfLifeHandler, request, response)
	*message = response.NewState
	//fmt.Println("Responded: " + response.Message)
}

func client(newWorld *[][]byte, p subParams.Params, server2 string, out chan<- int) {
	server := flag.String(server2, "52.87.245.176:8030", "IP:port string to connect to as server")
	flag.Parse()
	client, _ := rpc.Dial("tcp", *server)
	defer client.Close()

	makeCall(client, newWorld, p, out)

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

	x := subParams.Params{p.Turns, p.Threads, p.ImageWidth, p.ImageHeight}
	out := make(chan int)
	client(&newWorld, x, filename+"-"+strconv.Itoa(p.Turns)+"-"+strconv.Itoa(p.Threads), out)
	recieved := <-out
	fmt.Println(recieved)
	// TODO: Report the final state using FinalTurnCompleteEvent.
	c.ioCommand <- ioOutput
	filename = filename + "x" + strconv.Itoa(p.Turns)
	c.ioFilename <- filename

	for h := 0; h < p.ImageHeight; h++ {
		for w := 0; w < p.ImageWidth; w++ {

			c.ioOutput <- newWorld[h][w]

		}
	}

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
}
