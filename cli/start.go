package cli

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/xorgal/xtund/internal"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start systemd services and launch xtun daemon",
	Run: func(cmd *cobra.Command, args []string) {

		serviceRunning := internal.IsXtundServiceExists(true)
		s := internal.Service.XTUND
		if serviceRunning {
			log.Fatalf("%s already running", s)
		} else {
			internal.StartService(s, false)
		}
	},
}
