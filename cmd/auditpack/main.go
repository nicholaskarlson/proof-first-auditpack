package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nicholaskarlson/proof-first-auditpack/internal/auditpack"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "demo":
		demoCmd(os.Args[2:])
	case "run":
		runCmd(os.Args[2:])
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Println("Unknown command:", os.Args[1])
		fmt.Println()
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("proof-first-auditpack")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  auditpack demo --out ./out")
	fmt.Println("  auditpack run  --in  <dir> --out <dir>")
	fmt.Println()
	fmt.Println("v0: writes manifest.json + run_meta.json + manifest.sha256 (deterministic)")
}

func demoCmd(args []string) {
	fs := flag.NewFlagSet("demo", flag.ExitOnError)
	outDir := fs.String("out", "./out", "output directory")
	_ = fs.Parse(args)

	// Create a tiny deterministic demo input under outDir.
	inDir := filepath.Join(*outDir, "demo_input")
	_ = os.MkdirAll(inDir, 0o755)

	_ = os.WriteFile(filepath.Join(inDir, "hello.txt"), []byte("hello\n"), 0o644)
	_ = os.MkdirAll(filepath.Join(inDir, "nested"), 0o755)
	_ = os.WriteFile(filepath.Join(inDir, "nested", "world.txt"), []byte("world\n"), 0o644)

	opts := auditpack.DefaultOptions()
	opts.Version = "dev"
	opts.InputLabel = "demo_input"

	if err := auditpack.Build(inDir, *outDir, opts); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Printf("Demo complete. Wrote audit pack to %s\n", *outDir)
}

func runCmd(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	inDir := fs.String("in", "", "input directory")
	outDir := fs.String("out", "./out", "output directory")
	_ = fs.Parse(args)

	if *inDir == "" {
		fmt.Println("Error: --in is required")
		fmt.Println()
		usage()
		os.Exit(2)
	}

	opts := auditpack.DefaultOptions()
	opts.Version = "dev"
	// Record the input path as provided (stable for relative paths).
	opts.InputLabel = *inDir

	if err := auditpack.Build(*inDir, *outDir, opts); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Printf("Run complete. Wrote audit pack to %s\n", *outDir)
}
