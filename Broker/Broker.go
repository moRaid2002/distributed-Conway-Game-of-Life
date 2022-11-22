package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"strconv"
	"time"
	"uk.ac.bris.cs/gameoflife/gol/stubs"
)

var x = 1

func makeCall(client *rpc.Client, req stubs.Request, res *stubs.Response) {

	client.Call(stubs.GameOfLifeHandler, req, res)

}

type Broker struct{}

func (s *Broker) Client(req stubs.Request, res *stubs.Response) (err error) {
	if req.P.Turns == 0 {
		res.NewState = req.CurrentStates
		return
	}
	server := flag.String("server"+strconv.Itoa(x), "44.201.122.214:8030", "IP:port string to connect to as server")
	x++
	flag.Parse()
	client, _ := rpc.Dial("tcp", *server)
	defer client.Close()
	newState := req.CurrentStates
	response := new(stubs.Response)
	for turns := 0; turns < req.P.Turns; turns++ {
		request := stubs.Request{newState, req.P, 0, "", 0, req.P.ImageHeight}
		makeCall(client, request, response)

		newState = response.NewState
	}
	res.NewState = response.NewState
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

	rpc.Accept(listener)

	fmt.Println("end")

}
