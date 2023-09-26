//go:build linux
// +build linux

package internal

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/xorgal/xtun-core/pkg/config"
)

type ServiceConfig struct {
	DeviceName string
}

type Services struct {
	Daemon         string
	RoutingService string
}

//go:embed template/xtund.tmpl
var xtundT string

//go:embed template/iptables.tmpl
var iptablesT string

// CreateXtundService initializes the xtund service by creating and configuring
// its systemd service file. It then enables the service to run at startup.
func CreateXtundService() {
	t, err := template.New("xtund").Parse(xtundT)
	if err != nil {
		log.Fatalf("Cannot parse service template: %v", err)
	}
	file, err := os.Create(fmt.Sprintf("/etc/systemd/system/%s", Service.XTUND))
	if err != nil {
		log.Fatalf("Cannot create %v file: %v", Service.XTUND, err)
	}
	defer file.Close()
	err = t.Execute(file, nil)
	if err != nil {
		log.Fatalf("Cannot write %s file: %v", Service.XTUND, err)
	}

	cmd := exec.Command("systemctl", "enable", Service.XTUND)
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Cannot enable %s: %v", Service.XTUND, err)
	}
}

// CreateIptablesService initializes the iptables service by creating and configuring
// its systemd service file. It then enables the service to run at startup.
func CreateIptablesService(cfg config.Config) {
	t, err := template.New("iptables").Parse(iptablesT)
	if err != nil {
		log.Fatalf("Cannot parse service template: %v", err)
	}
	file, err := os.Create(fmt.Sprintf("/etc/systemd/system/%s", Service.IPTABLES))
	if err != nil {
		log.Fatalf("Cannot create %s file: %v", Service.IPTABLES, err)
	}
	defer file.Close()
	err = t.Execute(file, ServiceConfig{
		DeviceName: cfg.DeviceName,
	})
	if err != nil {
		log.Fatalf("Cannot write %s file: %v", Service.IPTABLES, err)
	}
	cmd := exec.Command("systemctl", "enable", Service.IPTABLES)
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Cannot enable %s: %v", Service.IPTABLES, err)
	}
}

// IsXtundServiceExists checks if the xtund service exists.
// If isRunning is true, it checks specifically if the service is currently running.
// If isRunning is false, it checks if the service exists, regardless of its current status (running or not).
// It returns true if the service meets the specified conditions, and false otherwise.
func IsXtundServiceExists(isRunning bool) bool {
	serviceExists, err := IsServiceExists(Service.XTUND, isRunning)
	if err != nil {
		return false
	}

	return serviceExists
}

// IsIptablesServiceExists checks if the xtun-iptables service exists.
// If isRunning is true, it checks specifically if the service is currently running.
// If isRunning is false, it checks if the service exists, regardless of its current status (running or not).
// It returns true if the service meets the specified conditions, and false otherwise.
func IsIptablesServiceExists(isRunning bool) bool {
	serviceExists, err := IsServiceExists(Service.IPTABLES, isRunning)
	if err != nil {
		return false
	}

	return serviceExists
}

// ReloadSystemd instructs systemd to reload its configuration. This is necessary
// after changes to systemd service files.
func ReloadSystemd() {
	cmd := exec.Command("systemctl", "daemon-reload")
	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to cast daemon-reload: %v", err)
	}
}

// StartService starts or reloads a systemd service. The serviceName parameter
// specifies the name of the service. If reload is set to true, the service will
// be reloaded instead of started.
func StartService(serviceName string, reload bool) {
	subCmd := "start"
	if reload {
		subCmd = "reload"
	}
	cmd := exec.Command("systemctl", subCmd, serviceName)
	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to %s %s: %v", subCmd, serviceName, err)
	}

	log.Printf("%s %sed", serviceName, subCmd)
}

// StopService attempts to stop a systemd service. The serviceName parameter
// specifies the name of the service to be stopped.
// The function first checks if the service exists and is currently running using the
// IsServiceExists function. If the service is running, it attempts to stop it using the
// "systemctl stop" command.
//
// If the service is successfully stopped, a log message is printed.
// If there is an error while stopping the service, the error is logged and execution is halted.
//
// Note: This function does not return any value. Errors are handled within the function
// by logging and halting execution.
func StopService(serviceName string) {
	serviceExists, err := IsServiceExists(serviceName, true)
	if err != nil {
		serviceExists = false
	}

	if serviceExists {
		cmd := exec.Command("systemctl", "stop", serviceName)
		err := cmd.Run()
		if err != nil {
			log.Fatalf("failed to stop %s: %v", serviceName, err)
		}

		log.Printf("%s stopped", serviceName)
	}
}

// StopService attempts to restart systemd service.
func RestartService(serviceName string) {
	serviceExists, err := IsServiceExists(serviceName, true)
	if err != nil {
		serviceExists = false
	}

	if serviceExists {
		cmd := exec.Command("systemctl", "restart", serviceName)
		err := cmd.Run()
		if err != nil {
			log.Fatalf("failed to restart %s: %v", serviceName, err)
		}

		log.Printf("%s restarted", serviceName)
	}
}

// IsServiceExists checks if a systemd service exists.
//
// The function takes a service name and a boolean flag isRunning as arguments.
// The serviceName is the name of the systemd service to check.
// If isRunning is set to true, the function checks if the service is currently running.
// If isRunning is set to false, the function checks if the service exists,
// regardless of whether it's currently running or not.
//
// The function returns a boolean indicating whether the service exists (and is running,
// if isRunning is true), and an error if something went wrong while trying to check the service.
// If the service does not exist, the function will return false, but no error.
func IsServiceExists(serviceName string, isRunning bool) (bool, error) {
	out, err := exec.Command("systemctl", "show", "--no-page", serviceName).Output()
	if err != nil {
		return false, err
	}

	props := make(map[string]string)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if line != "" {
			prop := strings.SplitN(line, "=", 2)
			props[prop[0]] = prop[1]
		}
	}

	if loadState, ok := props["LoadState"]; ok {
		if loadState == "not-found" {
			return false, nil // service does not exist
		}
	} else {
		return false, errors.New("could not get LoadState")
	}

	if isRunning {
		if activeState, ok := props["ActiveState"]; ok {
			if subState, ok := props["SubState"]; ok {
				return activeState == "active" && (subState == "running" || subState == "exited"), nil
			}
		}
		return false, errors.New("could not get ActiveState/SubState")
	}

	return true, nil // service exists
}
