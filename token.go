package main

type TokenType int

const (
	// Valeurs
	TOK_EOF     TokenType = iota
	TOK_INT               // 42, 0xFF, 0b1010
	TOK_FLOAT             // 3.14, 1e-5
	TOK_STRING            // "hello"
	TOK_CHAR              // 'A'
	TOK_IDENT             // my_var, Print

	// Types HolyC
	TOK_U0
	TOK_U8
	TOK_U16
	TOK_U32
	TOK_U64
	TOK_I8
	TOK_I16
	TOK_I32
	TOK_I64
	TOK_F64
	TOK_BOOL

	// Mots-clés
	TOK_IF
	TOK_ELSE
	TOK_WHILE
	TOK_DO
	TOK_FOR
	TOK_SWITCH
	TOK_CASE
	TOK_DEFAULT
	TOK_BREAK
	TOK_RETURN
	TOK_CLASS
	TOK_UNION
	TOK_PUBLIC
	TOK_EXTERN
	TOK_STATIC
	TOK_SIZEOF
	TOK_TRY
	TOK_CATCH
	TOK_GOTO
	TOK_INCLUDE
	TOK_DEFINE

	// Opérateurs
	TOK_PLUS     // +
	TOK_MINUS    // -
	TOK_STAR     // *
	TOK_SLASH    // /
	TOK_PERCENT  // %
	TOK_AMP      // &
	TOK_PIPE     // |
	TOK_CARET    // ^
	TOK_TILDE    // ~
	TOK_BANG     // !
	TOK_LT       // <
	TOK_GT       // >
	TOK_ASSIGN   // =
	TOK_DOT      // .
	TOK_ARROW    // ->
	TOK_HASH     // #

	// Opérateurs composés
	TOK_PLUS_PLUS   // ++
	TOK_MINUS_MINUS // --
	TOK_SHL         // <<
	TOK_SHR         // >>
	TOK_EQ          // ==
	TOK_NEQ         // !=
	TOK_LTE         // <=
	TOK_GTE         // >=
	TOK_AND_AND     // &&
	TOK_OR_OR       // ||
	TOK_PLUS_EQ     // +=
	TOK_MINUS_EQ    // -=
	TOK_STAR_EQ     // *=
	TOK_SLASH_EQ    // /=
	TOK_PERCENT_EQ  // %=
	TOK_AMP_EQ      // &=
	TOK_PIPE_EQ     // |=
	TOK_CARET_EQ    // ^=
	TOK_SHL_EQ      // <<=
	TOK_SHR_EQ      // >>=
	TOK_BACKTICK    // ` (power in HolyC)
	TOK_ELLIPSIS    // ...

	// Délimiteurs
	TOK_LPAREN    // (
	TOK_RPAREN    // )
	TOK_LBRACKET  // [
	TOK_RBRACKET  // ]
	TOK_LBRACE    // {
	TOK_RBRACE    // }
	TOK_SEMICOLON // ;
	TOK_COMMA     // ,
	TOK_COLON     // :
)

type Token struct {
	Type    TokenType
	Literal string
	IntVal  int64
	FloatVal float64
	Line    int
	Col     int
}

var keywords = map[string]TokenType{
	"U0": TOK_U0, "U8": TOK_U8, "U16": TOK_U16, "U32": TOK_U32, "U64": TOK_U64,
	"I8": TOK_I8, "I16": TOK_I16, "I32": TOK_I32, "I64": TOK_I64,
	"F64": TOK_F64, "Bool": TOK_BOOL,
	"if": TOK_IF, "else": TOK_ELSE,
	"while": TOK_WHILE, "do": TOK_DO, "for": TOK_FOR,
	"switch": TOK_SWITCH, "case": TOK_CASE, "default": TOK_DEFAULT,
	"break": TOK_BREAK, "return": TOK_RETURN,
	"class": TOK_CLASS, "union": TOK_UNION,
	"public": TOK_PUBLIC, "extern": TOK_EXTERN, "static": TOK_STATIC,
	"sizeof": TOK_SIZEOF,
	"try": TOK_TRY, "catch": TOK_CATCH,
	"goto": TOK_GOTO,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TOK_IDENT
}

func IsType(t TokenType) bool {
	return t >= TOK_U0 && t <= TOK_BOOL
}
