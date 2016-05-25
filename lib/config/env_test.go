package config

import (
	"testing"
	"os"
)

func TestGetEnvMap(t *testing.T) {
	os.Setenv("HELLO", "WORLD")
	env := GetEnvMap()

	val, ok := env["HELLO"]
	if ok == false {
		t.Errorf("Key HELLO is not defined")
	}
	if val != "WORLD" {
		t.Errorf("Value of key HELLO is incorrect")
	}
}
