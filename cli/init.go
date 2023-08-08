package cli

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/xorgal/xtun-core/pkg/config"
	"github.com/xorgal/xtun-core/pkg/netutil"
	"github.com/xorgal/xtund/internal"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the xtun daemon and create required systemd services",
	Run: func(cmd *cobra.Command, args []string) {
		if config.AppConfig.Protocol != "ws" && config.AppConfig.Protocol != "wss" {
			log.Fatalln("unknown protocol:", config.AppConfig.Protocol)
		}
		gateway, err := netutil.DiscoverGateway(true)
		if err != nil {
			log.Fatalf("failed to discover gateway: %v", err)
		}
		config.AppConfig.LocalGateway = gateway.String()
		config.AppConfig.ServerMode = true
		config.AppConfig.GlobalMode = false
		config.AppConfig.GUIMode = false
		config.AppConfig.InsecureSkipVerify = false

		errs := internal.MakeAllDirs()
		if len(errs) != 0 {
			for _, err := range errs {
				log.Fatal(err)
			}
		}

		internal.StopService(internal.Service.XTUND)
		internal.InitProtocol(config.AppConfig)
		err = internal.SaveConfigFile(config.AppConfig)
		if err != nil {
			log.Fatal(err)
		}

		serviceExists := internal.IsIptablesServiceExists(false)
		if !serviceExists {
			internal.CreateIptablesService(config.AppConfig)
		}

		serviceExists = internal.IsXtundServiceExists(false)
		if !serviceExists {
			internal.CreateXtundService()
		}

		log.Println("xtun successfully initialized! \U0001F680")
	},
}

func init() {
	initCmd.Flags().StringVarP(&config.AppConfig.ServerAddr, "server-address", "s", "", "Specify the server's IP address and port (Format: \"IP:port\")")
	initCmd.MarkFlagRequired("server-address")
	initCmd.Flags().StringVarP(&config.AppConfig.CIDR, "cidr", "c", "10.0.10.1/24", "Specify the CIDR block for the TUN device")
	initCmd.Flags().StringVarP(&config.AppConfig.Protocol, "protocol", "p", "wss", "Set the WebSocket protocol. Allowed values: \"ws\" or \"wss\"")
	initCmd.Flags().StringVarP(&config.AppConfig.DeviceName, "device-name", "n", "xtun", "Assign a custom name to the TUN device")
	initCmd.Flags().StringVarP(&config.AppConfig.Key, "key", "k", "xtun@2023", "Set the authentication key")
	initCmd.Flags().IntVarP(&config.AppConfig.MTU, "mtu", "m", 1500, "Specify the Maximum Transmission Unit (MTU) for the TUN device")
	initCmd.Flags().IntVarP(&config.AppConfig.BufferSize, "buffer-size", "b", 64*1024, "Set the size of the buffer for packet handling")
	initCmd.Flags().BoolVarP(&config.AppConfig.Compress, "compress", "z", false, "Enable compression")
}
