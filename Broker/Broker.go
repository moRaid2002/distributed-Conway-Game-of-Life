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
	server := flag.String("server-1-"+strconv.Itoa(x), "44.201.122.214:8030", "IP:port string to connect to as server")
	server2 := flag.String("server-2-"+strconv.Itoa(x), "54.90.210.50:8030", "IP:port string to connect to as server")
	server3 := flag.String("server-3-"+strconv.Itoa(x), "18.212.57.66:8030", "IP:port string to connect to as server")
	server4 := flag.String("server-4-"+strconv.Itoa(x), "35.172.135.168:8030", "IP:port string to connect to as server")
	x++
	flag.Parse()
	client, _ := rpc.Dial("tcp", *server)
	client2, _ := rpc.Dial("tcp", *server2)
	client3, _ := rpc.Dial("tcp", *server3)
	client4, _ := rpc.Dial("tcp", *server4)
	defer client.Close()
	defer client2.Close()
	newState := req.CurrentStates

	for turns := 0; turns < req.P.Turns; turns++ {
		request := stubs.Request{newState, req.P, 0, "", 4, 0, 1}
		request2 := stubs.Request{newState, req.P, 0, "", 4, int(req.P.ImageHeight / 4), 2}
		request3 := stubs.Request{newState, req.P, 0, "", 4, 2 * int(req.P.ImageHeight/4), 3}
		request4 := stubs.Request{newState, req.P, 0, "", 4, 4 * int(req.P.ImageHeight/4), 4}
		response := new(stubs.Response)
		response2 := new(stubs.Response)
		response3 := new(stubs.Response)
		response4 := new(stubs.Response)
		makeCall(client, request, response)
		makeCall(client2, request2, response2)
		makeCall(client3, request3, response3)
		makeCall(client4, request4, response4)
		newState = append(response.NewState, response2.NewState...)
		newState = append(newState, response3.NewState...)
		newState = append(newState, response4.NewState...)
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

	rpc.Accept(listener)

	fmt.Println("end")

}
