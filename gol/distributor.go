package gol

import (
	"flag"
	"net/rpc"
	"strconv"
	"uk.ac.bris.cs/gameoflife/gol/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

func makeCall(client *rpc.Client, message [][]byte, p Params) {
	request := stubs.Request{message, p}
	response := new(stubs.Response)
	client.Call(stubs.GameOfLifeHandler, request, response)
	//fmt.Println("Responded: " + response.Message)
}

func client(newWorld [][]byte, p Params) {
	server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	flag.Parse()
	client, _ := rpc.Dial("tcp", *server)
	defer client.Close()

	makeCall(client, newWorld, p)

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
			if newWorld[h][w] == 255 { // If the cell is alive then you need to notify Using CellFlipped event.
				c.events <- CellFlipped{0, util.Cell{w, h}}
			}
		}
	}

	// TODO: Execute all turns of the Game of Life.
	client(newWorld, p)

	// TODO: Report the final state using FinalTurnCompleteEvent.

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
