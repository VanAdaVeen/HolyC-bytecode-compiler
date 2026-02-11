package main

import (
	"fmt"
	"os"
	"strings"
)

// CodeGen transforme l'AST en une séquence d'Instructions (opcodes de la VM).
// Machine à pile : chaque expression pousse son résultat sur la pile.
type CodeGen struct {
	code   []Instruction
	errors []string

	// Table des builtins mappés directement sur un opcode
	builtins map[string]builtinInfo
}

type builtinInfo struct {
	op       Opcode
	argCount int
}

func NewCodeGen() *CodeGen {
	cg := &CodeGen{
		builtins: map[string]builtinInfo{
			// Fonctions HolyC mappées sur les 16 opcodes
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
		},
	}
	return cg
}

func (cg *CodeGen) errorf(format string, args ...any) {
	msg := fmt.Sprintf("codegen: "+format, args...)
	cg.errors = append(cg.errors, msg)
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
	// Détermine le nombre minimal d'octets nécessaires
	n := 0
	v := uval
	for v > 0 {
		n++
		v >>= 8
	}
	// Valeurs négatives nécessitent 8 octets complets
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
func (cg *CodeGen) Generate(prog *Program) []Instruction {
	for _, decl := range prog.Decls {
		cg.genNode(decl)
	}
	cg.emit(OP_STOP)
	return cg.code
}

func (cg *CodeGen) genNode(node Node) {
	switch n := node.(type) {
	case *Program:
		for _, d := range n.Decls {
			cg.genNode(d)
		}
	case *Block:
		for _, s := range n.Stmts {
			cg.genNode(s)
		}
	case *ExprStmt:
		cg.genExpr(n.Expr)
	case *VarDecl:
		// Pour l'instant, on génère juste l'expression d'initialisation
		if n.Init != nil {
			cg.genExpr(n.Init)
		}
	case *ReturnStmt:
		if n.Value != nil {
			cg.genExpr(n.Value)
		}
		cg.emit(OP_STOP)
	case *FuncDecl:
		if n.Body != nil {
			cg.genNode(n.Body)
		}
	case *IfStmt:
		// Pas d'opcode de branchement dans les 16 opcodes demandés,
		// on génère juste les expressions pour l'instant
		cg.genExpr(n.Cond)
		cg.genNode(n.Body)
		if n.Else != nil {
			cg.genNode(n.Else)
		}
	case *WhileStmt:
		cg.genExpr(n.Cond)
		cg.genNode(n.Body)
	case *ForStmt:
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

func (cg *CodeGen) genExpr(node Node) {
	switch n := node.(type) {
	case *IntLiteral:
		cg.emitPush(n.Value)

	case *FloatLiteral:
		// On encode le F64 en tant que I64 (bit pattern)
		// Pour l'instant on tronque
		cg.emitPush(int64(n.Value))

	case *StringLiteral:
		// Pas de support string dans les 16 opcodes
		cg.emitPush(0)

	case *Identifier:
		// Sans opcodes LOAD/STORE, on ne peut pas encore charger de variables
		// On émet un PUSH 0 comme placeholder
		cg.emitPush(0)

	case *BinaryExpr:
		cg.genBinaryExpr(n)

	case *UnaryExpr:
		cg.genUnaryExpr(n)

	case *CallExpr:
		cg.genCallExpr(n)

	case *AssignExpr:
		// Génère la valeur (sans stocker, pas d'opcode STORE)
		cg.genExpr(n.Value)

	case *PostfixExpr:
		cg.genExpr(n.Operand)

	case *SizeofExpr:
		size := typeSizeOf(n.TypeName)
		cg.emitPush(size)

	case *CastExpr:
		cg.genExpr(n.Expr)

	case *IndexExpr:
		cg.genExpr(n.Array)
		cg.genExpr(n.Index)

	case *MemberExpr:
		cg.genExpr(n.Object)

	default:
		cg.errorf("unhandled expression: %T", node)
	}
}

func (cg *CodeGen) genBinaryExpr(n *BinaryExpr) {
	// Pour les opérateurs signés vs non signés, on utilise le type HolyC
	// Par défaut I64 → signé

	cg.genExpr(n.Left)
	cg.genExpr(n.Right)

	switch n.Op {
	case TOK_PLUS:
		cg.emit(OP_ADD)
	case TOK_MINUS:
		cg.emit(OP_SUB)
	case TOK_STAR:
		cg.emit(OP_MUL)
	case TOK_SLASH:
		// Utilise SDIV par défaut (HolyC I64 est signé)
		cg.emit(OP_SDIV)
	case TOK_PERCENT:
		// Utilise SMOD par défaut
		cg.emit(OP_SMOD)
	case TOK_BACKTICK:
		// HolyC ` = power
		cg.emit(OP_EXP)
	case TOK_AMP:
		cg.emit(OP_AND)
	case TOK_PIPE:
		cg.emit(OP_OR)
	case TOK_CARET:
		cg.emit(OP_XOR)
	case TOK_SHL:
		cg.emit(OP_SHL)
	case TOK_SHR:
		cg.emit(OP_SHR) // non signé par défaut ; SAR pour signé si besoin
	case TOK_LT:
		cg.emit(OP_SLT) // signé par défaut (HolyC I64)
	case TOK_GT:
		cg.emit(OP_SGT)
	case TOK_EQ:
		cg.emit(OP_EQ)
	case TOK_NEQ:
		cg.emit(OP_EQ)
		cg.emit(OP_ISZERO) // NEQ = NOT(EQ)
	case TOK_LTE:
		cg.emit(OP_SGT)
		cg.emit(OP_ISZERO) // LTE = NOT(GT)
	case TOK_GTE:
		cg.emit(OP_SLT)
		cg.emit(OP_ISZERO) // GTE = NOT(LT)
	case TOK_AND_AND:
		// a && b → ISZERO(ISZERO(a)) * ISZERO(ISZERO(b)) → simplifié : les deux doivent être != 0
		cg.emit(OP_ISZERO)
		cg.emit(OP_ISZERO) // normalise b à 0/1
		cg.emit(OP_SWAP1)  // on a maintenant [b_norm, a]
		cg.emit(OP_ISZERO)
		cg.emit(OP_ISZERO) // normalise a à 0/1
		cg.emit(OP_AND)    // 1 & 1 = 1
	case TOK_OR_OR:
		cg.emit(OP_OR)
		cg.emit(OP_ISZERO)
		cg.emit(OP_ISZERO) // normalise à 0/1
	default:
		cg.errorf("unknown binary op: %d", n.Op)
	}
}

func (cg *CodeGen) genUnaryExpr(n *UnaryExpr) {
	cg.genExpr(n.Operand)
	switch n.Op {
	case TOK_MINUS:
		// -x = 0 - x : pile a [x], on pousse 0, swap, sub
		cg.emitPush(0)
		cg.emit(OP_SWAP1)
		cg.emit(OP_SUB)
	case TOK_TILDE:
		cg.emit(OP_NOT) // bitwise NOT
	case TOK_BANG:
		cg.emit(OP_ISZERO) // logical NOT
	case TOK_PLUS_PLUS, TOK_MINUS_MINUS:
		cg.emitPush(1)
		if n.Op == TOK_PLUS_PLUS {
			cg.emit(OP_ADD)
		} else {
			cg.emit(OP_SUB)
		}
	}
}

func (cg *CodeGen) genCallExpr(n *CallExpr) {
	// Vérifie si c'est un builtin mappé sur un opcode
	if info, ok := cg.builtins[n.Func]; ok {
		if len(n.Args) != info.argCount {
			cg.errorf("%s expects %d args, got %d", n.Func, info.argCount, len(n.Args))
			return
		}
		// Pousse les arguments sur la pile dans l'ordre
		for _, arg := range n.Args {
			cg.genExpr(arg)
		}
		cg.emit(info.op)
		return
	}
	// Fonction non reconnue → pas d'opcode CALL dans les 16
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
	return 8 // default pointer size
}

func tokenName(t TokenType) string {
	switch t {
	case TOK_PLUS:
		return "+"
	case TOK_MINUS:
		return "-"
	case TOK_STAR:
		return "*"
	case TOK_SLASH:
		return "/"
	case TOK_PERCENT:
		return "%"
	case TOK_AMP:
		return "&"
	case TOK_PIPE:
		return "|"
	case TOK_CARET:
		return "^"
	case TOK_TILDE:
		return "~"
	case TOK_BANG:
		return "!"
	case TOK_SHL:
		return "<<"
	case TOK_SHR:
		return ">>"
	case TOK_EQ:
		return "=="
	case TOK_NEQ:
		return "!="
	case TOK_LT:
		return "<"
	case TOK_GT:
		return ">"
	case TOK_LTE:
		return "<="
	case TOK_GTE:
		return ">="
	case TOK_AND_AND:
		return "&&"
	case TOK_OR_OR:
		return "||"
	case TOK_BACKTICK:
		return "`"
	case TOK_PLUS_PLUS:
		return "++"
	case TOK_MINUS_MINUS:
		return "--"
	}
	return fmt.Sprintf("tok(%d)", t)
}
