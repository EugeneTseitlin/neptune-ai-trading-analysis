package main

import (
	// "fmt"
	"math"
	"github.com/shopspring/decimal"
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
	pricePoints []decimal.Decimal
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
		pricePoints: []decimal.Decimal{},
		statWindowsPerSize: statWindowsPerSize,
	}
}

func (symbolData *SymbolData) addBatch(newPricePoints []decimal.Decimal) {
	
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
		
		prevAverage := statWindow.Average
		nextWindowEffectiveSizeDec := decimal.NewFromInt(int64(nextWindowEffectiveSize))
		prevWindowEffectiveSizeDec := decimal.NewFromInt(int64(prevWindowEffectiveSize))
		prevAverageAdjustedToNextWindowSize := prevAverage.Mul(prevWindowEffectiveSizeDec.Div(nextWindowEffectiveSizeDec))
		averageChangeFromSlidingWindow := sumSlice(pricePointsEnteringWindow).Sub(sumSlice(pricePointsLeavingWindow)).Div(nextWindowEffectiveSizeDec)
		nextAverage := prevAverageAdjustedToNextWindowSize.Add(averageChangeFromSlidingWindow)

		prevSumOfSquares := statWindow.SumOfSquares
		sumOfSquaresOfPointsLeavingWindow := calcSumOfSquares(pricePointsLeavingWindow)
		sumOfSquaresOfPointsEnteringWindow := calcSumOfSquares(pricePointsEnteringWindow)
		nextSumOfSquares := prevSumOfSquares.Sub(sumOfSquaresOfPointsLeavingWindow).Add(sumOfSquaresOfPointsEnteringWindow) 

		nextVariance := nextSumOfSquares.Div(nextWindowEffectiveSizeDec).Sub(nextAverage.Mul(nextAverage))

		statWindow.Last = symbolData.pricePoints[len(symbolData.pricePoints) - 1]
		statWindow.Average = nextAverage
		statWindow.SumOfSquares = nextSumOfSquares
		statWindow.Variance = nextVariance
	}
}

func calcSumOfSquares(points []decimal.Decimal) decimal.Decimal {
	var sumOfSquares decimal.Decimal
	for _, point := range points {
		squaredDeviation := point.Mul(point) 
		sumOfSquares = sumOfSquares.Add(squaredDeviation)
	}
	return sumOfSquares
}

func sumSlice(numbers []decimal.Decimal) decimal.Decimal {
	var sum decimal.Decimal
	for _, number := range numbers {
		sum = sum.Add(number)
	}
	return sum
}

type StatWindow struct {
	Min decimal.Decimal
	Max decimal.Decimal
	Last decimal.Decimal
	Average decimal.Decimal
	SumOfSquares decimal.Decimal
	Variance decimal.Decimal
}

func (engine *TradeAnalysisEngine) AddBatch(symbol string, newPricePoints []decimal.Decimal) {
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
