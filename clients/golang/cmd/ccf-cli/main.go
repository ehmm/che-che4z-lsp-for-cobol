package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/code4z/ccf-cli/pkg/model"
	"github.com/code4z/ccf-cli/pkg/vm"
)

func main() {
	inputPath := flag.String("i", "", "Input CFAST JSON file")
	outputPath := flag.String("o", "", "Output JSONL file")
	maxVMs := flag.Int("max-vms", 1000, "Maximum VM count")
	flag.Parse()

	if *inputPath == "" || *outputPath == "" {
		fmt.Println("Usage: ccf-cli-go -i <input_cfast.json> -o <output.jsonl> [--max-vms <number>]")
		os.Exit(1)
	}

	fmt.Printf("Phase B (Go): Loading CFAST %s...\n", *inputPath)
	cfastNodes, err := model.LoadCFAST(*inputPath)
	if err != nil {
		fmt.Printf("Error loading CFAST: %v\n", err)
		os.Exit(1)
	}

	if len(cfastNodes) == 0 {
		fmt.Println("Error: No program nodes found in CFAST")
		os.Exit(1)
	}

	outputFile, err := os.Create(*outputPath)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	writer := model.NewGraphWriter(outputFile)

	fmt.Printf("Phase C (Go): Analyzing and streaming CFG to %s...\n", *outputPath)
	
	// For each program in CFAST, run analysis
	for _, programNode := range cfastNodes {
		listing := vm.NewProgramListing(programNode)
		processor := vm.NewVirtualProcessor(listing, *maxVMs, writer)
		
		fmt.Printf("Starting analysis for program: %s\n", programNode.Name)
		processor.Run()
	}

	fmt.Println("Successfully generated Control Flow Graph (Go Implementation).")
}
