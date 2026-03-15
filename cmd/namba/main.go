package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Nam-Cheol/namba-ai/internal/namba"
)

func main() {
	configureUTF8Output()
	app := namba.NewApp(os.Stdout, os.Stderr)
	if err := app.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
