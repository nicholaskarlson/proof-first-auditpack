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
	case "verify":
		verifyCmd(os.Args[2:])
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
	fmt.Println("  auditpack demo   --out <dir>")
	fmt.Println("  auditpack run    --in  <dir> --out <dir>")
	fmt.Println("  auditpack verify --out <dir> [--in <dir>] [--strict]")
	fmt.Println()
	fmt.Println("v0: writes manifest.json + run_meta.json + manifest.sha256 (deterministic)")
	fmt.Println("v0.2+: verify checks pack integrity and (optionally) input tree integrity")
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

func verifyCmd(args []string) {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	outDir := fs.String("out", "./out", "audit pack directory")
	inDir := fs.String("in", "", "optional: original input directory to verify against manifest.json")
	strict := fs.Bool("strict", false, "if set: fail on extra input files not listed in manifest.json")
	_ = fs.Parse(args)

	if err := auditpack.VerifyPack(*outDir); err != nil {
		fmt.Println("VERIFY FAIL:", err)
		os.Exit(1)
	}
	fmt.Println("OK: pack integrity (manifest.sha256 + manifest.json invariants)")

	if *inDir != "" {
		if err := auditpack.VerifyInput(*inDir, *outDir, *strict); err != nil {
			fmt.Println("VERIFY FAIL:", err)
			os.Exit(1)
		}
		fmt.Println("OK: input tree matches manifest.json")
	}
}
