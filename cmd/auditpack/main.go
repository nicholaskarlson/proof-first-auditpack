package main

import (
	"flag"
	"fmt"
	"os"
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
	fmt.Println("v0: writes manifest.json + manifest.sha256 + run_meta.json")
}

func demoCmd(args []string) {
	fs := flag.NewFlagSet("demo", flag.ExitOnError)
	outDir := fs.String("out", "./out", "output directory")
	_ = fs.Parse(args)

	_ = os.MkdirAll(*outDir, 0o755)
	fmt.Printf("Demo placeholder: would write audit pack to %s\n", *outDir)
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

	_ = os.MkdirAll(*outDir, 0o755)
	fmt.Printf("Run placeholder: would hash %s and write audit pack to %s\n", *inDir, *outDir)
}
