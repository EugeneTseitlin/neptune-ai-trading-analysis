package main

import (
	// "fmt"
	"math"
	"sync"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/shopspring/decimal"
)

type TradeAnalysisEngine struct {
	symbolDataPerSymbol map[string]*SymbolData
	locksPerSymbol      map[string]*sync.RWMutex
	rootLock            sync.RWMutex
}

func NewTradeAnalysisEngine() *TradeAnalysisEngine {
	return &TradeAnalysisEngine{
		symbolDataPerSymbol: make(map[string]*SymbolData),
		locksPerSymbol:      make(map[string]*sync.RWMutex),
	}
}

type SymbolData struct {
	pricePoints        []decimal.Decimal
	statWindowsPerSize map[int]*StatsWindow
}

func NewSymbolData() *SymbolData {

	statWindowsPerSize := make(map[int]*StatsWindow)
	for i := 1; i <= 8; i++ {
		statWindowsPerSize[int(math.Pow10(i))] = NewStatWindow()
	}

	return &SymbolData{
		pricePoints:        []decimal.Decimal{},
		statWindowsPerSize: statWindowsPerSize,
	}
}

func DecLessFunc(a, b decimal.Decimal) bool {
	return a.LessThan(b)
}

func DecComparator(a, b interface{}) int {
	aDec, bDec := a.(decimal.Decimal), b.(decimal.Decimal)
	return aDec.Compare(bDec)
}

func NewStatWindow() *StatsWindow {
	return &StatsWindow{
		pointsTreeMap: *treemap.NewWith(DecComparator),
	}
}

func (symbolData *SymbolData) addBatch(newPricePoints []decimal.Decimal) {

	previousPricePointsSize := len(symbolData.pricePoints)
	symbolData.pricePoints = append(symbolData.pricePoints, newPricePoints...)

	for windowTargetSize, statsWindow := range symbolData.statWindowsPerSize {

		prevWindowEffectiveSize := min(windowTargetSize, previousPricePointsSize)
		nextWindowEffectiveSize := min(windowTargetSize, len(symbolData.pricePoints))

		prevWindowFirstIndex := len(symbolData.pricePoints) - prevWindowEffectiveSize - len(newPricePoints)

		pricePointsLeavingWindowSize := max(0, prevWindowEffectiveSize+len(newPricePoints)-windowTargetSize)
		pricePointsLeavingWindowLastIndex := prevWindowFirstIndex + pricePointsLeavingWindowSize
		pricePointsLeavingWindow := symbolData.pricePoints[prevWindowFirstIndex:pricePointsLeavingWindowLastIndex]

		pricePointsEnteringWindowSize := min(windowTargetSize, len(newPricePoints))
		pricePointsEnteringWindowFirstIndex := len(symbolData.pricePoints) - pricePointsEnteringWindowSize
		pricePointsEnteringWindow := symbolData.pricePoints[pricePointsEnteringWindowFirstIndex:]

		statsWindow.RemoveSliceFromTreeMap(pricePointsLeavingWindow)
		statsWindow.InsertSliceToTreeMap(pricePointsEnteringWindow)

		prevAverage := statsWindow.Average
		nextWindowEffectiveSizeDec := decimal.NewFromInt(int64(nextWindowEffectiveSize))
		prevWindowEffectiveSizeDec := decimal.NewFromInt(int64(prevWindowEffectiveSize))
		prevAverageAdjustedToNextWindowSize := prevAverage.Mul(prevWindowEffectiveSizeDec.Div(nextWindowEffectiveSizeDec))
		averageChangeFromSlidingWindow := sumSlice(pricePointsEnteringWindow).Sub(sumSlice(pricePointsLeavingWindow)).Div(nextWindowEffectiveSizeDec)
		nextAverage := prevAverageAdjustedToNextWindowSize.Add(averageChangeFromSlidingWindow)

		prevSumOfSquares := statsWindow.SumOfSquares
		sumOfSquaresOfPointsLeavingWindow := calcSumOfSquares(pricePointsLeavingWindow)
		sumOfSquaresOfPointsEnteringWindow := calcSumOfSquares(pricePointsEnteringWindow)
		nextSumOfSquares := prevSumOfSquares.Sub(sumOfSquaresOfPointsLeavingWindow).Add(sumOfSquaresOfPointsEnteringWindow)

		nextVariance := nextSumOfSquares.Div(nextWindowEffectiveSizeDec).Sub(nextAverage.Mul(nextAverage))

		nextMin, _ := statsWindow.pointsTreeMap.Min()
		nextMax, _ := statsWindow.pointsTreeMap.Max()
		statsWindow.Min = nextMin.(decimal.Decimal)
		statsWindow.Max = nextMax.(decimal.Decimal)
		statsWindow.Last = symbolData.pricePoints[len(symbolData.pricePoints)-1]
		statsWindow.Average = nextAverage
		statsWindow.SumOfSquares = nextSumOfSquares
		statsWindow.Variance = nextVariance
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

type Stats struct {
	Min      decimal.Decimal
	Max      decimal.Decimal
	Last     decimal.Decimal
	Average  decimal.Decimal
	Variance decimal.Decimal
}

func (a Stats) Equal(b Stats) bool {
	return a.Min.Equal(b.Min) &&
		a.Max.Equal(b.Max) &&
		a.Last.Equal(b.Last) &&
		a.Average.Equal(b.Average) &&
		a.Variance.Equal(b.Variance)
} 

type StatsWindow struct {
	Stats
	SumOfSquares  decimal.Decimal
	pointsTreeMap treemap.Map
}

func (sw StatsWindow) GetS() Stats {
	return sw.Stats
}

func (sw *StatsWindow) InsertToTreeMap(point decimal.Decimal) {
	counter, exists := sw.pointsTreeMap.Get(point)
	if exists {
		sw.pointsTreeMap.Put(point, counter.(int)+1)
	} else {
		sw.pointsTreeMap.Put(point, 1)
	}
}

func (sw *StatsWindow) InsertSliceToTreeMap(points []decimal.Decimal) {
	for _, point := range points {
		sw.InsertToTreeMap(point)
	}
}

func (sw *StatsWindow) RemoveFromTreeMap(point decimal.Decimal) {
	counter, exists := sw.pointsTreeMap.Get(point)
	if exists {
		if counter.(int) > 1 {
			sw.pointsTreeMap.Put(point, counter.(int)-1)
		} else {
			sw.pointsTreeMap.Remove(point)
		}
	}
}

func (sw *StatsWindow) RemoveSliceFromTreeMap(points []decimal.Decimal) {
	for _, point := range points {
		sw.RemoveFromTreeMap(point)
	}
}

func (engine *TradeAnalysisEngine) AddBatch(symbol string, newPricePoints []decimal.Decimal) {
	engine.rootLock.RLock()
	_, exists := engine.symbolDataPerSymbol[symbol]
	engine.rootLock.RUnlock()
	if !exists {
		engine.rootLock.Lock()
		_, reallyExists := engine.symbolDataPerSymbol[symbol]
		if !reallyExists {
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

func (engine *TradeAnalysisEngine) GetStats(symbol string, k int) Stats {
	engine.locksPerSymbol[symbol].Lock()
	defer engine.locksPerSymbol[symbol].Unlock()
	sd := engine.symbolDataPerSymbol[symbol]
	return sd.statWindowsPerSize[int(math.Pow10(k))].Stats
}
