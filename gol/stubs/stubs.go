package stubs

import (
	"uk.ac.bris.cs/gameoflife/gol/subParams"
)

var GameOfLifeHandler = "GameOfLife.EvaluateBoard"
var GameOfLifeAlive = "GameOfLife.GetAlive"

type Response struct {
	NewState [][]byte
}

type Request struct {
	CurrentStates *[][]byte
	P             subParams.Params
}
