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

// js to resize bar on load and on resize to fit in their container
const actionsWithEchartInstance = `
 const myChart = %MY_ECHARTS%;

 window.onload = function ()
     {
         myChart.resize();
     };

 window.onresize = function ()
     {
         myChart.resize();
     };`

func ConvertChartToTemplComponent(chart Renderable) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return chart.Render(w)
	})
}

// generateBarItems generates bar items from an array of float64
func generateBarItems(data []float64) []opts.BarData {
	items := make([]opts.BarData, 0)
	for _, item := range data {
		items = append(items, opts.BarData{Value: item})
	}
	return items
}

// BarStack generates a stacked or non stacked bar chart
func Bar(counts models.ChartData, stacked bool) *charts.Bar {

	bar := charts.NewBar()

	// name x axis
	bar.SetXAxis(counts.AxisX)

	// add datapoints
	for i, dataArr := range counts.Data {
		bar.AddSeries(counts.AxisY[i], generateBarItems(dataArr))
	}

	// stack bars
	if stacked {
		bar.SetSeriesOptions(charts.WithBarChartOpts(opts.BarChart{
			Stack: "stackA",
		}))
	}

	// bar.AddJSFuncs(actionsWithEchartInstance)

	return bar
}
