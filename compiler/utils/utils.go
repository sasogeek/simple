//package utils
//
//import (
//	"simple/lexer"
//)
//
//func Tokenize(l *lexer.Lexer) []lexer.Token {
//	// Collect all tokens
//	var tokens []lexer.Token
//	for {
//		tok := l.NextToken()
//		tokens = append(tokens, tok)
//		l.UpdateLastTwoTokens(tok)
//		//fmt.Println(tok)
//
//		// Always append NEWLINE and handle indentation if necessary
//		if tok.Type == lexer.TokenNewline {
//			// If the previous two tokens were COLON followed by NEWLINE, expect an INDENT
//			if l.LastTwoTokens[0].Type == lexer.TokenColon && l.LastTwoTokens[1].Type == lexer.TokenNewline {
//				indentToken := l.HandleIndentAfterColonNewline()
//				tokens = append(tokens, indentToken)
//			} else {
//				indentOrDedent := l.HandleNewline()
//				if indentOrDedent.Type != lexer.TokenNewline {
//					tokens = append(tokens, indentOrDedent)
//				}
//			}
//		}
//
//		// Stop if we reach EOF
//		if tok.Type == lexer.TokenEOF {
//			break
//		}
//	}
//	return tokens
//}

package utils

import (
	"github.com/sasogeek/simple/lexer"
)

// Tokenize collects all tokens from the lexer.
func Tokenize(l *lexer.Lexer) []lexer.Token {
	var tokens []lexer.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == lexer.TokenEOF {
			break
		}
	}
	return tokens
}
