package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: holyc-compiler <file.HC> [--hex | --asm | --bin] [-o output]\n")
		os.Exit(1)
	}

	filename := os.Args[1]
	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", filename, err)
		os.Exit(1)
	}

	mode := "asm"
	outFile := ""
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--hex":
			mode = "hex"
		case "--bin":
			mode = "bin"
		case "--asm":
			mode = "asm"
		case "-o":
			if i+1 < len(os.Args) {
				i++
				outFile = os.Args[i]
			} else {
				fmt.Fprintf(os.Stderr, "-o requires a filename\n")
				os.Exit(1)
			}
		}
	}
	// Si pas de -o, le fichier de sortie par défaut est <input>.hcb
	if outFile == "" && mode == "bin" {
		outFile = strings.TrimSuffix(filename, ".HC") + ".hcb"
	}

	// 1. Lexer
	lexer := NewLexer(string(src), filename)

	// 2. Parser
	parser := NewParser(lexer)
	program := parser.Parse()

	if len(parser.errors) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d parse error(s)\n", len(parser.errors))
		os.Exit(1)
	}

	// 3. Code generation
	cg := NewCodeGen()
	instructions := cg.Generate(program)

	if len(cg.errors) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d codegen warning(s)\n", len(cg.errors))
	}

	// 4. Output
	switch mode {
	case "asm":
		printAsm(instructions)
	case "hex":
		printHex(instructions)
	case "bin":
		writeBinFile(instructions, outFile)
	}
}

func printAsm(code []Instruction) {
	totalGas := 0
	for i, inst := range code {
		gas := 0
		if info, ok := opcodeInfo[inst.Op]; ok {
			gas = info.Gas
		}
		totalGas += gas
		fmt.Printf("  %04d  %-20s  ; 0x%02X  gas=%d\n", i, inst.String(), byte(inst.Op), gas)
	}
	fmt.Printf("\n; Total: %d instructions, estimated gas: %d\n", len(code), totalGas)
}

func printHex(code []Instruction) {
	for _, inst := range code {
		fmt.Printf("%02X", byte(inst.Op))
		if inst.Op.IsPush() {
			n := inst.Op.PushSize()
			buf := make([]byte, 8)
			binary.LittleEndian.PutUint64(buf, uint64(inst.Operand))
			for i := 0; i < n; i++ {
				fmt.Printf("%02X", buf[i])
			}
		}
	}
	fmt.Println()
}

func writeBinFile(code []Instruction, path string) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating %s: %v\n", path, err)
		os.Exit(1)
	}
	defer f.Close()

	for _, inst := range code {
		f.Write([]byte{byte(inst.Op)})
		if inst.Op.IsPush() {
			n := inst.Op.PushSize()
			buf := make([]byte, 8)
			binary.LittleEndian.PutUint64(buf, uint64(inst.Operand))
			f.Write(buf[:n])
		}
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", path)
}
