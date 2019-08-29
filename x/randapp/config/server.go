package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
)

var configTemplate *template.Template

type RAServerConfig struct {
	ExampleMetric float64 `mapstructure:"example_metric"`
}

func init() {
	var err error
	if configTemplate, err = template.New("configFileTemplate").Parse(defaultConfigTemplate); err != nil {
		panic(err)
	}
}

func DefaultRAServerConfig() *RAServerConfig {
	return &RAServerConfig{
		ExampleMetric: 0.0,
	}
}

func WriteConfigFile(configFilePath string, config *RAServerConfig) {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, config); err != nil {
		panic(err)
	}

	MustWriteFile(configFilePath, buffer.Bytes(), 0644)
}

func ReadConfigFile(configFilePath string) (config *RAServerConfig, err error) {
	return nil, nil
}

func MustWriteFile(filePath string, contents []byte, mode os.FileMode) {
	err := ioutil.WriteFile(filePath, contents, mode)
	if err != nil {
		panic(fmt.Sprintf("MustWriteFile failed: %v", err))
	}
}

const defaultConfigTemplate = `# This is a marketplace server TOML config file.
# For more information, see https://github.com/toml-lang/toml

##### common marketplace server config options #####

# Example
#example_metric = "{{ .ExampleMetric }}"
`
