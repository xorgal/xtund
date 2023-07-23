package cli

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/xorgal/xtund/internal"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop xtun daemon",
	Run: func(cmd *cobra.Command, args []string) {
		serviceRunning := internal.IsXtundServiceExists(true)
		s := internal.Service.XTUND
		if !serviceRunning {
			log.Fatalf("%s not running", s)
		} else {
			internal.StopService(s)
		}
	},
}
