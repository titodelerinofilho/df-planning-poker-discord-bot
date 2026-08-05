package main

import (
	"context"
	"fmt"
	"io"
	"os"
)

const startupMessage = "df planning poker bot started"

func main() {
	err := run(context.Background(), os.Stdout)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, stdout io.Writer) error {
	err := ctx.Err()

	if err != nil {
		return fmt.Errorf("start bot: %w", err)
	}

	_, err = fmt.Fprintln(stdout, startupMessage)

	if err != nil {
		return fmt.Errorf("write startup message: %w", err)
	}

	return nil
}
