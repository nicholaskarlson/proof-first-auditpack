package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nicholaskarlson/proof-first-auditpack/internal/auditpack"
)

var version = "dev"

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
	case "self-check", "selfcheck", "check":
		selfCheckCmd(os.Args[2:])
	case "version", "--version", "-v":
		versionCmd()
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
	fmt.Println("  auditpack run    --in  <dir> --out <dir> [--label <string>]")
	fmt.Println("  auditpack verify --pack <dir> [--in <dir>] [--strict]")
	fmt.Println("  auditpack self-check [--keep] [--strict]")
	fmt.Println("  auditpack version")
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
	opts.Version = version
	opts.InputLabel = "demo_input"

	if err := auditpack.Build(inDir, *outDir, opts); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if err := auditpack.VerifyPack(*outDir); err != nil {
		fmt.Println("VERIFY FAIL:", err)
		os.Exit(1)
	}
	if err := auditpack.VerifyInput(inDir, *outDir, true); err != nil {
		fmt.Println("VERIFY FAIL:", err)
		os.Exit(1)
	}

	fmt.Printf("Demo complete. Wrote audit pack to %s\n", *outDir)
}

func runCmd(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	inDir := fs.String("in", "", "input directory")
	outDir := fs.String("out", "./out", "output directory")
	label := fs.String("label", "", "optional: stable label recorded in manifest/meta (useful when --in is absolute)")
	_ = fs.Parse(args)

	if *inDir == "" {
		fmt.Println("Error: --in is required")
		fmt.Println()
		usage()
		os.Exit(2)
	}

	opts := auditpack.DefaultOptions()
	opts.Version = version
	// Record a stable label if provided; otherwise record the path as provided.
	if *label != "" {
		opts.InputLabel = *label
	} else {
		opts.InputLabel = *inDir
	}

	if err := auditpack.Build(*inDir, *outDir, opts); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Printf("Run complete. Wrote audit pack to %s\n", *outDir)
}

func verifyCmd(args []string) {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	packDir := fs.String("pack", "./out", "audit pack directory")
	outDir := fs.String("out", "", "deprecated alias for --pack")
	inDir := fs.String("in", "", "optional: original input directory to verify against manifest.json")
	strict := fs.Bool("strict", false, "if set: fail on extra input files not listed in manifest.json")
	_ = fs.Parse(args)

	// Back-compat: allow --out as alias for --pack.
	packExplicit := false
	outExplicit := false
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "pack":
			packExplicit = true
		case "out":
			outExplicit = true
		}
	})

	pack := *packDir
	if outExplicit && !packExplicit {
		pack = *outDir
	}
	if outExplicit && packExplicit && *outDir != "" && *outDir != *packDir {
		fmt.Println("Error: --pack and --out were both provided with different values")
		os.Exit(2)
	}

	if err := auditpack.VerifyPack(pack); err != nil {
		fmt.Println("VERIFY FAIL:", err)
		os.Exit(1)
	}
	fmt.Println("OK: pack integrity (manifest.sha256 + manifest.json invariants)")

	if *inDir != "" {
		if err := auditpack.VerifyInput(*inDir, pack, *strict); err != nil {
			fmt.Println("VERIFY FAIL:", err)
			os.Exit(1)
		}
		fmt.Println("OK: input tree matches manifest.json")
	}
}

func selfCheckCmd(args []string) {
	fs := flag.NewFlagSet("self-check", flag.ExitOnError)
	keep := fs.Bool("keep", false, "if set: keep the temp directory and print its path")
	strict := fs.Bool("strict", true, "if set: fail if input has extra files not listed in manifest.json")
	_ = fs.Parse(args)

	root, err := auditpack.SelfCheck(auditpack.SelfCheckOptions{
		Strict: *strict,
		Keep:   *keep,
	})
	if err != nil {
		fmt.Println("SELF-CHECK FAIL:", err)
		os.Exit(1)
	}

	if *keep {
		fmt.Println("OK: self-check passed")
		fmt.Println("Kept temp dir:", root)
	} else {
		fmt.Println("OK: self-check passed")
	}
}

func versionCmd() {
	fmt.Printf("proof-first-auditpack %s\n", version)
}
