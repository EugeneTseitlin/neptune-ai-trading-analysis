package main

import (
	"fmt"
	"testing"
)

func TestEgineAddBatch(t *testing.T) {
	engine := NewTradeAnalysisEngine()
	
	symbol1 := "MILK"
	batch1 := []float64{10, 12, 11, 11.5, 13}
	batch2 := []float64{15, 16, 16.5}
	batch3 := []float64{18, 19, 21, 20.5, 22, 22.5, 21.5}
	
	// expectWindow1 := StatWindow{
	// 	Last: 13,
	// 	Min: 10,
	// 	Max: 13,
	// 	Average: 11.5,
	// 	SumOfSquares: 666.25,
	// 	Variance: 11,
	// }

	// expectWindow2 := StatWindow{
	// 	Last: 16.5,
	// 	Min: 10,
	// 	Max: 16.5,
	// 	Average: 13.125,
	// SumOfSquares: 1419.5,
	// 	Variance: 5.171875,
	// }

	expectWindow3 := StatWindow{
		Last: 21.5,
		Min: 10,
		Max: 22.5,
		Average: 19.2,
		SumOfSquares: 3752,
		Variance: 6.56,
	}

	engine.AddBatch(symbol1, batch1)
	window1 := engine.GetStats(symbol1, 1)
	fmt.Println(window1)
	// Assert(t, window1, expectWindow1)

	engine.AddBatch(symbol1, batch2)
	window2 := engine.GetStats(symbol1, 1)
	fmt.Println(window2)
	// Assert(t, window2, expectWindow2)

	engine.AddBatch(symbol1, batch3)
	window3 := engine.GetStats(symbol1, 1)
	fmt.Println(window3)
	Assert(t, window3, expectWindow3)

}

func Assert[T comparable](t *testing.T, actual, expected T) {
	if actual != expected {
		t.Fatalf("Actual result: %v is not equal to expected: %v", actual, expected)
	}
}