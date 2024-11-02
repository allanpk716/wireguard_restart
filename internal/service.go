package internal

import (
	"fmt"
	"github.com/WQGroup/logger"
	"github.com/kardianos/service"
	"os"
)

func InstallService(configFile string) error {

	svcConfig := &service.Config{
		Name:        "RestartWireGuard",
		DisplayName: "Restart WireGuard",
		Description: "Monitors a domain's IP and reconnects WireGuard tunnel if the IP changes.",
		Arguments:   []string{fmt.Sprintf("--config=%s", configFile)},
	}

	svc, err := service.New(nil, svcConfig)
	if err != nil {
		return err
	}

	err = svc.Install()
	if err != nil {
		return err
	}

	return nil
}

func UninstallService() error {
	svcConfig := &service.Config{
		Name: "IPMonitor",
	}

	svc, err := service.New(nil, svcConfig)
	if err != nil {
		return err
	}

	err = svc.Uninstall()
	if err != nil {
		return err
	}

	return nil
}

func Run(configFile string) {
	config, err := loadConfig(configFile)
	if err != nil {
		logger.Errorln("Error loading configuration:", err)
		os.Exit(1)
	}

	ipMonitor := NewIPMonitor(config)
	go ipMonitor.Start()

	// Wait for the IP monitor to stop
	<-ipMonitor.Done()
}
