package config

import (
	"gopkg.in/yaml.v2"
	"reflect"
	"os"
	"io/ioutil"
	"fmt"
	"errors"
)

// define base config structure
type Config struct {
	Image     string   // image string without any version/tag
	Name      string   // running container name, e.g. "example-staging"
	Port      []string // list of ports to expose to
	Env       []string // list of environment variables set for container
	Mount     []string // list of mounts for container
	Bootstrap []string // list of commands to run after container start
	Upstream  []struct {
		Template string // upstream template
		Resource string // upstream file to write template to
		Command  string // command to gracefully reload upstream
	}
}

// read filename and return string data
func ReadFile(fileName string) string {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(CONFIGURATION_ERROR)
	}

	return string(data)
}

// parse config string and return Config struct
func ParseConfig(yml string, env map[string]string) Config {
	var cfg Config = Config{}

	// extract json
	err := yaml.Unmarshal([]byte(yml), &cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(CONFIGURATION_ERROR)
	}

	// expand environment variables
	expandConfig(&cfg, env)

	// validate config
	validateConfig(cfg)

	return cfg
}

// process config struct via reflection
func expandConfig(interfaceT interface{}, env map[string]string) {
	processReflection(reflect.ValueOf(interfaceT).Elem(), env)
}

// recursive reflection - check type and update
func processReflection(v reflect.Value, env map[string]string) {

	// iterate over struct metadata, find strings
	// and expand ${ENV} variables from provided keyMap
	for i := 0; i < v.NumField(); i++ {
		fieldMeta := v.Field(i)
		if (fieldMeta.CanSet() == false) {
			continue
		}

		// process every kind of type in freshly parsed config file
		switch (fieldMeta.Kind()) {

		// struct
		case reflect.Struct:
			processReflection(fieldMeta, env)
			break

		// string
		case reflect.String:
			fieldVal := reflect.ValueOf(fieldMeta.Interface())
			newVal := expandString(fieldVal.String(), env)
			fieldMeta.SetString(newVal)
			break

		// slice of strings
		case reflect.Slice:
			fieldVal := reflect.ValueOf(fieldMeta.Interface())
			for j := 0; j < fieldVal.Len(); j++ {
				if (fieldVal.Index(j).Kind() != reflect.String) {
					processReflection(fieldVal.Index(j), env)
				} else {
					newVal := expandString(fieldVal.Index(j).String(), env)
					fieldMeta.Index(j).SetString(newVal)
				}
			}
			break
		}
	}
}

// replace key in string with value defined in keyMap metadata
// in case of negative lookup - output error and exit
func expandString(str string, env map[string]string) string {
	result := os.Expand(str, func(varName string) string {
		if varVal, ok := env[varName]; ok {
			return varVal
		} else {
			fmt.Printf("Configuration error: Enironment variable \"%v\" is not set, but referenced.\n", varName)
			os.Exit(CONFIGURATION_ERROR)
		}
		return ""
	})

	return result
}

// validate configuration file for mandatory parameters
func validateConfig(cfg Config) error {
	err := false

	if cfg.Image == "" {
		fmt.Printf("Configuration error: image is not defined\n")
		err = true
	}

	if cfg.Name == "" {
		fmt.Println("Configuration error: name is not defined\n")
		err = true
	}

	if err == true {
		return errors.New("Configuration validation failed")
	}

	return nil;
}
