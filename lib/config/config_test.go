package config

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestConfigParse(t *testing.T) {

	configRawString := `
# base project in registry
image: registry.example.com:80/example/${PROJECT}
name: example-container

# environment variables to pass to container
env:
  - DB_HOST=${DATABASE_HOST}
  - DB_NAME=${DATABASE_NAME}

# port numbers to pass to container
# each port is available as %TCP_<port_number>% or %UDP_<port_number>
# by default, tcp is used.
port:
  - 80

# mounts
mount:
  - /local/path:/container/path

# commands to run right after container started
# but before upstream switched
bootstrap:
  - /bin/echo "Ok!"

# upstream templates for current container
upstream:
  # nginx tcp port 80 template
  - nginx:
    template: ${TEMPLATE_DATA}
    resource: ${TEMPLATE_FILE}
    command: ${PROXY_RELOAD_CMD}

  # haproxy tcp config
  - haproxy:
    template: example HAProxy config
    command: "sudo service haproxy reload"
`
	env := make(map[string]string)
	env["PROJECT"] = "project"
	env["DATABASE_HOST"] = "localhost"
	env["DATABASE_NAME"] = "test_db"
	env["TEMPLATE_DATA"] = `upstream node_sample {
	server http://127.0.0.1:%PORT_80%;
}
`
	env["TEMPLATE_FILE"] = "/etc/nginx/upstreams/node_sample_upstream.conf"
	env["PROXY_RELOAD_CMD"] = "sudo service nginx reload"
	config := ParseConfig(configRawString, env)

	// asserts
	assert.Equal(t, "registry.example.com:80/example/project", config.Image)
	assert.Equal(t, "DB_HOST=localhost", config.Env[0])
	assert.Equal(t, "DB_NAME=test_db", config.Env[1])
	assert.Equal(t, env["TEMPLATE_DATA"], config.Upstream[0].Template)
	assert.Equal(t, env["TEMPLATE_FILE"], config.Upstream[0].Resource)
	assert.Equal(t, "example HAProxy config", config.Upstream[1].Template)
	assert.Equal(t, "sudo service haproxy reload", config.Upstream[1].Command)
	assert.Equal(t, "/local/path:/container/path", config.Mount[0])
	assert.Equal(t, "/bin/echo \"Ok!\"", config.Bootstrap[0])
}
