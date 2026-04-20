package runningconfig

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/beevik/etree"
	"github.com/sdcio/config-server/apis/config/v1alpha1"
	"github.com/sdcio/kubectl-sdc/pkg/client"
	"sigs.k8s.io/yaml"
)

// DataClient defines the subset of the data client used by runningconfig.
type DataClient interface {
	Connect(ctx context.Context) error
	GetIntent(ctx context.Context, format client.Format, datastoreName, intentName string) (client.Intent, error)
	Close() error
}

// ValidFormats lists all supported output formats.
var ValidFormats = []client.Format{
	client.FormatJSON,
	client.FormatJSONIETF,
	client.FormatXML,
	client.FormatXPath,
	client.FormatYAML,
}

// FormatListString returns a comma-separated string of valid formats.
func FormatListString() string {
	return strings.Join(ValidFormatStrings(), ", ")
}

// ValidFormatStrings returns the list of valid format strings.
func ValidFormatStrings() []string {
	formatted := make([]string, len(ValidFormats))
	for i, f := range ValidFormats {
		formatted[i] = string(f)
	}
	return formatted
}

// ParseFormat converts a format string to the internal format enum.
func ParseFormat(formatStr string) (client.Format, error) {
	switch client.Format(strings.ToLower(formatStr)) {
	case client.FormatJSON:
		return client.FormatJSON, nil
	case client.FormatJSONIETF:
		return client.FormatJSONIETF, nil
	case client.FormatXML:
		return client.FormatXML, nil
	case client.FormatXPath:
		return client.FormatXPath, nil
	case client.FormatYAML:
		return client.FormatYAML, nil
	default:
		return "", fmt.Errorf("invalid format %q, must be one of: %s", formatStr, FormatListString())
	}
}

// Run connects to the data server and fetches the running configuration for the target.
func Run(ctx context.Context, cl RunningConfigClient, namespace, target string, format client.Format) (string, error) {

	// if the format is YAML, we actually need to request JSON from the server and convert it ourselves, since the server doesn't support YAML natively
	reqFormat := format
	if reqFormat == client.FormatYAML {
		reqFormat = client.FormatJSON
	}

	// Fetch the running configuration from the server
	runningConfig, err := cl.GetRunningConfig(ctx, namespace, target, reqFormat)
	if err != nil {
		return "", fmt.Errorf("failed to get running config: %w", err)
	}

	// Format the output based on the requested format
	switch format {
	case client.FormatJSON, client.FormatJSONIETF:
		var formatted bytes.Buffer
		err = json.Indent(&formatted, []byte(runningConfig.Value), "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to format JSON: %w", err)
		}
		return formatted.String(), nil
	case client.FormatXML:
		doc := etree.NewDocument()

		if err := doc.ReadFromString(runningConfig.Value); err != nil {
			return "", fmt.Errorf("failed to parse XML: %w", err)
		}
		wrapInConfig(doc)
		return doc.WriteToString()
	case client.FormatYAML:
		yamlVal, err := yaml.JSONToYAML([]byte(runningConfig.Value))
		if err != nil {
			return "", fmt.Errorf("failed to convert JSON to YAML: %w", err)
		}
		return string(yamlVal), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

func wrapInConfig(xmlDoc *etree.Document) {
	// make sure we have a root element
	// Create a new root <config>
	root := etree.NewElement("config")

	// Move all top-level elements under <config>
	for _, el := range xmlDoc.ChildElements() {
		root.AddChild(el)
	}

	// Reset document root
	xmlDoc.SetRoot(root)
	// set indent
	xmlDoc.Indent(2)
}

type RunningConfigClient interface {
	GetRunningConfig(ctx context.Context, namespace string, name string, format client.Format) (*v1alpha1.TargetRunningConfig, error)
}
