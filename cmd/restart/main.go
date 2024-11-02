package main

import (
	"flag"
	"github.com/WQGroup/logger"
	"github.com/allanpk716/wireguard_restart/internal"
	"os"
)

func main() {
	action := flag.String("action", "", "Action to perform (install, uninstall)")
	flag.Parse()

	configFile := "config.yaml"

	switch *action {
	case "install":
		err := internal.InstallService(configFile)
		if err != nil {
			logger.Errorln("Error installing service:", err)
			os.Exit(1)
		}
		logger.Infoln("Service installed successfully.")
	case "uninstall":
		err := internal.UninstallService()
		if err != nil {
			logger.Errorln("Error uninstalling service:", err)
			os.Exit(1)
		}
		logger.Infoln("Service uninstalled successfully.")
	default:
		internal.Run(configFile)
	}
}
