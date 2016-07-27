package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

const (
	DefaultBindAddr string = "tcp://localhost:8080"
	DefaultBaseDir  string = ""
)

type Config struct {
	BindAddr string `json:"bind"`
	BaseDir  string `json:"base_dir"`
}

func (config *Config) Load(filePath string) error {
	_, err := JSONLoad(filePath, &config)
	if err != nil {
		return err
	}

	return nil
}

func getSetting(config map[string]interface{}, setting string, kind reflect.Kind,
	mandatory bool, fallbackValue interface{}) (interface{}, error) {

	if _, ok := config[setting]; !ok {
		if mandatory {
			return fallbackValue, fmt.Errorf("missing mandatory setting `%s'", setting)
		}

		return fallbackValue, nil
	}

	configKind := reflect.ValueOf(config[setting]).Kind()
	if configKind != kind {
		return fallbackValue, fmt.Errorf("setting `%s' value should be of type %s and not %s", setting, kind.String(),
			configKind.String())
	}

	return config[setting], nil
}

// GetString returns the string value of a configuration setting.
func GetString(config map[string]interface{}, setting string, mandatory bool) (string, error) {
	value, err := getSetting(config, setting, reflect.String, mandatory, "")
	return value.(string), err
}

// GetInt returns the int value of a configuration setting.
func GetInt(config map[string]interface{}, setting string, mandatory bool) (int, error) {
	value, err := getSetting(config, setting, reflect.Float64, mandatory, 0.0)
	return int(value.(float64)), err
}

// GetFloat returns the float value of a configuration setting.
func GetFloat(config map[string]interface{}, setting string, mandatory bool) (float64, error) {
	value, err := getSetting(config, setting, reflect.Float64, mandatory, 0.0)
	return value.(float64), err
}

// GetBool returns the bool value of a configuration setting.
func GetBool(config map[string]interface{}, setting string, mandatory bool) (bool, error) {
	value, err := getSetting(config, setting, reflect.Bool, mandatory, false)
	return value.(bool), err
}

// GetStringSlice returns the string slice of a configuration setting.
func GetStringSlice(config map[string]interface{}, setting string, mandatory bool) ([]string, error) {
	value, err := getSetting(config, setting, reflect.Slice, mandatory, nil)
	if err != nil || value == nil {
		return nil, err
	}
	out := make([]string, 0)
	for _, v := range value.([]interface{}) {
		if reflect.ValueOf(v).Kind() != reflect.String {
			return nil, fmt.Errorf("setting `%s' should be slice of strings and not %s", setting,
				reflect.ValueOf(v).Kind().String())
		} else {
			out = append(out, v.(string))
		}
	}
	return out, nil
}

// GetStringMap returns the a string map of a configuration setting.
func GetStringMap(config map[string]interface{}, setting string, mandatory bool) (map[string]interface{}, error) {
	value, err := getSetting(config, setting, reflect.Map, mandatory, nil)
	return value.(map[string]interface{}), err
}

// GetJsonObj returns the JSON Object interface{} of a configuration setting.
func GetJsonObj(config map[string]interface{}, setting string, mandatory bool) (interface{}, error) {
	value, err := getSetting(config, setting, reflect.Map, mandatory, nil)
	return value, err
}

// GetJsonArray returns the JSON Array interface{} of a configuration setting.
func GetJsonArray(config map[string]interface{}, setting string, mandatory bool) (interface{}, error) {
	value, err := getSetting(config, setting, reflect.Slice, mandatory, nil)
	return value, err
}

// JSONLoad loads the JSON formatted data in result from the filesystem.
func JSONLoad(filePath string, result interface{}) (os.FileInfo, error) {
	// Load JSON data from file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, result); err != nil {
		return nil, jsonError(string(data), err)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return fileInfo, nil
}

func jsonError(data string, err error) error {
	syntax, ok := err.(*json.SyntaxError)
	if !ok {
		return err
	}

	lineStart := strings.LastIndex(data[:syntax.Offset], "\n")
	line, position := strings.Count(data[:syntax.Offset], "\n")+1, int(syntax.Offset)-lineStart-1

	return fmt.Errorf("%s (line: %d, pos: %d)", err, line, position)
}
