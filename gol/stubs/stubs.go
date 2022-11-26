package stubs

import (
	"uk.ac.bris.cs/gameoflife/gol/subParams"
)

var GameOfLifeHandler = "GameOfLife.EvaluateBoard"
var GameOfLifeStop = "GameOfLife.StopAll"
var GameOfLifeSend = "GameOfLife.Reset"

var BrokerClient = "Broker.Client"
var BrokerAlive = "Broker.AliveCell"
var BrokerKeyPress = "Broker.KeyPress"
var BrokerLiveView = "Broker.LiveView"
var BrokerIp = "Broker.AddIpServer"

type Response struct {
	NewState      [][]byte
	PreviousState [][]byte
	Alive         int
	Turn          int
	Flag          bool
}

type Request struct {
	CurrentStates [][]byte
	P             subParams.Params
	Turn          int
	Keypress      string
	NumberAWS     int
	Offset        int
	Server        int
	Ip            string
}
