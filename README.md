# holyc-compiler

A compiler that translates [HolyC](https://templeos.org) source files (`.HC`) into stack-based VM bytecode.

## Overview

holyc-compiler parses a subset of HolyC — the language of [TempleOS](https://templeos.org) — and emits a compact bytecode targeting a custom stack-based virtual machine. The instruction set is designed for arithmetic-heavy workloads such as smart contract execution, with precise gas metering per instruction.

The compilation pipeline is:

```
.HC source  →  Lexer  →  Parser (AST)  →  CodeGen  →  .hcb bytecode
```

## Instruction Set

The VM is a 64-bit stack machine. All values are `I64` (signed) or `U64` (unsigned) depending on the opcode.

| Opcode      | Code | Stack effect        | Description                          | Gas    |
|-------------|------|---------------------|--------------------------------------|--------|
| STOP        | 0x00 | —                   | Halt execution                       | 0      |
| ADD         | 0x01 | a, b → a+b          | Wrapping addition                    | 3      |
| MUL         | 0x02 | a, b → a×b          | Low 64 bits of product               | 5      |
| SUB         | 0x03 | a, b → a−b          | Wrapping subtraction                 | 3      |
| DIV         | 0x04 | a, b → a÷b          | Unsigned division (0 if b=0)         | 5      |
| SDIV        | 0x05 | a, b → a÷b          | Signed division                      | 5      |
| MOD         | 0x06 | a, b → a%b          | Unsigned modulo                      | 5      |
| SMOD        | 0x07 | a, b → a%b          | Signed modulo                        | 5      |
| ADDMOD      | 0x08 | a, b, m → (a+b)%m   | Addition modulo (65-bit intermediate)| 8      |
| MULMOD      | 0x09 | a, b, m → (a×b)%m   | Multiply modulo (128-bit intermediate)| 8     |
| EXP         | 0x0A | a, b → aᵇ           | Wrapping exponentiation              | 8+dyn  |
| SIGNEXTEND  | 0x0B | b, x → sext(x, b)   | Sign-extend x from byte b            | 5      |
| MULHI       | 0x0C | a, b → hi64(a×b)    | High 64 bits of 128-bit product      | 5      |
| MODEXP      | 0x0D | base, exp, mod → r  | Modular exponentiation               | 20     |
| ADDCARRY    | 0x0E | a, b, cin → sum, cout | Addition with carry (pushes 2)     | 5      |
| FIXMUL18    | 0x0F | a, b → (a×b)/10¹⁸  | Fixed-point multiply, 18 decimals    | 5      |
| LT          | 0x10 | a, b → a<b          | Unsigned less-than (0 or 1)          | 3      |
| GT          | 0x11 | a, b → a>b          | Unsigned greater-than (0 or 1)       | 3      |
| SLT         | 0x12 | a, b → a<b          | Signed less-than I64 (0 or 1)        | 3      |
| SGT         | 0x13 | a, b → a>b          | Signed greater-than I64 (0 or 1)     | 3      |
| EQ          | 0x14 | a, b → a==b         | Equality (0 or 1)                    | 3      |
| ISZERO      | 0x15 | a → a==0            | Zero test (0 or 1)                   | 3      |
| AND         | 0x16 | a, b → a&b          | Bitwise AND                          | 3      |
| OR          | 0x17 | a, b → a\|b         | Bitwise OR                           | 3      |
| XOR         | 0x18 | a, b → a^b          | Bitwise XOR                          | 3      |
| NOT         | 0x19 | a → ~a              | Bitwise NOT                          | 3      |
| BYTE        | 0x1A | i, x → byte[i]      | Byte i of x (0=MSB)                  | 3      |
| SHL         | 0x1B | shift, val → val<<shift | Left shift (logical)             | 3      |
| SHR         | 0x1C | shift, val → val>>shift | Right shift (logical, unsigned)  | 3      |
| SAR         | 0x1D | shift, val → val>>>shift | Right shift (arithmetic, signed) | 3      |
| CLZ         | 0x1E | a → clz(a)          | Count leading zeros (0..64)          | 3      |
| FIXDIV18    | 0x1F | a, b → (a×10¹⁸)/b  | Fixed-point divide, 18 decimals      | 5      |
| HASH        | 0x20 | offset, size → h    | Keccak-256 of mem[offset:offset+size]| 30†    |
| ROL         | 0x21 | shift, val → rotl   | Rotate left 64 bits                  | 3      |
| ROR         | 0x22 | shift, val → rotr   | Rotate right 64 bits                 | 3      |
| POPCNT      | 0x23 | a → popcount(a)     | Count bits set to 1                  | 3      |
| BSWAP       | 0x24 | a → bswap(a)        | Byte-swap (LE↔BE)                    | 3      |

† HASH gas = 30 + 6×⌈size/32⌉

Memory and control-flow opcodes (MSTORE, JUMP, RETURN…) are also available — see [`pkg/codegen/opcode.go`](pkg/codegen/opcode.go).

## HolyC Syntax Supported

```c
// Types
I64 x = 42;
U64 y = 0xFF;
F64 z = 3.14;

// Arithmetic operators → opcodes
I64 a = 10 + 20;   // ADD
I64 b = 10 - 3;    // SUB
I64 c = 6 * 7;     // MUL
I64 d = 10 / 3;    // SDIV
I64 e = 17 % 5;    // SMOD
I64 f = 2 ` 10;    // EXP  (backtick = power in HolyC)

// Builtin functions → opcodes
I64 r = AddMod(10, 20, 7);   // ADDMOD
I64 s = MulMod(10, 20, 7);   // MULMOD
I64 t = MulHi(a, b);         // MULHI
I64 u = ModExp(3, 7, 11);    // MODEXP
I64 v = AddCarry(a, b, 0);   // ADDCARRY (pushes sum + carry)
I64 w = FixMul18(a, b);      // FIXMUL18
I64 x = SignExtend(0, 255);  // SIGNEXTEND

// Unsigned variants (via builtins only)
I64 q = Div(a, b);    // DIV     (unsigned)
I64 p = Mod(a, b);    // MOD     (unsigned)

// Bit / fixed-point builtins
I64 z = Clz(0xFF00000000000000);       // CLZ     → 0
I64 y = FixDiv18(1000000000000000000,
                  500000000000000000);  // FIXDIV18 → 2×10¹⁸

// Functions with return
I64 Square(I64 x) {
  return x * x;
}
```

> Operators `/` and `%` map to signed opcodes SDIV/SMOD by default.
> Use `Div()` / `Mod()` builtins to get unsigned DIV/MOD.

## Encoding

Bytecode is encoded in **little-endian**. Each instruction is 1 byte (opcode) optionally followed by an immediate value:

| Instruction | Bytes                        |
|-------------|------------------------------|
| `PUSH0`     | `5F`                         |
| `PUSH1 v`   | `60 vv`                      |
| `PUSH8 v`   | `67 v1 v2 v3 v4 v5 v6 v7 v8` |
| any other   | `op`                         |

Output files use the `.hcb` extension (HolyC Bytecode).

## Installation

Requires [Go 1.21+](https://go.dev/dl/).

```bash
git clone <repo>
cd holyc-compiler
go build -o holyc ./cmd/holyc/
```

## Usage

```bash
# Human-readable assembly with gas costs (default)
./holyc file.HC

# Same, explicit flag
./holyc file.HC --asm

# Hex string (stdout)
./holyc file.HC --hex

# Binary file → file.hcb
./holyc file.HC --bin

# Binary file with custom output name
./holyc file.HC --bin -o output.bin
```

### Example

```bash
$ cat tests/test_simple.HC
I64 a = 3 + 4;
I64 b = MulMod(10, 20, 7);
I64 c = 2 ` 10;

$ ./holyc tests/test_simple.HC --asm
  0000  PUSH1 0x3             ; 0x60  gas=2
  0001  PUSH1 0x4             ; 0x60  gas=2
  0002  ADD                   ; 0x01  gas=3
  0003  PUSH1 0xA             ; 0x60  gas=2
  0004  PUSH1 0x14            ; 0x60  gas=2
  0005  PUSH1 0x7             ; 0x60  gas=2
  0006  MULMOD                ; 0x09  gas=8
  0007  PUSH1 0x2             ; 0x60  gas=2
  0008  PUSH1 0xA             ; 0x60  gas=2
  0009  EXP                   ; 0x0A  gas=8
  0010  STOP                  ; 0x00  gas=0
```

## Project Structure

```
holyc-compiler/
├── cmd/holyc/
│   └── main.go          # Entry point, CLI flags, output formatting
├── pkg/
│   ├── lexer/
│   │   ├── token.go     # Token types (TokenType constants)
│   │   └── lexer.go     # HolyC lexer
│   ├── parser/
│   │   ├── ast.go       # AST node types
│   │   └── parser.go    # Pratt parser
│   └── codegen/
│       ├── opcode.go    # Opcode definitions, Instruction type, gas table
│       └── codegen.go   # AST → bytecode code generator
├── tests/
│   ├── test_simple.HC   # One of each opcode
│   ├── test_arith.HC    # All builtins + operators
│   ├── test_return.HC   # Function return
│   └── test_vm.HC       # VM-oriented test
└── go.mod
```
