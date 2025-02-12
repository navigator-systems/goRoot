package cli

import (
	"fmt"
	"goRoot/config"
	"goRoot/server"

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
