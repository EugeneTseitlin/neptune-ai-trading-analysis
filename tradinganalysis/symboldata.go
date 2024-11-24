package tradinganalysis

import (
	"math"

	"github.com/shopspring/decimal"
)


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


func DecComparator(a, b interface{}) int {
	aDec, bDec := a.(decimal.Decimal), b.(decimal.Decimal)
	return aDec.Compare(bDec)
}

func (symbolData *SymbolData) addBatch(newPricePoints []decimal.Decimal) {

	previousPricePointsSize := len(symbolData.pricePoints)
	symbolData.pricePoints = append(symbolData.pricePoints, newPricePoints...)

	for windowTargetSize, statsWindow := range symbolData.statWindowsPerSize {

		prevWindowEffectiveSize := min(windowTargetSize, previousPricePointsSize)
		nextWindowEffectiveSize := min(windowTargetSize, len(symbolData.pricePoints))

		prevWindowFirstIndex := len(symbolData.pricePoints) - prevWindowEffectiveSize - len(newPricePoints)

		pricePointsLeavingWindowSize := max(0, prevWindowEffectiveSize + len(newPricePoints) - windowTargetSize)
		pricePointsLeavingWindowLastIndex := prevWindowFirstIndex + pricePointsLeavingWindowSize
		pricePointsLeavingWindow := symbolData.pricePoints[prevWindowFirstIndex:pricePointsLeavingWindowLastIndex]

		pricePointsEnteringWindowSize := min(windowTargetSize, len(newPricePoints))
		pricePointsEnteringWindowFirstIndex := len(symbolData.pricePoints) - pricePointsEnteringWindowSize
		pricePointsEnteringWindow := symbolData.pricePoints[pricePointsEnteringWindowFirstIndex:]

		statsWindow.RemoveSliceFromTreeMap(pricePointsLeavingWindow)
		statsWindow.InsertSliceToTreeMap(pricePointsEnteringWindow)

		nextWindowEffectiveSizeDec := decimal.NewFromInt(int64(nextWindowEffectiveSize))
		prevWindowEffectiveSizeDec := decimal.NewFromInt(int64(prevWindowEffectiveSize))
		
		prevAverage := statsWindow.Average
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
