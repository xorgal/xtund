// TODO:
// 1. Refactor `rootCmd` to display help menu if no subcommand provided
// 2. Interact with daemon by means of subcommands only
// 3. Add more subcommands: `start`, `stop`, `reload`, `status`, `stats`, `reset`
// 4. Refactor `init` subcommand so it makes initial configuration, but do not start systemd services
//
// Workflow example:
// `xtund init -xzs :3001 -k xtun@2023`
// `xtund start`
// `xtund status`
// `xtund stats`

package cli

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/xorgal/xtun-core/pkg/config"
	"github.com/xorgal/xtund/internal"
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

		internal.StartServer(config.AppConfig)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
}

func Execute() {
	rootCmd.Execute()
}
