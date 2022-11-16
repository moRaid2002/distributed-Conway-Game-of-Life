package stubs

import "uk.ac.bris.cs/gameoflife/gol/subParams"

var GameOfLifeHandler = "GameOfLife.EvaluateBoard"

type Response struct {
	NewState [][]byte
	//Out      chan int
}

type Request struct {
	CurrentStates *[][]byte
	P             subParams.Params
}
