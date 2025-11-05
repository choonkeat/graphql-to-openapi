package converter

import (
	"gopkg.in/yaml.v3"
)

// MarshalYAML converts an OpenAPI document to YAML
func MarshalYAML(doc *OpenAPIDocument) ([]byte, error) {
	return yaml.Marshal(doc)
}
