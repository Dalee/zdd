package release

import (
	"fmt"
	"errors"

	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
	"encoding/json"
	"bytes"
)

type downloadResponse struct {
	Status         string `json:"status"`
	Progress       string `json:"progress"`
}

// ensure image is in local cache
func (this *BuildMetadata) LoadImage() error {
	imageRef := this.getImageRef()

	fmt.Println(fmt.Sprintf("Ensure requested image \"%s\" is present in cache...", imageRef))
	if err := this.findImage(imageRef); err == nil {
		return nil
	}

	fmt.Println(fmt.Sprintf("Requested image \"%s\" not found in local cache...", imageRef))
	if err := this.fetchImage(imageRef); err != nil {
		return err
	}

	return this.findImage(imageRef)
}

// get imageRef, if it's not already defined, format imageRef as "image:version"
func (this *BuildMetadata) getImageRef() string {
	if this.ImageRef == "" {
		this.ImageRef = fmt.Sprintf(
			"%s:%s",
			this.cfg.Image,
			this.ImageTag,
		)
	}
	return this.ImageRef
}

// find release image in local cache
func (this *BuildMetadata) findImage(imageRef string) error {
	fmt.Println("Searching for image:", imageRef)
	options := types.ImageListOptions{MatchName: imageRef}
	imageList, err := this.docker.ImageList(context.Background(), options)
	if err != nil {
		return err
	}

	if len(imageList) > 0 {
		fmt.Println("Image found:", imageList[0].ID)
		this.ImageId = imageList[0].ID
		return nil
	}

	return errors.New(fmt.Sprintf("Image: \"%s\" not found", imageRef))
}

// pull image from registry/hub
func (this *BuildMetadata) fetchImage(imageRef string) error {
	// ok, not founded, trying to pull image from registry (or hub)
	fmt.Println("Trying to pull image:", imageRef)

	pullOptions := types.ImagePullOptions{}
	r, err := this.docker.ImagePull(context.Background(), imageRef, pullOptions)
	if err != nil {
		return err
	}

	defer r.Close()
	for {
		data := make([]byte, 2048)
		_, err := r.Read(data)
		if err != nil {
			break
		}

		data = bytes.Trim(data, "\x00")
		resp := &downloadResponse{}

		err = json.Unmarshal(data, resp)
		if (resp.Progress != "") {
			fmt.Printf("\r %s", resp.Progress)
		}
	}

	fmt.Println("")
	fmt.Println("Download done")
	return nil
}
