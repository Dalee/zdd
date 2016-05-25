package release

import (
	"zdd/lib/config"
	"github.com/docker/engine-api/client"
)

// holds upstream template data
type BuildPortBinding struct {
	PortKey    string // port placeholder, e.g. %TCP_80%
	Port       string // acquired port number, e.g 32768
	AddressKey string // %TCP_80_IP%
	Address    string // Ip address port is bound
}

// TBD. metadata with mounts for build
type BuildMountBinding struct {

}

// holds
type BuildMetadata struct {
	docker        *client.Client
	cfg           config.Config
	ports         []BuildPortBinding
	mounts        []BuildMountBinding
	rollback      bool

	ImageTag      string `json:"imageTag"`
	ImageRef      string `json:"imageRef"`
	ImageId       string `json:"imageId"`
	ContainerName string `json:"containerName"`
	ContainerId   string `json:"containerId"`
}

// create new deploy
func CreateNew(cfg config.Config, version string) (*BuildMetadata, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	// create new build
	build := &BuildMetadata{
		rollback: false,
		docker: cli,
		cfg: cfg,
		ports: make([]BuildPortBinding, 0),
		ImageTag: version,
	}

	return build, nil
}

// create rollback
func CreateRollback(cfg config.Config) (*BuildMetadata, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	// create rollback build
	build := &BuildMetadata{
		docker: cli,
		cfg: cfg,
		ports: make([]BuildPortBinding, 0),
		rollback: true,
	}

	// get from from deployLog
	if err := build.popFromDeployLog(); err != nil {
		return nil, err
	}

	return build, nil
}

// finalize build
func (this *BuildMetadata) Finalize() error {
	// log current release if deploy succeeded
	if this.rollback == false {
		if err := this.pushToDeployLog(); err != nil {
			return err
		}
	}

	// keep deploy log organized
	if err := this.truncateDeployLog(); err != nil {
		return err
	}

	return nil
}