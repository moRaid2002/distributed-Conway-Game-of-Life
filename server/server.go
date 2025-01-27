package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"time"
	"uk.ac.bris.cs/gameoflife/gol/stubs"
	"uk.ac.bris.cs/gameoflife/gol/subParams"
)

var index = 0

// send ip address to the broker
func SendIp(str string) { // IP Address if the broker is added here to call broker by the following IP address.
	server := flag.String("broker"+strconv.Itoa(index), "44.205.16.223:8030", "IP:port string to connect to as server")
	flag.Parse()
	client, _ := rpc.Dial("tcp", *server) // calling the broker by the specific IP address above and call the function.
	defer client.Close()
	req := stubs.Request{nil, *new(subParams.Params), 0, "", 0, 0, 0, str}
	res := new(stubs.Response)
	client.Call(stubs.BrokerIp, req, res)
	index++ // number of the server has to be unique, so need to add index, most of the time.
}

// process one turn of Game Of Life
func gameOfLife(p subParams.Params, newWorld [][]byte, startX int, endX int) [][]byte {
	var aliveCell = 0
	nextState := make([][]byte, endX-startX)
	for i := 0; i < endX-startX; i++ {
		nextState[i] = make([]byte, p.ImageWidth)
	}

	for x := startX; x < endX; x++ {
		for y := 0; y < p.ImageWidth; y++ {

			aliveCell = 0
			right := x + 1
			left := x - 1
			up := y - 1
			down := y + 1
			if x == p.ImageWidth-1 { // These four if statements are expressing the cases which is outside of the size
				right = 0
			}
			if x == 0 {
				left = p.ImageWidth - 1
			}
			if y == p.ImageHeight-1 {
				down = 0
			}
			if y == 0 {
				up = p.ImageHeight - 1
			}
			if newWorld[right][y] == 255 { // These if statements are calculating the alive cell of reachable points.
				aliveCell++
			}
			if newWorld[left][y] == 255 {
				aliveCell++
			}
			if newWorld[x][up] == 255 {
				aliveCell++
			}
			if newWorld[x][down] == 255 {
				aliveCell++
			}
			if newWorld[right][up] == 255 {
				aliveCell++
			}
			if newWorld[left][up] == 255 {
				aliveCell++
			}
			if newWorld[right][down] == 255 {
				aliveCell++
			}
			if newWorld[left][down] == 255 {
				aliveCell++
			}

			if aliveCell < 2 && newWorld[x][y] == 255 { // Setting up the rule of game life if the cell is alive or dead.
				nextState[x-startX][y] = 0
			}

			if aliveCell > 3 && newWorld[x][y] == 255 {
				nextState[x-startX][y] = 0
			}

			if (aliveCell == 2 || aliveCell == 3) && newWorld[x][y] == 255 {
				nextState[x-startX][y] = 255
			}

			if aliveCell == 3 && newWorld[x][y] == 0 {
				nextState[x-startX][y] = 255
			}
		}
	}
	return nextState
}

// worker function
func worker(p subParams.Params, newWorld [][]byte, out chan<- [][]byte, startX int, endX int) {
	newState := gameOfLife(p, newWorld, startX, endX)
	out <- newState // Sending the game of life function through the channel.
}

var end = false

type GameOfLife struct{}

// stop the server
func (s *GameOfLife) StopAll(req stubs.Request, res *stubs.Response) (err error) {
	fmt.Println("stopping")
	end = true // this will make the server shut down cleanly. look at the code below shut the listener.close , when the end becomes true then shut the all the server.
	return
}

// send the ip adress to the broker again
func (s *GameOfLife) Reset(req stubs.Request, res *stubs.Response) (err error) {
	resp, _ := http.Get("https://ifconfig.me")

	ipv4, _ := ioutil.ReadAll(resp.Body)

	SendIp(string(ipv4))
	fmt.Println("sent again")
	return
}

// devide the work inbtween threads and run gameoflife
func (s *GameOfLife) EvaluateBoard(req stubs.Request, res *stubs.Response) (err error) {

	var chanels []chan [][]byte
	var newstate [][]byte
	for threads := 0; threads < req.P.Threads; threads++ {
		chanels = append(chanels, make(chan [][]byte))
	}
	x := 0
	if req.NumberAWS%2 != 0 && req.Server == req.NumberAWS {
		x = req.P.ImageWidth % req.NumberAWS
	}

	for threads := 0; threads < req.P.Threads; threads++ { // Loop through all the threads.
		if threads == req.P.Threads-1 { // According to the condition match run the go Routine.
			go worker(req.P, req.CurrentStates, chanels[threads], req.Offset+(threads)*int(req.P.ImageHeight/(req.P.Threads*req.NumberAWS)), req.Offset+(req.P.ImageHeight/req.NumberAWS)+x) // worker funbction create the slices and devide the work between each threads then run the game of life function.
		} else {
			go worker(req.P, req.CurrentStates, chanels[threads], req.Offset+(threads)*int(req.P.ImageHeight/(req.P.Threads*req.NumberAWS)), req.Offset+(threads+1)*int(req.P.ImageHeight/(req.P.Threads*req.NumberAWS)))
		}
	}
	for threads := 0; threads < req.P.Threads; threads++ {
		received := <-chanels[threads] // Receiving the thread and append them together.
		newstate = append(newstate, received...)
	}

	req.CurrentStates = newstate

	res.NewState = req.CurrentStates

	return
}

// main - send ip address
func main() {
	fmt.Println("working")
	res, _ := http.Get("https://ifconfig.me")

	ipv4, _ := ioutil.ReadAll(res.Body)

	SendIp(string(ipv4))
	fmt.Println("sent")
	pAddr := flag.String("port", "8030", "Port to listen on")

	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&GameOfLife{})
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
