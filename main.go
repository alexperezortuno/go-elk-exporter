package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pteich/configstruct"

	"gitlab.com/sixbell/proyectos/brasil/claro/vas/elastic-query-export/export"
	"gitlab.com/sixbell/proyectos/brasil/claro/vas/elastic-query-export/flags"
)

var _ string

func main() {
	conf := flags.Flags{
		ElasticURL:       "http://localhost:9200",
		ElasticVerifySSL: true,
		Index:            "logs-*",
		Query:            "*",
		OutFormat:        flags.FormatCSV,
		Filename:         "",
		Outpath:          "/opt/sixlabs/reports",
		ScrollSize:       1000,
		Timefield:        "@timestamp",
		Convertdate:      false,
		Dateformat:       "2006-01-02 15:04:05",
		Timezone:         "America/Santiago",
		Debug:            false,
		LogLevel:         "info",
		LogOutput:        "/opt/sixlabs/logs",
		LogFormat:        "json",
		LogResult:        false,
		LogFile:          false,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	cmd := configstruct.NewCommand(
		"",
		"CLI tool to export data from ElasticSearch into a CSV or JSON file. https://gitlab.com/sixbell/proyectos/brasil/claro/vas/elastic-query-export",
		&conf,
		func(c *configstruct.Command, cfg interface{}) error {
			export.Run(ctx, cfg.(*flags.Flags))
			return nil
		},
	)

	err := cmd.ParseAndRun(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
