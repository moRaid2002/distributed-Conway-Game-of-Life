package stubs

import (
	"uk.ac.bris.cs/gameoflife/gol/subParams"
)

var GameOfLifeHandler = "GameOfLife.EvaluateBoard"
var GameOfLifeAlive = "GameOfLife.GetAlive"
var GameOfLifePress = "GameOfLife.Key"
var GameOfLifeLiveView = "GameOfLife.Out"
var GameOfLifeStop = "GameOfLife.Stop"
var GameOfLifeClientStop = "GameOfLife.StopClient"

var BrokerClient = "Broker.Client"

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
	Start         int
	End           int
}
