package main

// Chaque nœud de l'AST implémente Node.
type Node interface {
	nodeType() string
}

// ---- Expressions ----

type IntLiteral struct {
	Value int64
}
func (n *IntLiteral) nodeType() string { return "IntLiteral" }

type FloatLiteral struct {
	Value float64
}
func (n *FloatLiteral) nodeType() string { return "FloatLiteral" }

type StringLiteral struct {
	Value string
}
func (n *StringLiteral) nodeType() string { return "StringLiteral" }

type Identifier struct {
	Name string
}
func (n *Identifier) nodeType() string { return "Identifier" }

// Opération binaire: a + b, a * b, etc.
type BinaryExpr struct {
	Op    TokenType
	Left  Node
	Right Node
}
func (n *BinaryExpr) nodeType() string { return "BinaryExpr" }

// Opération unaire: -a, ~a, !a
type UnaryExpr struct {
	Op      TokenType
	Operand Node
}
func (n *UnaryExpr) nodeType() string { return "UnaryExpr" }

// Appel de fonction: Print("hello"), MulMod(a, b, m)
type CallExpr struct {
	Func string
	Args []Node
}
func (n *CallExpr) nodeType() string { return "CallExpr" }

// Accès tableau: a[i]
type IndexExpr struct {
	Array Node
	Index Node
}
func (n *IndexExpr) nodeType() string { return "IndexExpr" }

// Accès membre: a.x, a->x
type MemberExpr struct {
	Object Node
	Member string
	Arrow  bool // true pour ->, false pour .
}
func (n *MemberExpr) nodeType() string { return "MemberExpr" }

// Assignation: a = b, a += b, etc.
type AssignExpr struct {
	Op     TokenType // TOK_ASSIGN, TOK_PLUS_EQ, etc.
	Target Node
	Value  Node
}
func (n *AssignExpr) nodeType() string { return "AssignExpr" }

// Post-incrément/décrément: a++, a--
type PostfixExpr struct {
	Op      TokenType
	Operand Node
}
func (n *PostfixExpr) nodeType() string { return "PostfixExpr" }

// Cast HolyC postfix: expr(I64), expr(U8 *)
type CastExpr struct {
	Expr     Node
	TypeName string
}
func (n *CastExpr) nodeType() string { return "CastExpr" }

// sizeof(Type)
type SizeofExpr struct {
	TypeName string
}
func (n *SizeofExpr) nodeType() string { return "SizeofExpr" }

// ---- Statements ----

// Déclaration de variable: I64 x = 5;
type VarDecl struct {
	TypeName string
	Name     string
	Init     Node   // peut être nil
	IsPtr    bool
}
func (n *VarDecl) nodeType() string { return "VarDecl" }

// Statement expression: une expression suivie de ;
type ExprStmt struct {
	Expr Node
}
func (n *ExprStmt) nodeType() string { return "ExprStmt" }

// return expr;
type ReturnStmt struct {
	Value Node // peut être nil
}
func (n *ReturnStmt) nodeType() string { return "ReturnStmt" }

// if (cond) body [else elsebody]
type IfStmt struct {
	Cond Node
	Body Node
	Else Node // peut être nil
}
func (n *IfStmt) nodeType() string { return "IfStmt" }

// while (cond) body
type WhileStmt struct {
	Cond Node
	Body Node
}
func (n *WhileStmt) nodeType() string { return "WhileStmt" }

// for (init; cond; post) body
type ForStmt struct {
	Init Node
	Cond Node
	Post Node
	Body Node
}
func (n *ForStmt) nodeType() string { return "ForStmt" }

// { stmts... }
type Block struct {
	Stmts []Node
}
func (n *Block) nodeType() string { return "Block" }

// Déclaration de fonction
type FuncDecl struct {
	ReturnType string
	Name       string
	Params     []FuncParam
	Body       *Block
}
func (n *FuncDecl) nodeType() string { return "FuncDecl" }

type FuncParam struct {
	TypeName string
	Name     string
	Default  Node // HolyC supporte les valeurs par défaut
}

// Programme complet
type Program struct {
	Decls []Node
}
func (n *Program) nodeType() string { return "Program" }
