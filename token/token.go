// Package token contains constants which are used when lexing a program
// written in the monkey language, as done by the parser.
package token

// Type is a string
type Type string

// Token struct represent the lexer token
type Token struct {
	Type    Type
	Literal string
}

// pre-defined Type
const (
	AND             = "&&"
	ASSIGN          = "="
	ASTERISK        = "*"
	ASTERISK_EQUALS = "*="
	BACKTICK        = "`"
	BANG            = "!"
	CASE            = "case"
	COLON           = ":"
	COMMA           = ","
	CONST           = "CONST"
	CONTAINS        = "~="
	DEFAULT         = "DEFAULT"
	DEFINE_FUNCTION = "DEFINE_FUNCTION"
	DOTDOT          = ".."
	ELSE            = "ELSE"
	EOF             = "EOF"
	EQ              = "=="
	FALSE           = "FALSE"
	FLOAT           = "FLOAT"
	FOR             = "FOR"
	FOREACH         = "FOREACH"
	FUNCTION        = "FUNCTION"
	GT              = ">"
	GT_EQUALS       = ">="
	IDENT           = "IDENT"
	IF              = "IF"
	ILLEGAL         = "ILLEGAL"
	IN              = "IN"
	INT             = "INT"
	LBRACE          = "LBRACE"
	LBRACKET        = "["
	LET             = "LET"
	LPAREN          = "("
	LT              = "<"
	LT_EQUALS       = "<="
	MINUS           = "-"
	MINUS_EQUALS    = "-="
	MINUS_MINUS     = "--"
	MOD             = "%"
	NOT_CONTAINS    = "!~"
	NOT_EQ          = "!="
	NULL            = "null"
	OR              = "||"
	PERIOD          = "."
	PLUS            = "+"
	PLUS_EQUALS     = "+="
	PLUS_PLUS       = "++"
	POW             = "**"
	QUESTION        = "?"
	RBRACE          = "RBRACE"
	RBRACKET        = "]"
	REGEXP          = "REGEXP"
	RETURN          = "RETURN"
	RPAREN          = ")"
	SEMICOLON       = ";"
	SLASH           = "/"
	SLASH_EQUALS    = "/="
	STRING          = "STRING"
	SWITCH          = "switch"
	TRUE            = "TRUE"
)

// reversed keywords
var keywords = map[string]Type{
	"case":    CASE,
	"const":   CONST,
	"default": DEFAULT,
	"ELSE":    ELSE,
	"FALSE":   FALSE,
	"FN":      FUNCTION,
	"WHILE":   FOR,
	"foreach": FOREACH,
	"FUNC":    DEFINE_FUNCTION,
	"IF":      IF,
	"in":      IN,
	"LET":     LET,
	"NIL":     NULL,
	"RETURN":  RETURN,
	"switch":  SWITCH,
	"TRUE":    TRUE,
	"BEGIN":   LBRACE,
	"END":     RBRACE,
}

// LookupIdentifier used to determinate whether identifier is keyword nor not
func LookupIdentifier(identifier string) Type {
	if tok, ok := keywords[identifier]; ok {
		return tok
	}
	return IDENT
}
