package main

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
)

func TestEgineAddBatch(t *testing.T) {
	engine := NewTradeAnalysisEngine()
	
	symbol1 := "MILK"
	batch1 := ConvertFloatSliceToDec([]float64{10, 12, 11, 11.5, 13})
	batch2 := ConvertFloatSliceToDec([]float64{15, 16, 16.5})
	batch3 := ConvertFloatSliceToDec([]float64{18, 19, 21, 20.5, 22, 22.5, 21.5})
	
	expectWindow1 := StatWindow{
		Last: decimal.NewFromFloat(13),
		Min: decimal.NewFromFloat(10),
		Max: decimal.NewFromFloat(13),
		Average: decimal.NewFromFloat(11.5),
		SumOfSquares: decimal.NewFromFloat(666.25),
		Variance: decimal.NewFromFloat(11),
	}

	expectWindow2 := StatWindow{
		Last: decimal.NewFromFloat(16.5),
		Min: decimal.NewFromFloat(10),
		Max: decimal.NewFromFloat(16.5),
		Average: decimal.NewFromFloat(13.125),
		SumOfSquares: decimal.NewFromFloat(1419.5),
		Variance: decimal.NewFromFloat(5.171875),
	}

	expectWindow3 := StatWindow{
		Last: decimal.NewFromFloat(21.5),
		Min: decimal.NewFromFloat(15),
		Max: decimal.NewFromFloat(22.5),
		Average: decimal.NewFromFloat(19.2),
		SumOfSquares: decimal.NewFromFloat(3752),
		Variance: decimal.NewFromFloat(6.56),
	}

	engine.AddBatch(symbol1, batch1)
	window1 := engine.GetStats(symbol1, 1)
	fmt.Println(window1)
	Assert(t, window1, expectWindow1)

	engine.AddBatch(symbol1, batch2)
	window2 := engine.GetStats(symbol1, 1)
	fmt.Println(window2)
	Assert(t, window2, expectWindow2)

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

func ConvertFloatSliceToDec(numbers []float64) []decimal.Decimal {
	result := []decimal.Decimal{}
	for _, n := range numbers {
		result = append(result, decimal.NewFromFloat(n))
	}
	return result
}