package stubs

import (
	"uk.ac.bris.cs/gameoflife/gol/subParams"
)

var GameOfLifeHandler = "GameOfLife.EvaluateBoard"
var GameOfLifeAlive = "GameOfLife.GetAlive"
var GameOfLifePress = "GameOfLife.Key"

type Response struct {
	NewState [][]byte
	Alive    int
	Turn     int
}

type Request struct {
	CurrentStates *[][]byte
	P             subParams.Params
	Turn          int
	Keypress      string
}
