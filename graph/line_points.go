package graph

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)


type XY struct {

}

//折线图
func LinePointGraph( name string, xName string, yName string, xs []float64, ys []float64) {
	p, _ := plot.New()
	p.Title.Text = name
	p.X.Label.Text = xName
	p.Y.Label.Text = yName
	point := plotter.XYs{}
	for i, x := range xs {
		xy := plotter.XY{
			X: x,
			Y: ys[i],
		}
		point = append(point, xy)
	}

	plotutil.AddLinePoints(p, point)
	p.Save(4*vg.Inch, 4*vg.Inch, "/Users/zhengtao/Desktop/" + name + ".png")
}
