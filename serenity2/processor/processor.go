// base package for processor
package processor

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"

	"log"
)

// processor implements ProcessorPlugin
type Processor struct {
	// Process just one metric
	ProcessMetrics func(metrics []plugin.PluginMetricType) []plugin.PluginMetricType
}

// Process just passthrough content
func (p *Processor) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	log.Println("processor:Process called")

	metrics := []plugin.PluginMetricType{}
	switch contentType {
	case plugin.SnapGOBContentType:
		dec := gob.NewDecoder(bytes.NewBuffer(content))
		if err := dec.Decode(&metrics); err != nil {
			log.Printf("Error decoding: error=%v content=%v", err, content)
			return "", nil, err
		}
	default:
		log.Printf("Error unknown content type '%v'", contentType)
		return "", nil, errors.New(fmt.Sprintf("Unknown content type '%s'", contentType))
	}

	if p.ProcessMetrics != nil {
		log.Printf("PROCESS-METRICS: %#v\n", p.ProcessMetrics)

		metrics = p.ProcessMetrics(metrics)

		// encode
		b := &bytes.Buffer{}
		err := gob.NewEncoder(b).Encode(&metrics)
		if err != nil {
			log.Panicln("cannot encode metrics:", err)
		}
		content = b.Bytes()

	}

	for _, m := range metrics {
		log.Printf("PROCESSOR-DUMP: %v|%v|%v\n", m.Timestamp(), m.Namespace(), m.Data())
	}

	// passthrough
	return contentType, content, nil
}

// GetConfigPolicy returns empty policy
func (p *Processor) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return cpolicy.New(), nil
}
