package cli

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func ExecCli() {
	app := &cli.App{
		Name:  "goRoot",
		Usage: "A server to execute and manage root and C/C++ code.",
		Commands: []*cli.Command{
			startServer,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Println(err)
	}
}
