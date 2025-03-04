package cli

import (
	"fmt"
	"goRoot/config"
	"goRoot/exec"
	"goRoot/server"

	//"goRoot/utils"

	"github.com/urfave/cli/v2"
)

var startServer = &cli.Command{
	Name:  "start",
	Usage: "Start the goRoot server.",
	Action: func(c *cli.Context) error {
		println("Starting goRoot server...")
		serverConfig := config.ReadFile()
		fmt.Println("Server started on port:", serverConfig.Port)
		server.MainServer(serverConfig)
		return nil
	},
}

var cliMode = &cli.Command{
	Name:  "exec",
	Usage: "Execute goRooot from the command line.",
	Action: func(c *cli.Context) error {
		fmt.Println("Deploying service...")
		serverConfig := config.ReadFile()
		script := c.Args().Get(0)
		envVars := c.Args().Get(1)
		exec.CLIExec(serverConfig, script, envVars)

		return nil
	},
}
