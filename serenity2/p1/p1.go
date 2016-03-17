package main

import (
	"os"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/serenity2/processor"
)

// main
func main() {
	plugin.Start(
		plugin.NewPluginMeta(
			"p1", //name
			1,    //version
			plugin.ProcessorPluginType,
			nil,
			nil,
			// []string{plugin.SnapGOBContentType},
			// []string{plugin.SnapGOBContentType},
		),
		&processor.Processor{},
		os.Args[1],
	)
}
