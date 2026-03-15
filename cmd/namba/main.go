package main

import (
	"context"
	"fmt"
	"os"

	"github.com/namba-ai/namba/internal/namba"
)

func main() {
	app := namba.NewApp(os.Stdout, os.Stderr)
	if err := app.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
