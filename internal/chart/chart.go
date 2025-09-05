package chart

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// The charts all have a `Render(w io.Writer) error` method on them.
// That method is very similar to templ's Render method.
type Renderable interface {
	Render(w io.Writer) error
}

func ConvertChartToTemplComponent(chart Renderable) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return chart.Render(w)
	})
}

func generateBarItems(data []int) []opts.BarData {
	items := make([]opts.BarData, 0)
	for _, item := range data {
		items = append(items, opts.BarData{Value: item})
	}
	return items
}

func BarStack(counts models.ChartData) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "stack style",
		}),
	)

	// name x axis
	bar.SetXAxis(counts.AxisX)

	// add datapoints
	for i, dataArr := range counts.Data {
		bar.AddSeries(counts.AxisY[i], generateBarItems(dataArr))
	}

	// stack bars
	bar.SetSeriesOptions(charts.WithBarChartOpts(opts.BarChart{
		Stack: "stackA",
	}))

	return bar
}
