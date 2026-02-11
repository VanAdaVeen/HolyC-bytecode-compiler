package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: holyc-compiler <file.HC> [--hex | --asm | --bin]\n")
		os.Exit(1)
	}

	filename := os.Args[1]
	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", filename, err)
		os.Exit(1)
	}

	mode := "asm"
	if len(os.Args) > 2 {
		switch os.Args[2] {
		case "--hex":
			mode = "hex"
		case "--bin":
			mode = "bin"
		case "--asm":
			mode = "asm"
		}
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
		writeBin(instructions)
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

func writeBin(code []Instruction) {
	for _, inst := range code {
		os.Stdout.Write([]byte{byte(inst.Op)})
		if inst.Op.IsPush() {
			n := inst.Op.PushSize()
			buf := make([]byte, 8)
			binary.LittleEndian.PutUint64(buf, uint64(inst.Operand))
			os.Stdout.Write(buf[:n])
		}
	}
}
