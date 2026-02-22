package codegen

import (
	"fmt"
	"os"
	"strings"

	"holyc-compiler/pkg/lexer"
	"holyc-compiler/pkg/parser"
)

// CodeGen transforme l'AST en une séquence d'Instructions (opcodes de la VM).
type CodeGen struct {
	code     []Instruction
	Errors   []string
	builtins map[string]builtinInfo
}

type builtinInfo struct {
	op       Opcode
	argCount int
}

func NewCodeGen() *CodeGen {
	return &CodeGen{
		builtins: map[string]builtinInfo{
			"Add":        {OP_ADD, 2},
			"Mul":        {OP_MUL, 2},
			"Sub":        {OP_SUB, 2},
			"Div":        {OP_DIV, 2},
			"SDiv":       {OP_SDIV, 2},
			"Mod":        {OP_MOD, 2},
			"SMod":       {OP_SMOD, 2},
			"AddMod":     {OP_ADDMOD, 3},
			"MulMod":     {OP_MULMOD, 3},
			"Exp":        {OP_EXP, 2},
			"SignExtend": {OP_SIGNEXTEND, 2},
			"MulHi":      {OP_MULHI, 2},
			"ModExp":     {OP_MODEXP, 3},
			"AddCarry":   {OP_ADDCARRY, 3},
			"FixMul18":   {OP_FIXMUL18, 2},
			"Clz":        {OP_CLZ, 1},
			"FixDiv18":   {OP_FIXDIV18, 2},
			"Hash":       {OP_HASH, 2},
			"Rol":        {OP_ROL, 2},
			"Ror":        {OP_ROR, 2},
			"Popcnt":     {OP_POPCNT, 1},
			"Bswap":      {OP_BSWAP, 1},
		},
	}
}

func (cg *CodeGen) errorf(format string, args ...any) {
	msg := fmt.Sprintf("codegen: "+format, args...)
	cg.Errors = append(cg.Errors, msg)
	fmt.Fprintln(os.Stderr, msg)
}

func (cg *CodeGen) emit(op Opcode) {
	cg.code = append(cg.code, Instruction{Op: op})
}

func (cg *CodeGen) emitPush(val int64) {
	uval := uint64(val)
	if uval == 0 {
		cg.code = append(cg.code, Instruction{Op: OP_PUSH0})
		return
	}
	n := 0
	v := uval
	for v > 0 {
		n++
		v >>= 8
	}
	if val < 0 {
		n = 8
	}
	if n > 8 {
		n = 8
	}
	op := Opcode(byte(OP_PUSH1) + byte(n-1))
	cg.code = append(cg.code, Instruction{Op: op, Operand: val})
}

// Generate compile le programme entier et retourne le bytecode.
func (cg *CodeGen) Generate(prog *parser.Program) []Instruction {
	for _, decl := range prog.Decls {
		cg.genNode(decl)
	}
	cg.emit(OP_STOP)
	return cg.code
}

func (cg *CodeGen) genNode(node parser.Node) {
	switch n := node.(type) {
	case *parser.Program:
		for _, d := range n.Decls {
			cg.genNode(d)
		}
	case *parser.Block:
		for _, s := range n.Stmts {
			cg.genNode(s)
		}
	case *parser.ExprStmt:
		cg.genExpr(n.Expr)
	case *parser.VarDecl:
		if n.Init != nil {
			cg.genExpr(n.Init)
		}
	case *parser.ReturnStmt:
		if n.Value != nil {
			cg.genExpr(n.Value)
			cg.emitPush(0)
			cg.emit(OP_MSTORE)
			cg.emitPush(8)
			cg.emitPush(0)
			cg.emit(OP_RETURN)
		} else {
			cg.emitPush(0)
			cg.emitPush(0)
			cg.emit(OP_RETURN)
		}
	case *parser.FuncDecl:
		if n.Body != nil {
			cg.genNode(n.Body)
		}
	case *parser.IfStmt:
		cg.genExpr(n.Cond)
		cg.genNode(n.Body)
		if n.Else != nil {
			cg.genNode(n.Else)
		}
	case *parser.WhileStmt:
		cg.genExpr(n.Cond)
		cg.genNode(n.Body)
	case *parser.ForStmt:
		if n.Init != nil {
			cg.genNode(n.Init)
		}
		if n.Cond != nil {
			cg.genExpr(n.Cond)
		}
		cg.genNode(n.Body)
		if n.Post != nil {
			cg.genExpr(n.Post)
		}
	default:
		cg.errorf("unhandled node type: %T", node)
	}
}

func (cg *CodeGen) genExpr(node parser.Node) {
	switch n := node.(type) {
	case *parser.IntLiteral:
		cg.emitPush(n.Value)
	case *parser.FloatLiteral:
		cg.emitPush(int64(n.Value))
	case *parser.StringLiteral:
		cg.emitPush(0)
	case *parser.Identifier:
		cg.emitPush(0)
	case *parser.BinaryExpr:
		cg.genBinaryExpr(n)
	case *parser.UnaryExpr:
		cg.genUnaryExpr(n)
	case *parser.CallExpr:
		cg.genCallExpr(n)
	case *parser.AssignExpr:
		cg.genExpr(n.Value)
	case *parser.PostfixExpr:
		cg.genExpr(n.Operand)
	case *parser.SizeofExpr:
		cg.emitPush(typeSizeOf(n.TypeName))
	case *parser.CastExpr:
		cg.genExpr(n.Expr)
	case *parser.IndexExpr:
		cg.genExpr(n.Array)
		cg.genExpr(n.Index)
	case *parser.MemberExpr:
		cg.genExpr(n.Object)
	default:
		cg.errorf("unhandled expression: %T", node)
	}
}

func (cg *CodeGen) genBinaryExpr(n *parser.BinaryExpr) {
	cg.genExpr(n.Left)
	cg.genExpr(n.Right)

	switch n.Op {
	case lexer.TOK_PLUS:
		cg.emit(OP_ADD)
	case lexer.TOK_MINUS:
		cg.emit(OP_SUB)
	case lexer.TOK_STAR:
		cg.emit(OP_MUL)
	case lexer.TOK_SLASH:
		cg.emit(OP_SDIV)
	case lexer.TOK_PERCENT:
		cg.emit(OP_SMOD)
	case lexer.TOK_BACKTICK:
		cg.emit(OP_EXP)
	case lexer.TOK_AMP:
		cg.emit(OP_AND)
	case lexer.TOK_PIPE:
		cg.emit(OP_OR)
	case lexer.TOK_CARET:
		cg.emit(OP_XOR)
	case lexer.TOK_SHL:
		cg.emit(OP_SHL)
	case lexer.TOK_SHR:
		cg.emit(OP_SHR)
	case lexer.TOK_LT:
		cg.emit(OP_SLT)
	case lexer.TOK_GT:
		cg.emit(OP_SGT)
	case lexer.TOK_EQ:
		cg.emit(OP_EQ)
	case lexer.TOK_NEQ:
		cg.emit(OP_EQ)
		cg.emit(OP_ISZERO)
	case lexer.TOK_LTE:
		cg.emit(OP_SGT)
		cg.emit(OP_ISZERO)
	case lexer.TOK_GTE:
		cg.emit(OP_SLT)
		cg.emit(OP_ISZERO)
	case lexer.TOK_AND_AND:
		cg.emit(OP_ISZERO)
		cg.emit(OP_ISZERO)
		cg.emit(OP_SWAP1)
		cg.emit(OP_ISZERO)
		cg.emit(OP_ISZERO)
		cg.emit(OP_AND)
	case lexer.TOK_OR_OR:
		cg.emit(OP_OR)
		cg.emit(OP_ISZERO)
		cg.emit(OP_ISZERO)
	default:
		cg.errorf("unknown binary op: %d", n.Op)
	}
}

func (cg *CodeGen) genUnaryExpr(n *parser.UnaryExpr) {
	cg.genExpr(n.Operand)
	switch n.Op {
	case lexer.TOK_MINUS:
		cg.emitPush(0)
		cg.emit(OP_SWAP1)
		cg.emit(OP_SUB)
	case lexer.TOK_TILDE:
		cg.emit(OP_NOT)
	case lexer.TOK_BANG:
		cg.emit(OP_ISZERO)
	case lexer.TOK_PLUS_PLUS, lexer.TOK_MINUS_MINUS:
		cg.emitPush(1)
		if n.Op == lexer.TOK_PLUS_PLUS {
			cg.emit(OP_ADD)
		} else {
			cg.emit(OP_SUB)
		}
	}
}

func (cg *CodeGen) genCallExpr(n *parser.CallExpr) {
	if info, ok := cg.builtins[n.Func]; ok {
		if len(n.Args) != info.argCount {
			cg.errorf("%s expects %d args, got %d", n.Func, info.argCount, len(n.Args))
			return
		}
		for _, arg := range n.Args {
			cg.genExpr(arg)
		}
		cg.emit(info.op)
		return
	}
	cg.errorf("function '%s' not a builtin (no CALL opcode in current set)", n.Func)
	for _, arg := range n.Args {
		cg.genExpr(arg)
	}
}

func typeSizeOf(name string) int64 {
	switch strings.TrimSpace(strings.TrimRight(name, " *")) {
	case "U0", "I0":
		return 0
	case "U8", "I8", "Bool":
		return 1
	case "U16", "I16":
		return 2
	case "U32", "I32":
		return 4
	case "U64", "I64", "F64":
		return 8
	}
	return 8
}
