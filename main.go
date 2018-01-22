package main

import (
	"log"
	"os"

	"github.com/gogap/config"
	"github.com/gogap/go-pandoc/server"
	"github.com/urfave/cli"
)

import (
	_ "github.com/gogap/go-pandoc/pandoc/fetcher/data"
	_ "github.com/gogap/go-pandoc/pandoc/fetcher/http"
)

func main() {

	var err error

	defer func() {
		if err != nil {
			log.Printf("[go-pandoc]: %s\n", err.Error())
		}
	}()

	app := cli.NewApp()

	app.Usage = "A server for pandoc command"

	app.Commands = cli.Commands{
		cli.Command{
			Name:   "run",
			Usage:  "run pandoc service",
			Action: run,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config,c",
					Usage: "config filename",
					Value: "app.conf",
				},
				cli.StringFlag{
					Name:  "cwd",
					Usage: "change work dir",
				},
			},
		},
	}

	err = app.Run(os.Args)
}

func run(ctx *cli.Context) (err error) {

	cwd := ctx.String("cwd")
	if len(cwd) != 0 {
		err = os.Chdir(cwd)
	}

	if err != nil {
		return
	}

	configFile := ctx.String("config")

	conf := config.NewConfig(
		config.ConfigFile(configFile),
	)

	srv, err := server.New(conf)

	if err != nil {
		return
	}

	err = srv.Run()

	return
}
