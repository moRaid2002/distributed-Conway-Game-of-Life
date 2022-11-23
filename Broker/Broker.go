package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"strconv"
	"sync"
	"time"
	"uk.ac.bris.cs/gameoflife/gol/stubs"
)

var x = 1

func makeCall(client *rpc.Client, req stubs.Request, res *stubs.Response) {

	client.Call(stubs.GameOfLifeHandler, req, res)

}

type Broker struct{}

var currentState [][]byte
var Turn = 0
var index = 0
var end = false

var mutex = sync.Mutex{}

func (s *Broker) KeyPress(req stubs.Request, res *stubs.Response) (err error) {
	switch req.Keypress {
	case "s":
		res.NewState = currentState
	case "p":
		if index%2 == 0 {
			fmt.Println("Pausing")
			mutex.Lock()
		} else {
			fmt.Println("Continuing")
			mutex.Unlock()
		}
		index++
	case "k":
		end = true

	}

	return
}
func (s *Broker) AliveCell(req stubs.Request, res *stubs.Response) (err error) {
	mutex.Lock()
	count := 0

	for h := 0; h < req.P.ImageHeight; h++ {
		for w := 0; w < req.P.ImageWidth; w++ {
			if currentState[h][w] == 255 {
				count++
			}
		}
	}
	res.Alive = count
	res.Turn = Turn
	mutex.Unlock()

	return
}

func (s *Broker) Client(req stubs.Request, res *stubs.Response) (err error) {
	currentState = req.CurrentStates

	if req.P.Turns == 0 {
		res.NewState = req.CurrentStates
		return
	}

	server := flag.String("server-1-"+strconv.Itoa(x), "54.167.162.157:8030", "IP:port string to connect to as server")
	server2 := flag.String("server-2-"+strconv.Itoa(x), "52.200.11.84:8030", "IP:port string to connect to as server")
	server3 := flag.String("server-3-"+strconv.Itoa(x), "100.27.20.252:8030", "IP:port string to connect to as server")
	server4 := flag.String("server-4-"+strconv.Itoa(x), "44.211.72.146:8030", "IP:port string to connect to as server")
	x++
	flag.Parse()
	client, _ := rpc.Dial("tcp", *server)
	client2, _ := rpc.Dial("tcp", *server2)
	client3, _ := rpc.Dial("tcp", *server3)
	client4, _ := rpc.Dial("tcp", *server4)
	defer client.Close()
	defer client2.Close()
	newState := req.CurrentStates

	for turns := 0; turns < req.P.Turns && !end; turns++ {
		request := stubs.Request{newState, req.P, 0, "", 4, 0, 1}
		request2 := stubs.Request{newState, req.P, 0, "", 4, int(req.P.ImageHeight / 4), 2}
		request3 := stubs.Request{newState, req.P, 0, "", 4, 2 * int(req.P.ImageHeight/4), 3}
		request4 := stubs.Request{newState, req.P, 0, "", 4, 3 * int(req.P.ImageHeight/4), 4}
		response := new(stubs.Response)
		response2 := new(stubs.Response)
		response3 := new(stubs.Response)
		response4 := new(stubs.Response)
		makeCall(client, request, response)
		makeCall(client2, request2, response2)
		makeCall(client3, request3, response3)
		makeCall(client4, request4, response4)

		mutex.Lock()
		newState = append(response.NewState, response2.NewState...)
		newState = append(newState, response3.NewState...)
		newState = append(newState, response4.NewState...)

		Turn = turns + 1

		currentState = newState
		mutex.Unlock()
	}

	res.NewState = newState
	return
}

func main() {

	fmt.Println("Broker working")
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&Broker{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	go func() {

		for {
			if end {
				listener.Close()
			}
		}

	}()
	rpc.Accept(listener)

	fmt.Println("end")

}
