package graph

import "testing"

func TestLinePointGraph(t *testing.T) {
	name := "折线图"
	xName := "时间戳"
	yName := "收益"
	xs := []float64{1, 2, 3, 4, 5, 6}
	ys := []float64{1.1, 1,4, 1.3, 1.8, 1.9, 2.1}
	LinePointGraph(name, xName, yName, xs, ys)
}
