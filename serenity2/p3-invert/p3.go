package main

import (
	"log"
	"os"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/serenity2/processor"
)

// main
func main() {
	plugin.Start(
		plugin.NewPluginMeta(
			"p3-invert", //name
			1,           //version
			plugin.ProcessorPluginType,
			nil,
			nil,
			// []string{plugin.SnapGOBContentType},
			// []string{plugin.SnapGOBContentType},
		),
		&processor.Processor{
			ProcessMetrics: func(metrics []plugin.PluginMetricType) []plugin.PluginMetricType {

				for i, m := range metrics {

					// invert int (40 becomes -40)
					// assert is int
					if v, ok := m.Data_.(float64); ok {
						log.Printf("inverted: %#v\n", -v)
						m.AddData(-v) // m.Data_ = -v
					} else {
						log.Printf("cannot type assert to float64: %#v (type=%T)", m.Data_, m.Data_)
					}
					metrics[i] = m
				}

				return metrics
			},
		},
		os.Args[1],
	)
}
