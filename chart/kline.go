package chart

import (
	"black_hole/model"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"net/http"
	"strconv"
	"log"
)

var kLine *charts.Kline
var profitKLine *charts.Line

func httpserver(w http.ResponseWriter, _ *http.Request) {
	profitKLine.Render(w)
}

func StartHttpServer(strategyName string, coinName string, dataElement []model.BackTestDataElement) {
	profitKLine = charts.NewLine()
	profitKLine.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeInfographic}),
		charts.WithTitleOpts(opts.Title{
			Title:    "黑洞基金 " + strategyName,
			Subtitle: coinName + "历史回测情况",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			SplitNumber: 20,
			Scale:       true,
		}),
		charts.WithYAxisOpts(opts.YAxis{
			SplitLine: &opts.SplitLine{
				Show: true,
			},
		}),
	)
	dates := make([]string, 0)
	profits := make([]opts.LineData, 0)
	for _, element := range dataElement {
		value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", element.TotalProfit), 64)
		dates = append(dates, element.Date)
		profits = append(profits, opts.LineData{
			Value: value,
		})
	}
	log.Printf("[StartHttpServer]: len: %v, dataElement: %v", len(dates), dates)
	log.Printf("[StartHttpServer]: len: %v, profits: %v", len(profits), profits)
	profitKLine.SetXAxis(dates).AddSeries("收益倍数", profits).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{
			Show: true,
		}),
		charts.WithLineChartOpts(
			opts.LineChart{
				Smooth: true,
			}),
		charts.WithAreaStyleOpts(
			opts.AreaStyle{
				Opacity: 0.2,
			}),
	)

	http.HandleFunc("/", httpserver)
	http.ListenAndServe(":8804", nil)
}

func StartBiasHttpServer(strategyName string, coinName string, dataElement []model.BiasDataElement) {
	profitKLine = charts.NewLine()
	profitKLine.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeInfographic}),
		charts.WithTitleOpts(opts.Title{
			Title:    "黑洞基金 " + strategyName,
			Subtitle: coinName + "历史回测情况",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			SplitNumber: 20,
			Scale:       true,
		}),
		charts.WithYAxisOpts(opts.YAxis{
			SplitLine: &opts.SplitLine{
				Show: true,
			},
		}),
	)
	dates := make([]string, 0)
	bias := make([]opts.LineData, 0)
	for _, element := range dataElement {
		value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", element.Bias), 64)
		dates = append(dates, element.Data)
		bias = append(bias, opts.LineData{
			Value: value * 100,
		})
	}
	log.Printf("[StartHttpServer]: len: %v, dataElement: %v", len(dates), dates)
	log.Printf("[StartHttpServer]: len: %v, profits: %v", len(bias), bias)
	profitKLine.SetXAxis(dates).AddSeries("乖离率", bias).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{
			Show: true,
		}),
		charts.WithLineChartOpts(
			opts.LineChart{
				Smooth: true,
			}),
		charts.WithAreaStyleOpts(
			opts.AreaStyle{
				Opacity: 0.2,
			}),
	)

	http.HandleFunc("/", httpserver)
	http.ListenAndServe(":8804", nil)
}

