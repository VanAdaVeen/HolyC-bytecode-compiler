package main

import (
	"fmt"
	"os"
	"strings"
)

type Lexer struct {
	src     string
	pos     int
	line    int
	col     int
	ch      byte
	file    string
	errors  []string
}

func NewLexer(src, filename string) *Lexer {
	l := &Lexer{src: src, line: 1, col: 0, file: filename}
	l.advance()
	return l
}

func (l *Lexer) advance() {
	if l.pos >= len(l.src) {
		l.ch = 0
		return
	}
	l.ch = l.src[l.pos]
	l.pos++
	l.col++
	if l.ch == '\n' {
		l.line++
		l.col = 0
	}
}

func (l *Lexer) peek() byte {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *Lexer) errorf(format string, args ...any) {
	msg := fmt.Sprintf("%s:%d:%d: %s", l.file, l.line, l.col, fmt.Sprintf(format, args...))
	l.errors = append(l.errors, msg)
	fmt.Fprintln(os.Stderr, msg)
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' || l.ch == '\n' {
		l.advance()
	}
}

func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.advance()
	}
}

func (l *Lexer) skipBlockComment() {
	depth := 1
	for depth > 0 && l.ch != 0 {
		if l.ch == '/' && l.peek() == '*' {
			l.advance()
			l.advance()
			depth++
		} else if l.ch == '*' && l.peek() == '/' {
			l.advance()
			l.advance()
			depth--
		} else {
			l.advance()
		}
	}
}

func (l *Lexer) readNumber() Token {
	start := l.pos - 1
	line, col := l.line, l.col

	if l.ch == '0' && (l.peek() == 'x' || l.peek() == 'X') {
		l.advance() // skip 0
		l.advance() // skip x
		for isHexDigit(l.ch) {
			l.advance()
		}
		lit := l.src[start : l.pos-1]
		if l.ch != 0 && l.pos <= len(l.src) {
			lit = l.src[start : l.pos-1]
		}
		val := parseHex(l.src[start+2 : l.pos-1])
		return Token{TOK_INT, lit, val, 0, line, col}
	}

	if l.ch == '0' && (l.peek() == 'b' || l.peek() == 'B') {
		l.advance() // skip 0
		l.advance() // skip b
		for l.ch == '0' || l.ch == '1' {
			l.advance()
		}
		lit := l.src[start : l.pos-1]
		val := parseBin(l.src[start+2 : l.pos-1])
		return Token{TOK_INT, lit, val, 0, line, col}
	}

	for isDigit(l.ch) {
		l.advance()
	}

	isFloat := false
	if l.ch == '.' && isDigit(l.peek()) {
		isFloat = true
		l.advance()
		for isDigit(l.ch) {
			l.advance()
		}
	}
	if l.ch == 'e' || l.ch == 'E' {
		isFloat = true
		l.advance()
		if l.ch == '-' || l.ch == '+' {
			l.advance()
		}
		for isDigit(l.ch) {
			l.advance()
		}
	}

	end := l.pos - 1
	if l.ch != 0 {
		end = l.pos - 1
	}
	lit := l.src[start:end]

	if isFloat {
		fval := parseFloat(lit)
		return Token{TOK_FLOAT, lit, 0, fval, line, col}
	}
	ival := parseInt(lit)
	return Token{TOK_INT, lit, ival, 0, line, col}
}

func (l *Lexer) readIdent() Token {
	start := l.pos - 1
	line, col := l.line, l.col
	for isAlphaNumeric(l.ch) {
		l.advance()
	}
	end := l.pos - 1
	if l.ch != 0 {
		end = l.pos - 1
	}
	lit := l.src[start:end]
	return Token{LookupIdent(lit), lit, 0, 0, line, col}
}

func (l *Lexer) readString() Token {
	line, col := l.line, l.col
	l.advance() // skip opening "
	var sb strings.Builder
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.advance()
			switch l.ch {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case '\\':
				sb.WriteByte('\\')
			case '"':
				sb.WriteByte('"')
			case '0':
				sb.WriteByte(0)
			default:
				sb.WriteByte('\\')
				sb.WriteByte(l.ch)
			}
		} else {
			sb.WriteByte(l.ch)
		}
		l.advance()
	}
	l.advance() // skip closing "
	return Token{TOK_STRING, sb.String(), 0, 0, line, col}
}

func (l *Lexer) readCharConst() Token {
	line, col := l.line, l.col
	l.advance() // skip '
	var val int64
	shift := 0
	for l.ch != '\'' && l.ch != 0 && shift < 64 {
		var b byte
		if l.ch == '\\' {
			l.advance()
			switch l.ch {
			case 'n':
				b = '\n'
			case 't':
				b = '\t'
			case '0':
				b = 0
			case '\\':
				b = '\\'
			case '\'':
				b = '\''
			default:
				b = l.ch
			}
		} else {
			b = l.ch
		}
		val |= int64(b) << shift
		shift += 8
		l.advance()
	}
	l.advance() // skip closing '
	return Token{TOK_CHAR, "", val, 0, line, col}
}

func (l *Lexer) handlePreprocessor() Token {
	line, col := l.line, l.col
	l.advance() // skip #
	l.skipWhitespace()

	start := l.pos - 1
	for isAlpha(l.ch) {
		l.advance()
	}
	directive := l.src[start : l.pos-1]

	switch directive {
	case "include":
		l.skipWhitespace()
		if l.ch == '"' {
			tok := l.readString()
			return Token{TOK_INCLUDE, tok.Literal, 0, 0, line, col}
		}
	case "define":
		l.skipWhitespace()
		start = l.pos - 1
		for isAlphaNumeric(l.ch) {
			l.advance()
		}
		name := l.src[start : l.pos-1]
		l.skipWhitespace()
		valStart := l.pos - 1
		for l.ch != '\n' && l.ch != 0 {
			l.advance()
		}
		valEnd := l.pos - 1
		if l.ch != 0 {
			valEnd = l.pos - 1
		}
		_ = strings.TrimSpace(l.src[valStart:valEnd])
		return Token{TOK_DEFINE, name, 0, 0, line, col}
	}
	// Skip to end of line for unknown preprocessor directives
	for l.ch != '\n' && l.ch != 0 {
		l.advance()
	}
	return l.NextToken()
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	line, col := l.line, l.col

	if l.ch == 0 {
		return Token{TOK_EOF, "", 0, 0, line, col}
	}

	// Comments
	if l.ch == '/' {
		if l.peek() == '/' {
			l.skipLineComment()
			return l.NextToken()
		}
		if l.peek() == '*' {
			l.advance()
			l.advance()
			l.skipBlockComment()
			return l.NextToken()
		}
	}

	// Numbers
	if isDigit(l.ch) {
		return l.readNumber()
	}

	// Identifiers / keywords
	if isAlpha(l.ch) {
		return l.readIdent()
	}

	// Strings
	if l.ch == '"' {
		return l.readString()
	}

	// Char constants
	if l.ch == '\'' {
		return l.readCharConst()
	}

	// Preprocessor
	if l.ch == '#' {
		return l.handlePreprocessor()
	}

	// Operators and punctuation
	ch := l.ch
	l.advance()

	switch ch {
	case '+':
		if l.ch == '+' {
			l.advance()
			return Token{TOK_PLUS_PLUS, "++", 0, 0, line, col}
		}
		if l.ch == '=' {
			l.advance()
			return Token{TOK_PLUS_EQ, "+=", 0, 0, line, col}
		}
		return Token{TOK_PLUS, "+", 0, 0, line, col}
	case '-':
		if l.ch == '-' {
			l.advance()
			return Token{TOK_MINUS_MINUS, "--", 0, 0, line, col}
		}
		if l.ch == '=' {
			l.advance()
			return Token{TOK_MINUS_EQ, "-=", 0, 0, line, col}
		}
		if l.ch == '>' {
			l.advance()
			return Token{TOK_ARROW, "->", 0, 0, line, col}
		}
		return Token{TOK_MINUS, "-", 0, 0, line, col}
	case '*':
		if l.ch == '=' {
			l.advance()
			return Token{TOK_STAR_EQ, "*=", 0, 0, line, col}
		}
		return Token{TOK_STAR, "*", 0, 0, line, col}
	case '/':
		if l.ch == '=' {
			l.advance()
			return Token{TOK_SLASH_EQ, "/=", 0, 0, line, col}
		}
		return Token{TOK_SLASH, "/", 0, 0, line, col}
	case '%':
		if l.ch == '=' {
			l.advance()
			return Token{TOK_PERCENT_EQ, "%=", 0, 0, line, col}
		}
		return Token{TOK_PERCENT, "%", 0, 0, line, col}
	case '&':
		if l.ch == '&' {
			l.advance()
			return Token{TOK_AND_AND, "&&", 0, 0, line, col}
		}
		if l.ch == '=' {
			l.advance()
			return Token{TOK_AMP_EQ, "&=", 0, 0, line, col}
		}
		return Token{TOK_AMP, "&", 0, 0, line, col}
	case '|':
		if l.ch == '|' {
			l.advance()
			return Token{TOK_OR_OR, "||", 0, 0, line, col}
		}
		if l.ch == '=' {
			l.advance()
			return Token{TOK_PIPE_EQ, "|=", 0, 0, line, col}
		}
		return Token{TOK_PIPE, "|", 0, 0, line, col}
	case '^':
		if l.ch == '=' {
			l.advance()
			return Token{TOK_CARET_EQ, "^=", 0, 0, line, col}
		}
		return Token{TOK_CARET, "^", 0, 0, line, col}
	case '~':
		return Token{TOK_TILDE, "~", 0, 0, line, col}
	case '!':
		if l.ch == '=' {
			l.advance()
			return Token{TOK_NEQ, "!=", 0, 0, line, col}
		}
		return Token{TOK_BANG, "!", 0, 0, line, col}
	case '<':
		if l.ch == '<' {
			l.advance()
			if l.ch == '=' {
				l.advance()
				return Token{TOK_SHL_EQ, "<<=", 0, 0, line, col}
			}
			return Token{TOK_SHL, "<<", 0, 0, line, col}
		}
		if l.ch == '=' {
			l.advance()
			return Token{TOK_LTE, "<=", 0, 0, line, col}
		}
		return Token{TOK_LT, "<", 0, 0, line, col}
	case '>':
		if l.ch == '>' {
			l.advance()
			if l.ch == '=' {
				l.advance()
				return Token{TOK_SHR_EQ, ">>=", 0, 0, line, col}
			}
			return Token{TOK_SHR, ">>", 0, 0, line, col}
		}
		if l.ch == '=' {
			l.advance()
			return Token{TOK_GTE, ">=", 0, 0, line, col}
		}
		return Token{TOK_GT, ">", 0, 0, line, col}
	case '=':
		if l.ch == '=' {
			l.advance()
			return Token{TOK_EQ, "==", 0, 0, line, col}
		}
		return Token{TOK_ASSIGN, "=", 0, 0, line, col}
	case '`':
		return Token{TOK_BACKTICK, "`", 0, 0, line, col}
	case '.':
		if l.ch == '.' && l.peek() == '.' {
			l.advance()
			l.advance()
			return Token{TOK_ELLIPSIS, "...", 0, 0, line, col}
		}
		return Token{TOK_DOT, ".", 0, 0, line, col}
	case '(':
		return Token{TOK_LPAREN, "(", 0, 0, line, col}
	case ')':
		return Token{TOK_RPAREN, ")", 0, 0, line, col}
	case '[':
		return Token{TOK_LBRACKET, "[", 0, 0, line, col}
	case ']':
		return Token{TOK_RBRACKET, "]", 0, 0, line, col}
	case '{':
		return Token{TOK_LBRACE, "{", 0, 0, line, col}
	case '}':
		return Token{TOK_RBRACE, "}", 0, 0, line, col}
	case ';':
		return Token{TOK_SEMICOLON, ";", 0, 0, line, col}
	case ',':
		return Token{TOK_COMMA, ",", 0, 0, line, col}
	case ':':
		return Token{TOK_COLON, ":", 0, 0, line, col}
	}

	l.errorf("unexpected character: '%c' (0x%02X)", ch, ch)
	return l.NextToken()
}

// Helpers

func isDigit(ch byte) bool    { return ch >= '0' && ch <= '9' }
func isHexDigit(ch byte) bool { return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') }
func isAlpha(ch byte) bool    { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' }
func isAlphaNumeric(ch byte) bool { return isAlpha(ch) || isDigit(ch) }

func parseInt(s string) int64 {
	var val int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			val = val*10 + int64(c-'0')
		}
	}
	return val
}

func parseHex(s string) int64 {
	var val int64
	for _, c := range s {
		val <<= 4
		if c >= '0' && c <= '9' {
			val |= int64(c - '0')
		} else if c >= 'a' && c <= 'f' {
			val |= int64(c-'a') + 10
		} else if c >= 'A' && c <= 'F' {
			val |= int64(c-'A') + 10
		}
	}
	return val
}

func parseBin(s string) int64 {
	var val int64
	for _, c := range s {
		val <<= 1
		if c == '1' {
			val |= 1
		}
	}
	return val
}

func parseFloat(s string) float64 {
	var val float64
	fmt.Sscanf(s, "%f", &val)
	return val
}
