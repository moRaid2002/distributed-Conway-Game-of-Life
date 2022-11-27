package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/rpc"
	"strconv"
	"sync"
	"time"
	"uk.ac.bris.cs/gameoflife/gol/stubs"
)

var x = 1
var y = 1
var IpAddresses []string

func makeCall(client *rpc.Client, req stubs.Request, res *stubs.Response) {

	client.Call(stubs.GameOfLifeHandler, req, res)

}
func Encode(State [][]byte) {
	newWorld := State
	for h := 0; h < len(newWorld); h++ {
		for w := 0; w < len(newWorld); w++ {
			if newWorld[h][w] == 255 {
				newWorld[h][w] = 1
			}
		}
	}
	encoded := make([]int, len(newWorld))
	for i := 0; i < len(newWorld); i++ {
		x := 0
		for h := 0; h < len(newWorld); h++ {
			x = x + int(newWorld[i][h])*(int(math.Pow(2, float64(h))))
		}
		encoded[i] = x
	}
	fmt.Println(encoded)
	Decode(encoded)
}
func Decode(encode []int) {
	newWorld := make([][]byte, len(encode))
	for i := 0; i < len(encode); i++ {
		newWorld[i] = make([]byte, len(encode))
	}
	for h := 0; h < len(newWorld); h++ {
		for w := 0; w < len(newWorld); w++ {
			newWorld[h][w] = byte(encode[h] % 2)
			encode[h] = int(encode[h] / 2)
		}
	}
	for h := 0; h < len(newWorld); h++ {
		for w := 0; w < len(newWorld); w++ {
			if newWorld[h][w] == 1 {
				newWorld[h][w] = 255
			}
		}
	}
	fmt.Println(newWorld)
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

func AddIp(str string) {
	for i := range IpAddresses {
		if IpAddresses[i] == str {
			return
		}
	}
	IpAddresses = append(IpAddresses, str)
}
func (s *Broker) AddIpServer(req stubs.Request, res *stubs.Response) (err error) {
	AddIp(req.Ip)
	fmt.Println(req.Ip)
	fmt.Println("Ip received")
	return
}

func (s *Broker) LiveView(req stubs.Request, res *stubs.Response) (err error) {
	if req.P.ImageHeight == len(currentState) {
		mutex.Lock()
		res.NewState = currentState
		res.Turn = Turn
		mutex.Unlock()
	}
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
	fmt.Println("enter")

	currentState = req.CurrentStates

	if req.P.Turns == 0 {
		currentState = req.CurrentStates
		res.NewState = req.CurrentStates

		return
	}

	var servers []*string
	var Clients []*rpc.Client

	for i := range IpAddresses {
		servers = append(servers, flag.String("server-"+strconv.Itoa(i)+"-"+strconv.Itoa(x), IpAddresses[i]+":8030", "IP:port string to connect to as server"))
	}

	x++
	flag.Parse()
	for i := range servers {
		clients, _ := rpc.Dial("tcp", *servers[i])
		Clients = append(Clients, clients)

	}

	numberOfAWS := len(servers)

	newState := req.CurrentStates
	turns := 0
	if simiend {
		newState = currentState
		turns = Turn
		simiend = false
	}
	currentState = req.CurrentStates
	fmt.Println(req.CurrentStates)
	Encode(req.CurrentStates)
	for turns < req.P.Turns && !end && !simiend {

		var requests []stubs.Request
		var responses []*stubs.Response

		for i := 0; i < numberOfAWS; i++ {
			requests = append(requests, stubs.Request{newState, req.P, 0, "", numberOfAWS, i * int(req.P.ImageHeight/numberOfAWS), i + 1, ""})
			responses = append(responses, new(stubs.Response))
		}

		for i := 0; i < numberOfAWS; i++ {
			makeCall(Clients[i], requests[i], responses[i])
		}

		mutex.Lock()
		newState = nil

		for i := 0; i < numberOfAWS; i++ {

			newState = append(newState, responses[i].NewState...)

		}
		if len(newState) != req.P.ImageHeight {

			IpAddresses = nil
			for i := range Clients {
				Clients[i].Call(stubs.GameOfLifeSend, new(stubs.Response), new(stubs.Request))
			}
			numberOfAWS = len(IpAddresses)
			servers = nil

			for i := range IpAddresses {
				servers = append(servers, flag.String("server-"+strconv.Itoa(i)+"-"+"-"+strconv.Itoa(x)+"-"+strconv.Itoa(y), IpAddresses[i]+":8030", "IP:port string to connect to as server"))
			}
			y++
			Clients = nil
			for i := range servers {
				clients, _ := rpc.Dial("tcp", *servers[i])
				Clients = append(Clients, clients)
			}

			newState = currentState
		} else {

			turns++
			Turn = turns
			currentState = newState
		}
		mutex.Unlock()
	}
	if end {
		for i := 0; i < numberOfAWS; i++ {
			StopAll(Clients[i])
		}
		stop = true
	}

	res.NewState = newState
	fmt.Println(res.NewState)
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
