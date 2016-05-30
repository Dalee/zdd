package release

import (
	"fmt"
	"time"
	"strings"

	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"github.com/docker/engine-api/types"
	"github.com/docker/go-connections/nat"
	"golang.org/x/net/context"
	"errors"
)

// empty struct used in port and volumes configuration
type emptyStructDef struct {
}

func isContainerRunning(container types.Container) bool {
	return strings.HasPrefix(container.Status, "Up ")
}

// create container definitions
func (this *BuildMetadata) CreateContainer() error {
	if this.rollback == true {
		fmt.Println("Checking if previous container exists..")
		if found, err := this.loadRollbackContainer(); found == true || err != nil {
			fmt.Println("Previous container found, it will be used for rollback")
			return err
		}

		fmt.Println("Previous container not found, container will be created from scratch")
	}

	fmt.Println("Creating container")
	portExposed, portBindings := this.getConfigPortBinding()
	mountExposed, mountBindings := this.getConfigMountBinding()
	containerName := this.getContainerName()

	// define container config
	containerConfig := container.Config{
		Image: this.ImageId,
		Env: this.cfg.Env,
		ExposedPorts: portExposed,
		Volumes: mountExposed,
	}

	// define host config
	hostConfig := container.HostConfig{
		PortBindings:portBindings,
		Binds: mountBindings,
		AutoRemove: false,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
			MaximumRetryCount: 999999,
		},
	}

	// define network config
	networkConfig := network.NetworkingConfig{}

	// create container
	result, err := this.docker.ContainerCreate(
		context.Background(),
		&containerConfig,
		&hostConfig,
		&networkConfig,
		containerName,
	)

	if err != nil {
		return err
	}

	fmt.Println("Container created:", result.ID)
	this.ContainerId = result.ID
	return nil
}

// start container and request information
func (this *BuildMetadata) StartContainer() error {
	fmt.Println("Starting container:", this.ContainerId)
	if err := this.docker.ContainerStart(context.Background(), this.ContainerId, this.ContainerId); err != nil {
		return err
	}

	// inspect container after start
	fmt.Println("Container started, collecting ports acquired by container")
	if err := this.loadContainerPortBinding(); err != nil {
		return err
	}

	return nil
}

func (this *BuildMetadata) ExecCommands() error {
	fmt.Println("Executing commands..")

	for _, cmd := range this.cfg.Bootstrap {
		fmt.Println("Starting cmd:", cmd)

		createConfig := types.ExecConfig{Cmd: []string{"sh", "-c", cmd}}
		resp, err := this.docker.ContainerExecCreate(context.Background(), this.ContainerId, createConfig)
		if err != nil {
			return err
		}

		execConfig := types.ExecStartCheck{}
		if err := this.docker.ContainerExecStart(context.Background(), resp.ID, execConfig); err != nil {
			return err
		}

		// TODO: add timeout for command
		for {
			// make delay delay and check status of command
			time.Sleep(1 * time.Second)

			// fetch command result
			execInfo, err := this.docker.ContainerExecInspect(context.Background(), resp.ID)
			if err != nil {
				return err
			}

			// if command still running, continue
			if execInfo.Running == true {
				continue
			}

			// if not, check status
			if execInfo.ExitCode != 0 {
				return errors.New("Bootstrap commands finished with non-zero exit status")
			}

			fmt.Println("Command successfully finished:", cmd)
			break
		}
	}

	return nil
}

// stop all previous release running containers
func (this *BuildMetadata) StopOld() error {
	var result error

	options := types.ContainerListOptions{All: true}
	containerList, err := this.docker.ContainerList(context.Background(), options)
	if err != nil {
		return err
	}

	// stop container with same deployed name
	stopIdList := make([]string, 0)
	for _, container := range containerList {
		// skip stopped and our new alive container
		if !isContainerRunning(container) || container.ID == this.ContainerId {
			continue
		}

		for _, name := range container.Names {
			nameCleared := strings.TrimLeft(name, "/")
			if strings.HasPrefix(nameCleared, this.cfg.Name) {
				stopIdList = append(stopIdList, container.ID)
				break
			}
		}
	}

	for _, containerId := range stopIdList {
		fmt.Println("Stopping:", containerId)
		if err := this.docker.ContainerStop(context.Background(), containerId, 10); err != nil {
			result = err
			break
		}
	}

	return result
}

// get container name, if container name is not defined, format as "name.version.timestamp"
func (this *BuildMetadata) getContainerName() string {
	if this.ContainerName == "" {
		this.ContainerName = fmt.Sprintf(
			"%s.%s.%d",
			this.cfg.Name,
			this.ImageTag,
			time.Now().UnixNano() / 1000000,
		)
	}
	return this.ContainerName
}

// build docker port configuration definitions
func (this *BuildMetadata) getConfigPortBinding() (map[nat.Port]struct{}, nat.PortMap) {

	portExposed := make(map[nat.Port]struct{}, 0)
	portBinding := make(nat.PortMap)

	for _, portDefinition := range this.cfg.Port {
		proto, portNum := nat.SplitProtoPort(portDefinition)
		port, err := nat.NewPort(proto, portNum)
		if err == nil {
			portExposed[port] = emptyStructDef{}
			portBinding[port] = make([]nat.PortBinding, 0)
			portBinding[port] = append(portBinding[port], nat.PortBinding{
				HostIP: "127.0.0.1",
				HostPort: "0",
			})
		}
	}

	return portExposed, portBinding
}

// split path by ":", return left and right parts
func splitPathDefinition(pathDefinition string) (string, string) {
	indexLeft := strings.Index(pathDefinition, ":")
	if (indexLeft < 0) {
		// shorthand record just path
		fmt.Println(pathDefinition)
	}

	left := pathDefinition[:indexLeft]
	right := pathDefinition[indexLeft + 1:]

	return left, right
}

// build mount exposes and bindings for container
func (this *BuildMetadata) getConfigMountBinding() (map[string]struct{}, []string) {
	mountExposed := make(map[string]struct{}, 0)
	mountBinding := make([]string, 0)

	for _, mountDefinition := range this.cfg.Mount {
		_, containerPath := splitPathDefinition(mountDefinition)
		mountExposed[containerPath] = emptyStructDef{}
		mountBinding = append(mountBinding, mountDefinition)
	}

	return mountExposed, mountBinding
}

// fetch port bindings from running container
func (this *BuildMetadata) loadContainerPortBinding() (error) {
	json, err := this.docker.ContainerInspect(context.Background(), this.ContainerId)
	if err != nil {
		return err
	}

	for containerPort, binding := range json.NetworkSettings.Ports {
		if len(binding) > 0 {
			hostPort := binding[0].HostPort
			hostIp := binding[0].HostIP

			// format new port key
			portKey := fmt.Sprintf(
				"%%%s_%s%%",
				strings.ToUpper(containerPort.Proto()),
				containerPort.Port(),
			)

			// format ip key
			ipKey := fmt.Sprintf(
				"%%%s_%s_IP%%",
				strings.ToUpper(containerPort.Proto()),
				containerPort.Port(),
			)

			this.ports = append(this.ports, BuildPortBinding{
				PortKey: portKey,
				Port: hostPort,
				AddressKey: ipKey,
				Address: hostIp,
			})
		}
	}

	return nil
}

// in case of rollback, first, check existence of previous container
// if container found
func (this *BuildMetadata) loadRollbackContainer() (bool, error) {

	options := types.ContainerListOptions{All:true}
	containerList, err := this.docker.ContainerList(context.Background(), options)
	if err != nil {
		return false, err
	}

	result := false
	for _, container := range containerList {
		if container.ID == this.ContainerId {
			result = true
			break
		}
	}

	return result, nil
}
