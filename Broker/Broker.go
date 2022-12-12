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

var index1 = 1
var index2 = 1
var IpAddresses []string

// calls all servers
func makeCall(client *rpc.Client, req stubs.Request, res *stubs.Response, channel chan *rpc.Call) {
	client.Go(stubs.GameOfLifeHandler, req, res, channel) // calling the server.

}

// stops all servers
func StopAll(client *rpc.Client) {
	req := new(stubs.Request)
	res := new(stubs.Response)
	client.Call(stubs.GameOfLifeStop, req, res) // go to StopAll function to shut down the server cleanly.

}

type Broker struct{}

// global variables to store state
var currentState [][]byte
var Turn = 0
var index = 0
var end = false
var stop = false
var simiend = false

var mutex = sync.Mutex{}

// adds the ip address of a server
func AddIp(str string) {
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

// reports the live view of the current state
func (s *Broker) LiveView(req stubs.Request, res *stubs.Response) (err error) {
	if req.P.ImageHeight == len(currentState) {
		mutex.Lock() // mutex lock is needed to avoid the mismatch when we process state and turn and then sending using cellFlipped event channel.
		res.NewState = currentState
		res.Turn = Turn
		mutex.Unlock()
	}
	return
}

// processes all key presses sent from client
func (s *Broker) KeyPress(req stubs.Request, res *stubs.Response) (err error) {
	switch req.Keypress {
	case "s":
		res.NewState = currentState
		res.Turn = Turn
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

// calculates alive cells at a certain turn
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

// main function that processes all the turns
func (s *Broker) Client(req stubs.Request, res *stubs.Response) (err error) {

	currentState = req.CurrentStates

	if req.P.Turns == 0 { // if there is nothing to do then just return the currentstate.
		currentState = req.CurrentStates
		res.NewState = req.CurrentStates

		return
	}

	var servers []*string
	var Clients []*rpc.Client // add the client to the array.

	for i := range IpAddresses {
		servers = append(servers, flag.String("server-"+strconv.Itoa(i)+"-"+strconv.Itoa(index1), IpAddresses[i]+":8030", "IP:port string to connect to as server"))
	}

	index1++ // just to make different number of IP address
	flag.Parse()
	for i := range servers {
		clients, err := rpc.Dial("tcp", *servers[i]) // This line create clients.

		if err != nil {
			fmt.Println("server disconnected ")
			servers = append(servers[:i], servers[i+1:]...)             // if it is an error then remove the server from the list.
			IpAddresses = append(IpAddresses[:i], IpAddresses[i+1:]...) // remove the IP address.
		} else {
			Clients = append(Clients, clients) // otherwise, add the client to the array.
		}

	}

	numberOfAWS := len(servers)
	var channels []chan *rpc.Call
	for i := 0; i < numberOfAWS; i++ { //creating the channel which is same number of AWS nodes then make a array of channels.
		channels = append(channels, make(chan *rpc.Call, 10))
	}
	newState := req.CurrentStates
	turns := 0
	if simiend { // because ' simiend ' is global variable so when 'q' is pressed it stays true, so need to set it to false.
		newState = currentState
		turns = Turn
		simiend = false
	}
	currentState = req.CurrentStates

	for turns < req.P.Turns && !end && !simiend { // press 'q' or 'k' exit the loop.

		var requests []stubs.Request
		var responses []*stubs.Response

		for i := 0; i < numberOfAWS; i++ {
			requests = append(requests, stubs.Request{newState, req.P, 0, "", numberOfAWS, i * int(req.P.ImageHeight/numberOfAWS), i + 1, ""})
			responses = append(responses, new(stubs.Response)) //empty now, just starting
		}
		channels = nil // to call the server again.
		for i := 0; i < numberOfAWS; i++ {
			channels = append(channels, make(chan *rpc.Call, 10))
		}
		for i := 0; i < numberOfAWS; i++ {
			makeCall(Clients[i], requests[i], responses[i], channels[i]) // calling the servers
		}

		mutex.Lock()
		newState = nil
		for i := 0; i < numberOfAWS; i++ {
			select {
			case <-channels[i]:
				newState = append(newState, responses[i].NewState...) // append the response got from server, and check the number of the server. line 159 (server.go)
			}
		}
		//---------- Fault tolerance -------//
		if len(newState) != req.P.ImageHeight { // if one sever shut down, response will be empty

			IpAddresses = nil // if that is the case set IP address to 0 (empty) and
			for i := range Clients {
				Clients[i].Call(stubs.GameOfLifeSend, new(stubs.Response), new(stubs.Request))
			} // call sever function reset to process the work. then go to SendIP functions  then call AddIpServer and AddIp functions to process to get fresh IP Address
			numberOfAWS = len(IpAddresses)
			servers = nil
			if numberOfAWS == 0 { // if there is no AWS nodes are running then there is nothing.
				fmt.Println("Error , all servers disconnected ")
				return err
			}
			fmt.Println("server disconnected, continue with " + strconv.Itoa(numberOfAWS) + " servers")
			for i := range IpAddresses { // carry on working with the number of the AWS nodes
				servers = append(servers, flag.String("server-"+strconv.Itoa(i)+"-"+"-"+strconv.Itoa(index1)+"-"+strconv.Itoa(index2), IpAddresses[i]+":8030", "IP:port string to connect to as server"))
			}
			index2++
			Clients = nil // client has to be empty to update the client
			for i := range servers {
				clients, _ := rpc.Dial("tcp", *servers[i])
				Clients = append(Clients, clients) // create new client
			}

			newState = currentState // set the state to the previous one and redo the turn again. because of the one stopped server.
		} else if len(IpAddresses) > numberOfAWS {

			numberOfAWS = len(IpAddresses) // update the number of the IP address
			servers = nil                  // we set server is empty to create new number of the server according to the number of IP address

			fmt.Println("server back, continue with " + strconv.Itoa(numberOfAWS) + " servers")
			for i := range IpAddresses { // carry on working with the number of the AWS nodes
				servers = append(servers, flag.String("server-"+strconv.Itoa(i)+"-"+"-"+strconv.Itoa(index1)+"-"+strconv.Itoa(index2), IpAddresses[i]+":8030", "IP:port string to connect to as server"))
			}
			index2++
			Clients = nil // client has to be empty to update the client
			for i := range servers {
				clients, err := rpc.Dial("tcp", *servers[i]) // create new client
				//Clients = append(Clients, clients)
				if err != nil { // if there is an error then remove the server and IP address.
					servers = append(servers[:i], servers[i+1:]...)             // remove server
					IpAddresses = append(IpAddresses[:i], IpAddresses[i+1:]...) // remove IP address
				} else {
					Clients = append(Clients, clients) // create new client
				}

			}
			newState = currentState // // set the state to the previous one and redo the turn again. because of the one stopped server.
			//----------------
		} else { // if none of the server is disconnected or coming back then process as usual.

			turns++
			Turn = turns
			currentState = newState
		}
		mutex.Unlock()
	}
	if end { //' k' is pressed then become true then stop all the server.
		for i := 0; i < numberOfAWS; i++ {
			StopAll(Clients[i])
		}
		stop = true // then close the listener which is the Broker. down below.
	}
	for i := 0; i < numberOfAWS; i++ {
		Clients[i].Close()
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
