package release

import (
	"fmt"
	"errors"
	"time"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
	"strings"
)


// simple check to make sure container is running after start
func (this *BuildMetadata) CheckAlive() error {
	fmt.Println("Ensure container is still running, will wait for some time..")
	time.Sleep(3 * time.Second) // 3 sec is enough?

	options := types.ContainerListOptions{All: true}
	containerList, err := this.docker.ContainerList(context.Background(), options)
	if err != nil {
		return err
	}

	err = errors.New("Container is not running")
	for _, container := range containerList {
		fmt.Println("Checking:", container.ID)
		if container.ID == this.ContainerId {
			if strings.HasPrefix(container.Status, "Up ") {
				fmt.Println("Container is still alive, moving further...")
				err = nil
			}
			break
		}
	}

	return err
}
