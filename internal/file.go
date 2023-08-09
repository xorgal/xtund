package internal

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"reflect"

	"github.com/xorgal/xtun-core/pkg/config"
)

type IBinaryMetadata struct {
	BinaryFile  string
	Description string
	Version     string
}

type IDirPath struct {
	BinaryDir string
	ConfigDir string
	Data      string
	Log       string
}

type IFilePath struct {
	BinaryPath    string
	ConfigPath    string
	AllocatorPath string
}

var WorkDirCommonName = "xtun"
var ConfigFile = "config.json"
var AllocatorDBFile = "allocdb"

var BinaryMetadata = IBinaryMetadata{
	BinaryFile:  "xtund",
	Description: "Command Line Interface (CLI) tool for managing the xtun daemon",
	Version:     AppVersion,
}

var DirPath = IDirPath{
	BinaryDir: "/usr/local/sbin",
	ConfigDir: fmt.Sprintf("/etc/%s", WorkDirCommonName),
	Data:      fmt.Sprintf("/var/lib/%s", WorkDirCommonName),
	Log:       fmt.Sprintf("/var/log/%s", WorkDirCommonName),
}

var FilePath = IFilePath{
	BinaryPath:    fmt.Sprintf("%s/%s", DirPath.BinaryDir, BinaryMetadata.BinaryFile),
	ConfigPath:    fmt.Sprintf("%s/%s", DirPath.ConfigDir, ConfigFile),
	AllocatorPath: fmt.Sprintf("%s/%s", DirPath.Data, AllocatorDBFile),
}

// MakeAppDirs loops through the `DirPath` struct and makes directories
// for each field if they don't exist
func MakeAllDirs() []error {
	var errs []error
	v := reflect.ValueOf(DirPath)
	for i := 0; i < v.NumField(); i++ {
		path := v.Field(i).Interface().(string)
		_, err := mkdir(path, 0755)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// SaveConfigFile parses `config.Config` and attempts to write data
// to `FilePath.ConfigPath` at `DirPath.ConfigDir`
func SaveConfigFile(config config.Config) error {
	file, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(FilePath.ConfigPath, file, 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadConfigFile() error {
	file, err := os.ReadFile(FilePath.ConfigPath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, &config.AppConfig)
	if err != nil {
		return err
	}
	return nil
}

func IsConfigFileExists() bool {
	if _, err := os.Stat(FilePath.ConfigPath); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			log.Printf("error reading %s: %v", FilePath.ConfigPath, err)
			return false
		}
	} else {
		return true
	}
}

func IsAllocatorFileExists() bool {
	if _, err := os.Stat(FilePath.AllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			log.Printf("error reading %s: %v", FilePath.ConfigPath, err)
			return false
		}
	} else {
		return true
	}
}

func mkdir(path string, perm fs.FileMode) (fs.FileInfo, error) {
	if i, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, perm)
		if err != nil {
			return nil, err
		} else {
			return i, nil
		}
	}
	return nil, nil
}
