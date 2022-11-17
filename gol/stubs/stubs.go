package stubs

import (
	"sync"
	"uk.ac.bris.cs/gameoflife/gol/subParams"
)

var GameOfLifeHandler = "GameOfLife.EvaluateBoard"
var GameOfLifeAlive = "GameOfLife.GetAlive"

type Response struct {
	NewState [][]byte
	Alive    int
	Turn     int
}

type Request struct {
	CurrentStates *[][]byte
	P             subParams.Params
	CurrentState  [][]byte
	CurrentTurn   int
	Mutex         sync.Mutex
}
