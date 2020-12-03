package lexer_rewrite

import (
	"fmt"
	"testing"
)

func newToken(t TokenType, v string) Token {
	return Token{Type: t, Value: v}
}

type testCase struct {
	input  string
	output []Token
}

var testCases = []testCase{
	{"add\n", []Token{newToken(TokenTypeText, "add"), newToken(TokenTypeNewLine, "\n"), newToken(TokenTypeEndOfFile, "")}},
	{"add 5\n", []Token{newToken(TokenTypeText, "add"), newToken(TokenTypeInteger, "5"), newToken(TokenTypeNewLine, "\n"), newToken(TokenTypeEndOfFile, "")}},
	{"add 0x5f\n", []Token{newToken(TokenTypeText, "add"), newToken(TokenTypeInteger, "0x5f"), newToken(TokenTypeNewLine, "\n"), newToken(TokenTypeEndOfFile, "")}},
	{"add 0b101\n", []Token{newToken(TokenTypeText, "add"), newToken(TokenTypeInteger, "0b101"), newToken(TokenTypeNewLine, "\n"), newToken(TokenTypeEndOfFile, "")}},
	{"move 23\n", []Token{newToken(TokenTypeText, "move"), newToken(TokenTypeInteger, "23"), newToken(TokenTypeNewLine, "\n"), newToken(TokenTypeEndOfFile, "")}},
	{"move label\n", []Token{newToken(TokenTypeText, "move"), newToken(TokenTypeText, "label"), newToken(TokenTypeNewLine, "\n"), newToken(TokenTypeEndOfFile, "")}},
	{
		"move 11, A\n",
		[]Token{
			newToken(TokenTypeText, "move"),
			newToken(TokenTypeInteger, "11"),
			newToken(TokenTypeComma, ","),
			newToken(TokenTypeText, "A"),
			newToken(TokenTypeNewLine, "\n"),
			newToken(TokenTypeEndOfFile, ""),
		},
	},
	{"label:\n", []Token{newToken(TokenTypeLabel, "label"), newToken(TokenTypeNewLine, "\n"), newToken(TokenTypeEndOfFile, "")}},
	{
		"move (fp)\n",
		[]Token{
			newToken(TokenTypeText, "move"),
			newToken(TokenTypeLeftParen, "("),
			newToken(TokenTypeText, "fp"),
			newToken(TokenTypeRightParen, ")"),
			newToken(TokenTypeNewLine, "\n"),
			newToken(TokenTypeEndOfFile, ""),
		},
	},
	{
		"move (label[3])\n",
		[]Token{
			newToken(TokenTypeText, "move"),
			newToken(TokenTypeLeftParen, "("),
			newToken(TokenTypeText, "label"),
			newToken(TokenTypeLeftBracket, "["),
			newToken(TokenTypeInteger, "3"),
			newToken(TokenTypeRightBracket, "]"),
			newToken(TokenTypeRightParen, ")"),
			newToken(TokenTypeNewLine, "\n"),
			newToken(TokenTypeEndOfFile, ""),
		},
	},
	{
		"move (fp+1)\n",
		[]Token{
			newToken(TokenTypeText, "move"),
			newToken(TokenTypeLeftParen, "("),
			newToken(TokenTypeText, "fp"),
			newToken(TokenTypePlus, "+"),
			newToken(TokenTypeInteger, "1"),
			newToken(TokenTypeRightParen, ")"),
			newToken(TokenTypeNewLine, "\n"),
			newToken(TokenTypeEndOfFile, ""),
		},
	},
	{
		"move (fp+1), B\n",
		[]Token{
			newToken(TokenTypeText, "move"),
			newToken(TokenTypeLeftParen, "("),
			newToken(TokenTypeText, "fp"),
			newToken(TokenTypePlus, "+"),
			newToken(TokenTypeInteger, "1"),
			newToken(TokenTypeRightParen, ")"),
			newToken(TokenTypeComma, ","),
			newToken(TokenTypeText, "B"),
			newToken(TokenTypeNewLine, "\n"),
			newToken(TokenTypeEndOfFile, ""),
		},
	},
	{
		"move A, 67\nadd 13\n",
		[]Token{
			newToken(TokenTypeText, "move"),
			newToken(TokenTypeText, "A"),
			newToken(TokenTypeComma, ","),
			newToken(TokenTypeInteger, "67"),
			newToken(TokenTypeNewLine, "\n"),
			newToken(TokenTypeText, "add"),
			newToken(TokenTypeInteger, "13"),
			newToken(TokenTypeNewLine, "\n"),
			newToken(TokenTypeEndOfFile, ""),
		},
	},
}

func TestLexer(t *testing.T) {
	for _, tc := range testCases {
		l := NewLexer()
		l.Load(tc.input)
		tokens, err := l.Run()
		if err != nil {
			t.Error(err)
			return
		}

		fmt.Println(tokens)
		fmt.Println()

		if len(tc.output) != len(tokens) {
			t.Errorf("expected %v tokens and got %v", len(tc.output), len(tokens))
			return
		}

		for i, token := range tc.output {
			if token.Type != tokens[i].Type {
				t.Errorf("expected type %v and got %v", token.Type, tokens[i].Type)
			}

			if token.Value != tokens[i].Value {
				t.Errorf("expected value %s and got %s for token type %v", token.Value, tokens[i].Value, token.Type)
			}
		}
	}
}