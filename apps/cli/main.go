package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fastygo/lab/packages/adapters/noop"
	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator"
	"github.com/fastygo/lab/packages/orchestrator/memory"
	"github.com/fastygo/lab/packages/orchestrator/stub"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "version":
		fmt.Printf("lab %s\n", version)
	case "labs":
		for _, id := range orchestrator.KnownLabs() {
			fmt.Println(id)
		}
	case "run":
		if err := cmdRun(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `lab — FastyGo laboratory constructor CLI

Usage:
  lab version
  lab labs
  lab run -f <manifest.yaml>

`)
}

func cmdRun(args []string) error {
	var path string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-f", "--file":
			i++
			if i >= len(args) {
				return fmt.Errorf("-f requires a path")
			}
			path = args[i]
		default:
			return fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if path == "" {
		return fmt.Errorf("manifest path required (-f)")
	}
	m, err := domain.LoadManifest(path)
	if err != nil {
		return err
	}
	eng := orchestrator.New(
		[]orchestrator.TargetAdapter{noop.New()},
		[]orchestrator.Runner{stub.New()},
		memory.New(),
	)
	report, err := eng.Run(context.Background(), m)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return err
	}
	switch report.Status {
	case domain.StatusFail:
		os.Exit(1)
	case domain.StatusWarn:
		os.Exit(0)
	}
	return nil
}
