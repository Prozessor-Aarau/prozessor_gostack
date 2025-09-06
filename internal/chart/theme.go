package chart

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func setNordTheme(chart any) {
	switch c := chart.(type) {
	case *charts.Bar:
		c.SetGlobalOptions(
			charts.WithInitializationOpts(opts.Initialization{
				BackgroundColor: "transparent", // Transparent background
			}),
			charts.WithTitleOpts(opts.Title{
				TitleStyle: &opts.TextStyle{
					Color: "#2e3440", // Dark text for light background
				},
			}),
			charts.WithTooltipOpts(opts.Tooltip{
				Show:            opts.Bool(true),
				BackgroundColor: "#eceff4", // Light background for tooltip
				BorderColor:     "#4c566a",
			}),
			charts.WithLegendOpts(opts.Legend{
				TextStyle: &opts.TextStyle{
					Color: "#3b4252", // Dark text
				},
			}),
			charts.WithXAxisOpts(opts.XAxis{
				AxisLabel: &opts.AxisLabel{
					Color: "#3b4252",
				},
				AxisLine: &opts.AxisLine{
					LineStyle: &opts.LineStyle{
						Color: "#4c566a",
					},
				},
			}),
			charts.WithYAxisOpts(opts.YAxis{
				AxisLabel: &opts.AxisLabel{
					Color: "#3b4252",
				},
				AxisLine: &opts.AxisLine{
					LineStyle: &opts.LineStyle{
						Color: "#4c566a",
					},
				},
				SplitLine: &opts.SplitLine{
					LineStyle: &opts.LineStyle{
						Color: "#d8dee9",
					},
				},
			}),
			charts.WithColorsOpts(opts.Colors{
				"#5e81ac", "#81a1c1", "#88c0d0", "#8fbcbb",
				"#bf616a", "#d08770", "#ebcb8b", "#a3be8c",
				"#b48ead", "#4c566a", "#2e3440",
			}),
		)
	case *charts.Pie:
		c.SetGlobalOptions(
			charts.WithInitializationOpts(opts.Initialization{
				BackgroundColor: "transparent",
			}),
			charts.WithTitleOpts(opts.Title{
				TitleStyle: &opts.TextStyle{
					Color: "#2e3440",
				},
			}),
			charts.WithLegendOpts(opts.Legend{
				TextStyle: &opts.TextStyle{
					Color: "#3b4252",
				},
				Left:   "80%",
				Orient: "vertical",
			}),
			charts.WithColorsOpts(opts.Colors{
				"#5e81ac", "#81a1c1", "#88c0d0", "#8fbcbb",
				"#bf616a", "#d08770", "#ebcb8b", "#a3be8c",
				"#b48ead", "#4c566a", "#2e3440",
			}),
		)
		c.SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Show:      opts.Bool(true),
				Color:     "#2e3440", // Dark text for light background
				Formatter: "{b}: {c}",
			}),
		)
	}
}

func setDraculaTheme(chart any) {
	switch c := chart.(type) {
	case *charts.Bar:
		c.SetGlobalOptions(
			charts.WithInitializationOpts(opts.Initialization{
				BackgroundColor: "transparent", // Transparent background
			}),
			charts.WithTitleOpts(opts.Title{
				TitleStyle: &opts.TextStyle{
					Color: "#f8f8f2", // Dracula foreground
				},
			}),
			charts.WithTooltipOpts(opts.Tooltip{
				Show:            opts.Bool(true),
				BackgroundColor: "#282a36", // Solid dark background for tooltip
				BorderColor:     "#6272a4",
			}),
			charts.WithLegendOpts(opts.Legend{
				TextStyle: &opts.TextStyle{
					Color: "#f8f8f2",
				},
			}),
			charts.WithXAxisOpts(opts.XAxis{
				AxisLabel: &opts.AxisLabel{
					Color: "#f8f8f2",
				},
				AxisLine: &opts.AxisLine{
					LineStyle: &opts.LineStyle{
						Color: "#6272a4",
					},
				},
			}),
			charts.WithYAxisOpts(opts.YAxis{
				AxisLabel: &opts.AxisLabel{
					Color: "#f8f8f2",
				},
				AxisLine: &opts.AxisLine{
					LineStyle: &opts.LineStyle{
						Color: "#6272a4",
					},
				},
				SplitLine: &opts.SplitLine{
					LineStyle: &opts.LineStyle{
						Color: "#44475a",
					},
				},
			}),
			charts.WithColorsOpts(opts.Colors{
				"#ff79c6", "#bd93f9", "#8be9fd", "#50fa7b",
				"#ffb86c", "#ff5555", "#f1fa8c", "#6272a4",
				"#e91e63", "#9c27b0", "#00bcd4",
			}))
	case *charts.Pie:
		c.SetGlobalOptions(
			charts.WithInitializationOpts(opts.Initialization{
				BackgroundColor: "transparent",
			}),
			charts.WithTitleOpts(opts.Title{
				TitleStyle: &opts.TextStyle{
					Color: "#f8f8f2",
				},
			}),
			charts.WithLegendOpts(opts.Legend{
				TextStyle: &opts.TextStyle{
					Color: "#f8f8f2",
				},
				Left:   "80%",
				Orient: "vertical",
			}),
			charts.WithColorsOpts(opts.Colors{
				"#ff79c6", "#bd93f9", "#8be9fd", "#50fa7b",
				"#ffb86c", "#ff5555", "#f1fa8c", "#6272a4",
				"#e91e63", "#9c27b0", "#00bcd4",
			})) // Set label options for pie chart
		c.SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Show:      opts.Bool(true),
				Color:     "#f8f8f2", // Dracula foreground color
				Formatter: "{b}: {c}",
			}),
		)
	}
}
