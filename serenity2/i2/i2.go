package main

import (
	"os"

	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/serenity2/input"
)

////////////
//  main  //
////////////
func main() {

	meta := plugin.NewPluginMeta(
		"i2",
		1,
		plugin.CollectorPluginType,
		nil,
		nil,
		// []string{plugin.SnapGOBContentType},
		// []string{plugin.SnapGOBContentType},
		// optional options ???
		plugin.Unsecure(true),
		plugin.RoutingStrategy(plugin.DefaultRouting),
		plugin.CacheTTL(1100*time.Millisecond),
	)
	// ???
	meta.RPCType = plugin.JSONRPC

	plugin.Start(
		meta,
		input.Input{
			Namespace: []string{"serenity2", "input", "metric2"},
		},
		os.Args[1],
	)
}
