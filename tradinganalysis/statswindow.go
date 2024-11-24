package tradinganalysis

import (
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/shopspring/decimal"
)

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

func NewStatWindow() *StatsWindow {
	return &StatsWindow{
		pointsTreeMap: *treemap.NewWith(DecComparator),
	}
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
