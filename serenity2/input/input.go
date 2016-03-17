package input

import (
	"log"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
)

// Input collector plugin for workflow testing purposes.
type Input struct {
	Namespace []string
}

// CollectMetrics returns one metric (value)
func (i Input) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	log.Println("input: CollectMetrics called")

	return []plugin.PluginMetricType{
		plugin.PluginMetricType{
			Data_:      42,
			Namespace_: i.Namespace,
			Timestamp_: time.Now(),
		},
	}, nil

}

// GetMetricTypes returns one metric (namespace)
func (i Input) GetMetricTypes(cfg plugin.PluginConfigType) ([]plugin.PluginMetricType, error) {
	log.Println("input: GetMetricTypes called")
	return []plugin.PluginMetricType{
		plugin.PluginMetricType{
			Namespace_: i.Namespace,
		},
	}, nil
}

// GetConfigPolicy returns empty policy
func (Input) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return cpolicy.New(), nil
}
