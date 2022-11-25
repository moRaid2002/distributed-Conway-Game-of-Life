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
var IpAddresses []string

func makeCall(client *rpc.Client, req stubs.Request, res *stubs.Response) {

	client.Call(stubs.GameOfLifeHandler, req, res)

}
func AddIp(str string) {
	IpAddresses = append(IpAddresses, str)
}
func StopAll(client *rpc.Client) {
	req := new(stubs.Request)
	res := new(stubs.Response)
	client.Call(stubs.GameOfLifeStop, req, res)

}

type Broker struct{}

var currentState [][]byte
var Turn = 0
var index = 0
var end = false
var stop = false
var simiend = false

var mutex = sync.Mutex{}

func (s *Broker) LiveView(req stubs.Request, res *stubs.Response) (err error) {
	mutex.Lock()
	res.NewState = currentState
	res.Turn = Turn
	mutex.Unlock()
	return
}
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
	case "q":
		simiend = true
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
	mutex.Lock()
	IpAddresses = nil
	AddIp("18.206.124.19")
	AddIp("18.204.195.121")
	AddIp("34.201.65.245")
	AddIp("54.89.102.20")
	mutex.Unlock()
	var servers []*string
	var Clients []*rpc.Client
	for i := range IpAddresses {
		servers = append(servers, flag.String("server-"+strconv.Itoa(i)+"-"+strconv.Itoa(x), IpAddresses[i]+":8030", "IP:port string to connect to as server"))
	}

	/*
		server := flag.String("server-1-"+strconv.Itoa(x), "18.206.124.19:8030", "IP:port string to connect to as server")
		server2 := flag.String("server-2-"+strconv.Itoa(x), "18.204.195.121:8030", "IP:port string to connect to as server")
		server3 := flag.String("server-3-"+strconv.Itoa(x), "34.201.65.245:8030", "IP:port string to connect to as server")
		server4 := flag.String("server-4-"+strconv.Itoa(x), "54.89.102.20:8030", "IP:port string to connect to as server")

	*/
	x++
	flag.Parse()
	for i := range servers {
		clients, _ := rpc.Dial("tcp", *servers[i])
		Clients = append(Clients, clients)
	}
	/*
		client, _ := rpc.Dial("tcp", *server)
		client2, _ := rpc.Dial("tcp", *server2)
		client3, _ := rpc.Dial("tcp", *server3)
		client4, _ := rpc.Dial("tcp", *server4)

	*/
	//defer client.Close()
	//defer client2.Close()

	numberOfAWS := len(servers)

	newState := req.CurrentStates
	turns := 0
	if simiend {
		newState = currentState
		turns = Turn
		simiend = false
	}
	for turns < req.P.Turns && !end && !simiend {
		var requests []stubs.Request
		var responses []*stubs.Response
		fmt.Println(requests, responses)
		for i := 0; i < numberOfAWS; i++ {
			requests = append(requests, stubs.Request{newState, req.P, 0, "", numberOfAWS, i * int(req.P.ImageHeight/numberOfAWS), i + 1})
			responses = append(responses, new(stubs.Response))
		}
		/*
			request := stubs.Request{newState, req.P, 0, "", 4, 0, 1}
			request2 := stubs.Request{newState, req.P, 0, "", 4, int(req.P.ImageHeight / 4), 2}
			request3 := stubs.Request{newState, req.P, 0, "", 4, 2 * int(req.P.ImageHeight/4), 3}
			request4 := stubs.Request{newState, req.P, 0, "", 4, 3 * int(req.P.ImageHeight/4), 4}
			response := new(stubs.Response)
			response2 := new(stubs.Response)
			response3 := new(stubs.Response)
			response4 := new(stubs.Response)

		*/
		for i := 0; i < numberOfAWS; i++ {
			makeCall(Clients[i], requests[i], responses[i])
		}
		/*
			makeCall(client, request, response)
			makeCall(client2, request2, response2)
			makeCall(client3, request3, response3)
			makeCall(client4, request4, response4)


		*/
		mutex.Lock()
		newState = nil
		fmt.Println(len(responses[0].NewState), len(responses[1].NewState), len(responses[2].NewState), len(responses[3].NewState))
		//fmt.Println(responses[0].NewState)

		for i := 0; i < numberOfAWS; i++ {

			newState = append(newState, responses[i].NewState...)
			fmt.Println(i, newState)

		}
		fmt.Println(newState)
		/*
			newState = append(response.NewState, response2.NewState...)
			newState = append(newState, response3.NewState...)
			newState = append(newState, response4.NewState...)

		*/
		turns++
		Turn = turns
		currentState = newState

		mutex.Unlock()
	}
	/*if end {
		StopAll(client)
		StopAll(client2)
		StopAll(client3)
		StopAll(client4)
		stop = true
	}

	*/

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
			if stop {
				listener.Close()
			}
		}

	}()
	rpc.Accept(listener)

	fmt.Println("end")

}
