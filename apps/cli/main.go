package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator"
	"github.com/fastygo/lab/packages/orchestrator/memory"
	"github.com/fastygo/lab/packages/registry"
)

const version = "0.2.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "version":
		fmt.Printf("lab %s\n", version)
	case "labs":
		for _, id := range registry.KnownLabs() {
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
  lab run -f <manifest.yaml> [-o|--out <report.json>]

Labs: demo, quality, org, sec, static-web

`)
}

func cmdRun(args []string) error {
	var path, outPath string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-f", "--file":
			i++
			if i >= len(args) {
				return fmt.Errorf("-f requires a path")
			}
			path = args[i]
		case "-o", "--out":
			i++
			if i >= len(args) {
				return fmt.Errorf("-o requires a path")
			}
			outPath = args[i]
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
	root := registry.FindRepoRoot()
	eng := orchestrator.New(
		registry.DefaultAdapters(root),
		registry.DefaultRunners(),
		memory.New(),
	)
	report, err := eng.Run(context.Background(), m)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return err
	}
	if _, err := os.Stdout.Write(buf.Bytes()); err != nil {
		return err
	}
	if outPath != "" {
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return fmt.Errorf("create out dir: %w", err)
		}
		if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
			return fmt.Errorf("write -o file: %w", err)
		}
	}
	if report.Status == domain.StatusFail {
		os.Exit(1)
	}
	return nil
}
