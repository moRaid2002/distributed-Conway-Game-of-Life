package stubs

import "uk.ac.bris.cs/gameoflife/gol"

var GameOfLifeHandler = "GameOfLife.EvaluateBoard"

type Response struct {
	NewState [][]byte
}

type Request struct {
	CurrentStates [][]byte
	P             gol.Params
}
