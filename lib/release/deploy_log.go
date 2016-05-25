package release

import (
	"fmt"
	"os/user"
	"path/filepath"
	"os"
	"encoding/json"
	"io/ioutil"
	"strings"
	"errors"
	"bytes"
)

const (
	DEPLOY_LOG_DIR = ".zdd"
	DEPLOY_LOG_MAX_LENGTH = 10
)

// write successful release to deploy log
func (this *BuildMetadata) pushToDeployLog() error {
	fmt.Println("Updating deploy log")
	f, err := this.getDeployLogFile(os.O_CREATE | os.O_APPEND | os.O_WRONLY)
	if err != nil {
		return err
	}

	defer f.Close()
	releaseByteBlob, err := json.Marshal(this)
	if err != nil {
		return err
	}

	releaseByteBlob = bytes.Trim(releaseByteBlob, "\x00")
	releaseStrBlob := fmt.Sprintf("%s\n", string(releaseByteBlob))
	if _, err = f.WriteString(releaseStrBlob); err != nil {
		return err
	}

	return nil
}

// restore previous release from deploy log
func (this *BuildMetadata) popFromDeployLog() error {
	fmt.Println("Fetching data from deploy log..")
	f, err := this.getDeployLogFile(os.O_RDWR)
	if err != nil {
		return err
	}

	defer f.Close()
	releaseByteBlobList, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	releaseStrBlobList := string(releaseByteBlobList)
	releaseStrBlobList = strings.TrimRight(releaseStrBlobList, "\n")
	releaseList := strings.Split(releaseStrBlobList, "\n")
	if (len(releaseList) < 2) {
		return errors.New("Currently only one release deployed, can't rollback")
	}

	releaseStrBlob := releaseList[len(releaseList) - 2]
	releaseByteBlob := []byte(releaseStrBlob)
	releaseByteBlob = bytes.Trim(releaseByteBlob, "\x00")
	if err := json.Unmarshal(releaseByteBlob, this); err != nil {
		return err
	}

	return nil
}

// keep deploy log organized
func (this *BuildMetadata) truncateDeployLog() error {
	f, err := this.getDeployLogFile(os.O_RDWR)
	if err != nil {
		return err
	}

	defer f.Close()
	releaseByteBlobList, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	releaseStrBlobList := string(releaseByteBlobList)
	releaseStrBlobList = strings.TrimRight(releaseStrBlobList, "\n")
	entryList := strings.Split(releaseStrBlobList, "\n")

	// if rollback is happened, truncate last release
	if this.rollback == true {
		entryList = entryList[:len(entryList) - 1]
	}

	// if number of entries excesses DEPLOY_LOG_MAX_LENGTH,
	// keep only DEPLOY_LOG_MAX_LENGTH number of elements
	if len(entryList) > DEPLOY_LOG_MAX_LENGTH {
		entryList = entryList[len(entryList) - DEPLOY_LOG_MAX_LENGTH:]
	}

	if err := f.Truncate(0); err != nil {
		return err
	}

	for _, entry := range entryList {
		if _, err := f.WriteString(fmt.Sprintf("%s\n", entry)); err != nil {
			return err
		}
	}

	return nil
}

// get file handle to deploy log file
func (this *BuildMetadata) getDeployLogFile(flag int) (*os.File, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	deployLogPath := filepath.Join(usr.HomeDir, DEPLOY_LOG_DIR)
	if _, err = os.Stat(deployLogPath); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(deployLogPath, 0755); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	deployLogFilePath := filepath.Join(deployLogPath, fmt.Sprintf("%s.deploy_log", this.cfg.Name))
	if err != nil {
		return nil, err
	}

	return os.OpenFile(deployLogFilePath, flag, 0644)
}
