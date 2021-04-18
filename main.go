package main

import (
	"log"
	"os"

	"github.com/dacci/lambda-bridge/sqs"
	"github.com/dacci/lambda-bridge/util"
	"github.com/urfave/cli/v2"
)

var Running = true

func main() {
	app := &cli.App{
		Usage: "Invokes Lambda container",
		Commands: []*cli.Command{
			sqs.Command,
		},
	}

	go util.Notify()

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
