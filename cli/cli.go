package cli

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/xorgal/xtun-core/pkg/config"
	"github.com/xorgal/xtund/internal"
	"github.com/xorgal/xtund/server"
)

var rootCmd = &cobra.Command{
	Use:     internal.BinaryMetadata.BinaryFile,
	Long:    internal.BinaryMetadata.Description,
	Version: internal.BinaryMetadata.Version,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd:   true,
		DisableDescriptions: true,
		DisableNoDescFlag:   true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := internal.LoadConfigFile()
		if err != nil {
			log.Fatalf("failed to load configuration: %v", err)
		}

		iptablesServiceRunning := internal.IsIptablesServiceExists(true)
		if !iptablesServiceRunning {
			f := internal.BinaryMetadata.BinaryFile
			s := internal.Service.IPTABLES
			log.Fatalf("%s did not setup system routes. Did you run \"%s start\"?", s, f)
		}

		server.StartServer(config.AppConfig)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(statusCmd)
}

func Execute() {
	rootCmd.Execute()
}
