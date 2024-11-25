package tradinganalysis

import (
	"errors"
	"math"
	"sync"

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

var ErrSymbolNotFound = errors.New("symbol not found")

func (engine *TradeAnalysisEngine) GetStats(symbol string, k int) (Stats, error) {
	sd, exists := engine.symbolDataPerSymbol[symbol]
	if exists {
		engine.locksPerSymbol[symbol].RLock()
		defer engine.locksPerSymbol[symbol].RUnlock()
		return sd.statWindowsPerSize[int(math.Pow10(k))].Stats, nil
	}
	return Stats{}, ErrSymbolNotFound
}
