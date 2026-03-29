package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/Nam-Cheol/namba-ai/internal/namba"
)

func main() {
	configureUTF8Output()
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 1 && args[0] == "--version" {
		_, err := fmt.Fprintln(stdout, namba.VersionLine())
		return err
	}

	app := namba.NewApp(stdout, stderr)
	return app.Run(context.Background(), args)
}
