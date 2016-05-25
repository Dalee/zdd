package release

import (
	"strings"
	"io/ioutil"
	"os/exec"
	"fmt"
)

//
func (this *BuildMetadata) UpdateUpstream() error {
	fmt.Println("Updating upstream config..")

	if err := this.writeUpstreamConfig(); err != nil {
		return err
	}

	if err := this.reloadUpstream(); err != nil {
		return err
	}

	return nil
}

//
func (this *BuildMetadata) writeUpstreamConfig() error {
	var result error

	for _, upstream := range this.cfg.Upstream {
		template := upstream.Template
		for _, port := range this.ports {
			template = strings.Replace(template, port.AddressKey, port.Address, -1)
			template = strings.Replace(template, port.PortKey, port.Port, -1)
		}

		templateData := []byte(template)
		if err := ioutil.WriteFile(upstream.Resource, templateData, 0644); err != nil {
			result = err
			break
		}
	}

	return result
}

//
func (this *BuildMetadata) reloadUpstream() error {
	var result error

	for _, upstream := range this.cfg.Upstream {
		cmdList := strings.Fields(upstream.Command)
		_, err := exec.Command(cmdList[0], cmdList[1:]...).Output()
		if err != nil {
			result = err
			break
		}
	}

	return result
}

