package main

import (
	"os"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/serenity2/output"
)

func main() {
	plugin.Start(
		plugin.NewPluginMeta(
			"o2",
			1,
			plugin.PublisherPluginType,
			nil,
			nil,
			// []string{plugin.SnapGOBContentType},
			// []string{plugin.SnapGOBContentType},
		),
		&output.Output{},
		os.Args[1],
	)
}
