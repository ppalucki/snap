package output

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"
)

type Output struct{}

func (o *Output) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	log.Println("output:Publish called")
	metrics := []plugin.PluginMetricType{}

	// can I get unknown type !!!!!!!!!!!! ?
	switch contentType {
	case plugin.SnapGOBContentType:
		dec := gob.NewDecoder(bytes.NewBuffer(content))
		if err := dec.Decode(&metrics); err != nil {
			log.Printf("Error decoding: error=%v content=%v", err, content)
			return err
		}
	default:
		log.Printf("Error unknown content type '%v'", contentType)
		return errors.New(fmt.Sprintf("Unknown content type '%s'", contentType))
	}

	for _, m := range metrics {
		log.Printf("PUBLISHER-DUMP: %v|%v|%v\n", m.Timestamp(), m.Namespace(), m.Data())
	}

	return nil
}

func (f *Output) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return cpolicy.New(), nil
}
