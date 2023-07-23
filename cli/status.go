package cli

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/xorgal/xtund/internal"
)

var printFullReport bool

var statusOK = "OK"
var statusNotFound = "File not found"
var statusStopped = "Stopped"

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print status for xtun services",
	Run: func(cmd *cobra.Command, args []string) {
		xtundServiceExists := internal.IsXtundServiceExists(false)
		xtundServiceRunning := internal.IsXtundServiceExists(true)
		if printFullReport {
			f := internal.BinaryMetadata.BinaryFile
			v := internal.BinaryMetadata.Version

			log.Printf("%s v%s", f, v)
			log.Println("\nFilesystem:")
			configFileExists := internal.IsConfigFileExists()
			if !configFileExists {
				log.Printf("  %s: %s", internal.FilePath.ConfigPath, fmtStatus(statusNotFound))
			} else {
				log.Printf("  %s: %s", internal.FilePath.ConfigPath, fmtStatus(statusOK))
			}

			allocatorFileExists := internal.IsAllocatorFileExists()
			if !allocatorFileExists {
				log.Printf("  %s: %s", internal.FilePath.AllocatorPath, fmtStatus(statusNotFound))
			} else {
				log.Printf("  %s: %s\n", internal.FilePath.AllocatorPath, fmtStatus(statusOK))
			}

			log.Println("\nsystemd services:")
			iptablesServiceExists := internal.IsIptablesServiceExists(false)
			iptablesServiceRunning := internal.IsIptablesServiceExists(true)
			printServiceStatusFmt(internal.Service.IPTABLES, iptablesServiceExists, iptablesServiceRunning, " ")
			printServiceStatusFmt(internal.Service.XTUND, xtundServiceExists, xtundServiceRunning, "         ")
			log.Println()
		} else {
			printServiceStatus(internal.Service.XTUND, xtundServiceExists, xtundServiceRunning)
		}
	},
}

func init() {
	statusCmd.Flags().BoolVarP(&printFullReport, "full", "f", false, "Print full status report")
}

func printServiceStatus(service string, fileOK bool, isRunning bool) {
	if !fileOK {
		log.Printf("%s: not found", service)
	} else if isRunning {
		log.Printf("%s: running", service)
	} else {
		log.Printf("%s: stopped", service)
	}
}

func printServiceStatusFmt(service string, fileOK bool, isRunning bool, tab string) {
	if !fileOK {
		log.Printf("  %s:%s%s", service, tab, fmtStatus(statusNotFound))
	} else if isRunning {
		log.Printf("  %s:%s%s", service, tab, fmtStatus(statusOK))
	} else {
		log.Printf("  %s:%s%s", service, tab, fmtStatus(statusStopped))
	}
}

func fmtStatus(status string) string {
	return fmt.Sprintf("[ %s ]", status)
}
