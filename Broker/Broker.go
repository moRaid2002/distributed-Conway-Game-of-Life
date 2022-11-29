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
var y = 1
var IpAddresses []string
var IpAddressesCopy []string

func makeCall(client *rpc.Client, req stubs.Request, res *stubs.Response, channel chan *rpc.Call) {
	client.Go(stubs.GameOfLifeHandler, req, res, channel)

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
	IpAddressesCopy = append(IpAddressesCopy, str)
	for i := range IpAddresses {
		if IpAddresses[i] == str {
			return
		}
	}
	fmt.Println(str)
	fmt.Println("Ip received")
	IpAddresses = append(IpAddresses, str)
	//IpAddressesCopy = append(IpAddressesCopy, str)
}
func (s *Broker) AddIpServer(req stubs.Request, res *stubs.Response) (err error) {
	AddIp(req.Ip)
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
		clients, err := rpc.Dial("tcp", *servers[i])
		//Clients = append(Clients, clients)
		if err != nil {
			fmt.Println("server disconnected ")
			servers = append(servers[:i], servers[i+1:]...)
			IpAddresses = append(IpAddresses[:i], IpAddresses[i+1:]...)
		} else {
			Clients = append(Clients, clients)
		}

	}

	numberOfAWS := len(servers)
	var channels []chan *rpc.Call
	for i := 0; i < numberOfAWS; i++ {
		channels = append(channels, make(chan *rpc.Call, 10))
	}
	newState := req.CurrentStates
	turns := 0
	if simiend {
		newState = currentState
		turns = Turn
		simiend = false
	}
	currentState = req.CurrentStates

	for turns < req.P.Turns && !end && !simiend {

		var requests []stubs.Request
		var responses []*stubs.Response

		for i := 0; i < numberOfAWS; i++ {
			requests = append(requests, stubs.Request{newState, req.P, 0, "", numberOfAWS, i * int(req.P.ImageHeight/numberOfAWS), i + 1, ""})
			responses = append(responses, new(stubs.Response))
		}
		channels = nil
		for i := 0; i < numberOfAWS; i++ {
			channels = append(channels, make(chan *rpc.Call, 10))
		}
		for i := 0; i < numberOfAWS; i++ {
			makeCall(Clients[i], requests[i], responses[i], channels[i])
		}

		mutex.Lock()
		newState = nil
		for i := 0; i < numberOfAWS; i++ {
			select {
			case <-channels[i]:
				newState = append(newState, responses[i].NewState...)
			}
		}

		if len(newState) != req.P.ImageHeight {

			IpAddresses = nil
			for i := range Clients {
				Clients[i].Call(stubs.GameOfLifeSend, new(stubs.Response), new(stubs.Request))
			}
			numberOfAWS = len(IpAddresses)
			servers = nil
			if numberOfAWS == 0 {
				fmt.Println("Error , all servers disconnected ")
				return err
			}
			fmt.Println("server disconnected, continue with " + strconv.Itoa(numberOfAWS) + " servers")
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
		} else if len(IpAddresses) > numberOfAWS {

			numberOfAWS = len(IpAddresses)
			servers = nil

			fmt.Println("server back, continue with " + strconv.Itoa(numberOfAWS) + " servers")
			for i := range IpAddresses {
				servers = append(servers, flag.String("server-"+strconv.Itoa(i)+"-"+"-"+strconv.Itoa(x)+"-"+strconv.Itoa(y), IpAddresses[i]+":8030", "IP:port string to connect to as server"))
			}
			y++
			Clients = nil
			for i := range servers {
				clients, err := rpc.Dial("tcp", *servers[i])
				//Clients = append(Clients, clients)
				if err != nil {
					servers = append(servers[:i], servers[i+1:]...)
					IpAddresses = append(IpAddresses[:i], IpAddresses[i+1:]...)
				} else {
					Clients = append(Clients, clients)
				}

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
