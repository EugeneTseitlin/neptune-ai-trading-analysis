package main

import (
	// "fmt"
	"math"
	"sync"
)

type TradeAnalysisEngine struct {
	symbolDataPerSymbol map[string]*SymbolData
	locksPerSymbol map[string]*sync.RWMutex
	rootLock sync.RWMutex
}

func NewTradeAnalysisEngine() *TradeAnalysisEngine {
	return &TradeAnalysisEngine{
		symbolDataPerSymbol: make(map[string]*SymbolData),
		locksPerSymbol: make(map[string]*sync.RWMutex),
	}
}

type SymbolData struct {
	pricePoints []float64
	statWindowsPerSize map[int]*StatWindow
}

// func (sd SymbolData) Print() {
// 	fmt.Println("Price points: %s", sd.pricePoints);
// 	for targetWindowSize, window := range sd.statWindowsPerSize {
// 		fmt.Println("Window with target size: %s", targetWindowSize)
// 		fmt.Println(window)
// 	}
// }

func NewSymbolData() *SymbolData {

	statWindowsPerSize := make(map[int]*StatWindow)
	for i := 1; i <= 8; i++ {
		statWindowsPerSize[int(math.Pow10(i))] = &StatWindow{}
	}

	return &SymbolData{
		pricePoints: []float64{},
		statWindowsPerSize: statWindowsPerSize,
	}
}

func (symbolData *SymbolData) addBatch(newPricePoints []float64) {
	
	previousPricePointsSize := len(symbolData.pricePoints)
	symbolData.pricePoints = append(symbolData.pricePoints, newPricePoints...)

	for windowTargetSize, statWindow := range symbolData.statWindowsPerSize {

		prevWindowEffectiveSize := min(windowTargetSize, previousPricePointsSize)
		nextWindowEffectiveSize := min(windowTargetSize, len(symbolData.pricePoints))
		
		prevWindowFirstIndex := len(symbolData.pricePoints) - prevWindowEffectiveSize - len(newPricePoints)
		
		pricePointsLeavingWindowSize := max(0, prevWindowEffectiveSize + len(newPricePoints) - windowTargetSize)
		pricePointsLeavingWindowLastIndex := prevWindowFirstIndex + pricePointsLeavingWindowSize
		pricePointsLeavingWindow := symbolData.pricePoints[prevWindowFirstIndex:pricePointsLeavingWindowLastIndex]

		pricePointsEnteringWindowSize := min(windowTargetSize, len(newPricePoints))
		pricePointsEnteringWindowFirstIndex := len(symbolData.pricePoints) - pricePointsEnteringWindowSize
		pricePointsEnteringWindow := symbolData.pricePoints[pricePointsEnteringWindowFirstIndex:]
		
		prevAvgerage := statWindow.Average
		nextAverage := prevAvgerage * (float64(prevWindowEffectiveSize) / float64(nextWindowEffectiveSize)) + 
			(sumSlice(pricePointsEnteringWindow) - sumSlice(pricePointsLeavingWindow)) / float64(nextWindowEffectiveSize)

		prevSumOfSquares := statWindow.SumOfSquares
		sumOfSquaresOfPointsLeavingWindow := calcSumOfSquares(pricePointsLeavingWindow)
		sumOfSquaresOfPointsEnteringWindow := calcSumOfSquares(pricePointsEnteringWindow)
		nextSumOfSquares := prevSumOfSquares - sumOfSquaresOfPointsLeavingWindow + sumOfSquaresOfPointsEnteringWindow

		nextVariance := nextSumOfSquares / float64(nextWindowEffectiveSize) - nextAverage * nextAverage

		statWindow.Last = symbolData.pricePoints[len(symbolData.pricePoints) - 1]
		statWindow.Average = nextAverage
		statWindow.SumOfSquares = nextSumOfSquares
		statWindow.Variance = nextVariance
	}
}

func calcSumOfSquares(points []float64) float64 {
	var sumOfSquares float64
	for _, point := range points {
		squaredDeviation := point * point 
		sumOfSquares += squaredDeviation
	}
	return sumOfSquares
}

type Numeric interface {
	~int | ~int32 | ~int64 | ~float32 | ~float64 | ~uint
}

func sumSlice[T Numeric](numbers []T) T {
	var sum T
	for _, number := range numbers {
		sum += number
	}
	return sum
}

type StatWindow struct {
	Min float64
	Max float64
	Last float64
	Average float64
	SumOfSquares float64
	Variance float64
}

func (engine *TradeAnalysisEngine) AddBatch(symbol string, newPricePoints []float64) {
	engine.rootLock.RLock()
	_, exists := engine.symbolDataPerSymbol[symbol]
	engine.rootLock.RUnlock()
	if !exists {
		engine.rootLock.Lock()
		_, reallyExists := engine.symbolDataPerSymbol[symbol]
		if (!reallyExists) {
			engine.symbolDataPerSymbol[symbol] = NewSymbolData()
			engine.locksPerSymbol[symbol] = &sync.RWMutex{}
		}
		engine.rootLock.Unlock()
	}

	engine.locksPerSymbol[symbol].Lock()
	symbolData := engine.symbolDataPerSymbol[symbol]
	symbolData.addBatch(newPricePoints)
	engine.locksPerSymbol[symbol].Unlock()
}

func (engine *TradeAnalysisEngine) GetStats(symbol string, k int) StatWindow {
	engine.locksPerSymbol[symbol].Lock()
	defer engine.locksPerSymbol[symbol].Unlock()
	sd := engine.symbolDataPerSymbol[symbol]
	return *sd.statWindowsPerSize[int(math.Pow10(k))]
}
