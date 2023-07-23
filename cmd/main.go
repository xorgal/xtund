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

package main

import (
	"github.com/xorgal/xtund/cli"
	"github.com/xorgal/xtund/internal"
)

func init() {
	internal.SetupLogger()
}

func main() {
	cli.Execute()
}
