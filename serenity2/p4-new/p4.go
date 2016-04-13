package main

import (
	"log"
	"os"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/serenity2/processor"
)

// main
func main() {
	plugin.Start(
		plugin.NewPluginMeta(
			"p4-new", //name
			1,        //version
			plugin.ProcessorPluginType,
			nil,
			nil,
			// []string{plugin.SnapGOBContentType},
			// []string{plugin.SnapGOBContentType},
		),
		&processor.Processor{
			ProcessMetrics: func(metrics []plugin.PluginMetricType) []plugin.PluginMetricType {
				// add new metrics 43

				log.Printf("new: %#v\n", 43)
				new := plugin.PluginMetricType{
					Data_:      43,
					Namespace_: []string{"serenity2", "process", "metric-p4-new"},
					Timestamp_: time.Now(),
				}

				metrics = append(metrics, new)
				return metrics
			},
		},
		os.Args[1],
	)
}
