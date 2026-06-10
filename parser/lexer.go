package parser

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/2389-research/dippin-lang/ir"
)

type TokenType int

const (
	TokenError TokenType = iota
	TokenEOF
	TokenNewline
	TokenIndent
	TokenOutdent
	TokenKeyword
	TokenIdentifier
	TokenOperator
	TokenLiteral
	TokenColon
	TokenComma
	TokenArrow
	TokenBackArrow
	TokenLParen
	TokenRParen
	TokenLBracket // '[' — bracket syntax not supported; triggers a parse error
	TokenRawBlock // Raw text block (multiline prompt/command content)
)

type Token struct {
	Type     TokenType
	Value    string
	Location ir.SourceLocation
}

func (t Token) String() string {
	return fmt.Sprintf("%v(%q)@%d:%d", t.Type, t.Value, t.Location.Line, t.Location.Column)
}

type Lexer struct {
	input string
	lines []string // original lines for raw extraction

	line        int
	col         int
	indentStack []int
	tokens      []Token
	tokenIdx    int
}

func NewLexer(input string, filename string) *Lexer {
	l := &Lexer{
		input:       input,
		lines:       strings.Split(input, "\n"),
		line:        1,
		col:         1,
		indentStack: []int{0},
	}
	l.lex(filename)
	return l
}

func (l *Lexer) NextToken() Token {
	if l.tokenIdx >= len(l.tokens) {
		return Token{Type: TokenEOF, Location: ir.SourceLocation{Line: l.line, Column: l.col}}
	}
	t := l.tokens[l.tokenIdx]
	l.tokenIdx++
	return t
}

func (l *Lexer) PeekToken() Token {
	if l.tokenIdx >= len(l.tokens) {
		return Token{Type: TokenEOF, Location: ir.SourceLocation{Line: l.line, Column: l.col}}
	}
	return l.tokens[l.tokenIdx]
}

// lineIndent returns the number of leading whitespace bytes in a line.
// Tabs and spaces each count as one byte. This is used for indentation
// tracking where the returned value is also used as a byte offset
// into the string (e.g., line[lineIndent(line):] to get content).
func lineIndent(line string) int {
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	return i
}

// isBlankOrComment returns true if a line is empty, whitespace-only, or a comment-only line.
// A comment-only line has only optional whitespace followed by #.
func isBlankOrComment(line string) bool {
	trimmed := strings.TrimSpace(line)
	return len(trimmed) == 0 || trimmed[0] == '#'
}

// findUnquotedHash scans a string for a # character that is not inside double quotes.
// Returns the index of the # or -1 if not found.
func findUnquotedHash(s string) int {
	inQuote := false
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			inQuote = !inQuote
		}
		if !inQuote && s[i] == '#' {
			return i
		}
	}
	return -1
}

// stripComment removes a trailing comment from a line, but only if # is preceded by
// whitespace (not inside a value). Does NOT strip # that starts the line content.
func stripComment(line string) string {
	trimmed := strings.TrimRight(line, " \t\r")
	idx := findUnquotedHash(trimmed)
	if idx < 0 {
		return trimmed
	}
	if isStrippableHash(trimmed, idx) {
		return trimmed[:idx]
	}
	return trimmed
}

// isStrippableHash returns true if the # at idx should be treated as a comment start.
func isStrippableHash(line string, idx int) bool {
	if idx == lineIndent(line) {
		return true
	}
	return idx > 0 && (line[idx-1] == ' ' || line[idx-1] == '\t')
}

func (l *Lexer) lex(filename string) {
	i := 0
	for i < len(l.lines) {
		advanced, newI := l.lexOneLine(i, filename)
		if advanced {
			i = newI
			continue
		}
		i++
	}
	l.emitRemainingOutdents(filename)
	l.tokens = append(l.tokens, Token{Type: TokenEOF, Location: ir.SourceLocation{File: filename, Line: l.line, Column: 1}})
}

// lexOneLine processes a single source line and returns whether it was fully
// handled (true with a new line index) or should just be skipped (false).
func (l *Lexer) lexOneLine(i int, filename string) (bool, int) {
	line := l.lines[i]
	l.line = i + 1
	l.col = 1
	trimmed := strings.TrimRight(line, " \t\r")

	if l.shouldSkipLine(trimmed) {
		return false, 0
	}

	trimmed = stripComment(trimmed)
	if len(strings.TrimSpace(trimmed)) == 0 {
		return false, 0
	}

	indent := lineIndent(trimmed)
	content := trimmed[indent:]

	l.emitIndentTokens(indent, filename)

	if isKeyColonLine(content) {
		return l.lexKeyColonBlock(i, indent, content, line, filename)
	}

	l.lexLine(content, filename)
	l.tokens = append(l.tokens, Token{Type: TokenNewline, Location: ir.SourceLocation{File: filename, Line: l.line, Column: len(line) + 1}})
	return false, 0
}

// shouldSkipLine returns true if the line is blank/comment and should be skipped.
func (l *Lexer) shouldSkipLine(trimmed string) bool {
	return isBlankOrComment(trimmed)
}

// emitIndentTokens emits indent/outdent tokens based on the current indentation level.
func (l *Lexer) emitIndentTokens(indent int, filename string) {
	currIndent := l.indentStack[len(l.indentStack)-1]
	if indent > currIndent {
		l.indentStack = append(l.indentStack, indent)
		l.tokens = append(l.tokens, Token{Type: TokenIndent, Location: ir.SourceLocation{File: filename, Line: l.line, Column: 1}})
		return
	}
	for indent < currIndent && len(l.indentStack) > 1 {
		l.indentStack = l.indentStack[:len(l.indentStack)-1]
		l.tokens = append(l.tokens, Token{Type: TokenOutdent, Location: ir.SourceLocation{File: filename, Line: l.line, Column: 1}})
		currIndent = l.indentStack[len(l.indentStack)-1]
	}
}

// emitRemainingOutdents emits outdent tokens for any remaining indentation levels.
func (l *Lexer) emitRemainingOutdents(filename string) {
	for len(l.indentStack) > 1 {
		l.indentStack = l.indentStack[:len(l.indentStack)-1]
		l.tokens = append(l.tokens, Token{Type: TokenOutdent, Location: ir.SourceLocation{File: filename, Line: l.line, Column: 1}})
	}
}

// lexKeyColonBlock handles a "key:" line, potentially followed by a multiline block.
func (l *Lexer) lexKeyColonBlock(i, indent int, content, line, filename string) (bool, int) {
	keyEnd := strings.Index(content, ":")
	key := content[:keyEnd]
	loc := ir.SourceLocation{File: filename, Line: l.line, Column: indent + 1}
	l.tokens = append(l.tokens, Token{Type: TokenIdentifier, Value: key, Location: loc})
	l.tokens = append(l.tokens, Token{Type: TokenColon, Value: ":", Location: ir.SourceLocation{File: filename, Line: l.line, Column: indent + keyEnd + 1}})

	blockEnd, ok := l.tryCollectBlock(i, indent, filename)
	if ok {
		return true, blockEnd
	}

	// No multiline block follows — just a key: with empty value
	l.tokens = append(l.tokens, Token{Type: TokenNewline, Location: ir.SourceLocation{File: filename, Line: l.line, Column: len(line) + 1}})
	return false, 0
}

// tryCollectBlock looks ahead for an indented block after a key: line.
// Returns (blockEnd, true) if a block was found and collected, or (0, false) otherwise.
func (l *Lexer) tryCollectBlock(i, indent int, filename string) (int, bool) {
	nextContentLine := l.findNextContentLine(i + 1)
	if nextContentLine >= len(l.lines) {
		return 0, false
	}
	nextIndent := lineIndent(l.lines[nextContentLine])
	if nextIndent <= indent {
		return 0, false
	}

	blockEnd := l.findBlockEnd(nextContentLine, indent)
	rawText := l.extractRawBlock(nextContentLine, blockEnd, nextIndent)
	l.tokens = append(l.tokens, Token{Type: TokenNewline, Location: ir.SourceLocation{File: filename, Line: l.line, Column: 1}})
	l.tokens = append(l.tokens, Token{Type: TokenRawBlock, Value: rawText, Location: ir.SourceLocation{File: filename, Line: nextContentLine + 1, Column: nextIndent + 1}})
	l.tokens = append(l.tokens, Token{Type: TokenNewline, Location: ir.SourceLocation{File: filename, Line: blockEnd + 1, Column: 1}})
	return blockEnd, true
}

// findNextContentLine finds the next non-blank line starting from idx.
func (l *Lexer) findNextContentLine(idx int) int {
	for idx < len(l.lines) {
		nl := strings.TrimRight(l.lines[idx], " \t\r")
		if len(strings.TrimSpace(nl)) != 0 {
			return idx
		}
		idx++
	}
	return idx
}

// findBlockEnd finds the end of a multiline block (where indentation drops).
func (l *Lexer) findBlockEnd(start, indent int) int {
	blockEnd := start
	for blockEnd < len(l.lines) {
		bl := l.lines[blockEnd]
		blTrimmed := strings.TrimRight(bl, " \t\r")
		if len(strings.TrimSpace(blTrimmed)) == 0 {
			blockEnd++
			continue
		}
		if lineIndent(blTrimmed) <= indent {
			break
		}
		blockEnd++
	}
	return blockEnd
}

// isKeyColonLine checks if the line content (after indent) is just "identifier:"
// with nothing after the colon (or only whitespace).
func isKeyColonLine(content string) bool {
	colonIdx := strings.Index(content, ":")
	if colonIdx <= 0 {
		return false
	}
	key := content[:colonIdx]
	for _, ch := range key {
		if !isIdentRune(ch) {
			return false
		}
	}
	after := strings.TrimSpace(content[colonIdx+1:])
	return len(after) == 0
}

// isIdentRune returns true if a rune is valid in an identifier (alphanumeric or underscore).
func isIdentRune(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'
}

// extractRawBlock extracts raw text from original lines, stripping indent prefix.
// startIdx and endIdx are 0-based line indices.
func (l *Lexer) extractRawBlock(startIdx, endIdx, indent int) string {
	var result []string
	for i := startIdx; i < endIdx && i < len(l.lines); i++ {
		result = append(result, stripIndentPrefix(l.lines[i], indent))
	}
	return trimTrailingBlankLines(result)
}

// stripIndentPrefix removes up to `indent` whitespace bytes from the front of a line.
func stripIndentPrefix(line string, indent int) string {
	line = strings.TrimRight(line, "\r")
	j := 0
	for j < len(line) && j < indent && (line[j] == ' ' || line[j] == '\t') {
		j++
	}
	return line[j:]
}

// trimTrailingBlankLines removes trailing blank lines from a slice and joins the rest.
func trimTrailingBlankLines(lines []string) string {
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

func (l *Lexer) lexLine(line string, filename string) {
	i := 0
	colOffset := l.col + (l.indentStack[len(l.indentStack)-1])

	for i < len(line) {
		i = skipWhitespace(line, i)
		if i >= len(line) {
			break
		}
		loc := ir.SourceLocation{File: filename, Line: l.line, Column: colOffset + i}
		i = l.lexOneToken(line, i, loc)
	}
}

// lexOneToken tries each token type and returns the new position.
func (l *Lexer) lexOneToken(line string, i int, loc ir.SourceLocation) int {
	lexers := []func(string, int, ir.SourceLocation) (int, bool){
		l.tryLexPunctuation,
		l.tryLexArrow,
		l.tryLexOperator,
		l.tryLexQuotedString,
		l.tryLexIdentifier,
	}
	for _, fn := range lexers {
		if newI, ok := fn(line, i, loc); ok {
			return newI
		}
	}
	return i + 1
}

// skipWhitespace advances the index past any whitespace characters.
func skipWhitespace(line string, i int) int {
	for i < len(line) && unicode.IsSpace(rune(line[i])) {
		i++
	}
	return i
}

// punctuationTokens maps single-character punctuation to token types.
var punctuationTokens = map[byte]TokenType{
	':': TokenColon,
	',': TokenComma,
	'(': TokenLParen,
	')': TokenRParen,
	'[': TokenLBracket,
}

// tryLexPunctuation handles single-character punctuation: : , ( ) [
func (l *Lexer) tryLexPunctuation(line string, i int, loc ir.SourceLocation) (int, bool) {
	tokType, ok := punctuationTokens[line[i]]
	if !ok {
		return i, false
	}
	l.tokens = append(l.tokens, Token{Type: tokType, Value: string(line[i]), Location: loc})
	return i + 1, true
}

// tryLexArrow handles two-character arrows: -> and <-
func (l *Lexer) tryLexArrow(line string, i int, loc ir.SourceLocation) (int, bool) {
	if strings.HasPrefix(line[i:], "->") {
		l.tokens = append(l.tokens, Token{Type: TokenArrow, Value: "->", Location: loc})
		return i + 2, true
	}
	if strings.HasPrefix(line[i:], "<-") {
		l.tokens = append(l.tokens, Token{Type: TokenBackArrow, Value: "<-", Location: loc})
		return i + 2, true
	}
	return i, false
}

// twoCharOperators is the set of valid two-character operators.
var twoCharOperators = map[string]bool{
	"==": true, "!=": true, "<=": true, ">=": true,
}

// isOperatorChar returns true if the byte is a valid operator start character.
func isOperatorChar(ch byte) bool {
	return ch == '=' || ch == '!' || ch == '<' || ch == '>'
}

// tryLexOperator handles comparison operators: ==, !=, <=, >=, =, <, >, !
func (l *Lexer) tryLexOperator(line string, i int, loc ir.SourceLocation) (int, bool) {
	if !isOperatorChar(line[i]) {
		return i, false
	}
	if i+1 < len(line) && twoCharOperators[line[i:i+2]] {
		l.tokens = append(l.tokens, Token{Type: TokenOperator, Value: line[i : i+2], Location: loc})
		return i + 2, true
	}
	l.tokens = append(l.tokens, Token{Type: TokenOperator, Value: string(line[i]), Location: loc})
	return i + 1, true
}

// tryLexQuotedString handles double-quoted string literals with escape sequences.
func (l *Lexer) tryLexQuotedString(line string, i int, loc ir.SourceLocation) (int, bool) {
	if line[i] != '"' {
		return i, false
	}
	content, newI := readQuotedContent(line, i+1)
	l.tokens = append(l.tokens, Token{Type: TokenLiteral, Value: content, Location: loc})
	return newI, true
}

// readQuotedContent reads characters from line[start:] until an unescaped closing quote.
// Returns the content string and the position after the closing quote.
func readQuotedContent(line string, start int) (string, int) {
	var content strings.Builder
	i := start
	for i < len(line) && line[i] != '"' {
		i += appendQuotedChar(&content, line, i)
	}
	if i < len(line) {
		i++ // skip closing quote
	}
	return content.String(), i
}

// appendQuotedChar appends one character (handling escapes) and returns how many bytes were consumed.
func appendQuotedChar(b *strings.Builder, line string, i int) int {
	if line[i] == '\\' && i+1 < len(line) {
		b.WriteByte(line[i+1])
		return 2
	}
	b.WriteByte(line[i])
	return 1
}

// tryLexIdentifier handles alphanumeric identifiers including _, -, ., /
func (l *Lexer) tryLexIdentifier(line string, i int, loc ir.SourceLocation) (int, bool) {
	if !isAlphaNum(line[i]) {
		return i, false
	}
	start := i
	for i < len(line) && isIdentWordChar(line[i]) {
		i++
	}
	l.tokens = append(l.tokens, Token{Type: TokenIdentifier, Value: line[start:i], Location: loc})
	return i, true
}

// isIdentWordChar returns true if the byte is valid within an identifier word.
func isIdentWordChar(ch byte) bool {
	return isAlphaNum(ch) || ch == '_' || ch == '-' || ch == '.' || ch == '/'
}

// RawValueText extracts the raw value text from a line, starting after the colon
// following the field name. Used for single-line values like "fidelity: summary:medium".
// Strips inline comments (# preceded by whitespace) from the extracted value.
func (l *Lexer) RawValueText(lineNum int) string {
	if lineNum < 1 || lineNum > len(l.lines) {
		return ""
	}
	line := l.lines[lineNum-1]
	line = strings.TrimRight(line, " \t\r")
	colonIdx := strings.Index(line, ":")
	if colonIdx < 0 {
		return ""
	}
	val := strings.TrimSpace(line[colonIdx+1:])
	return maybeStripComment(val)
}

// maybeStripComment strips an inline comment from val unless val starts with # or ".
func maybeStripComment(val string) string {
	if len(val) == 0 || val[0] == '#' || val[0] == '"' || val[0] == '\'' {
		return val
	}
	return strings.TrimSpace(stripComment(val))
}

func isAlphaNum(ch byte) bool {
	return isLetter(ch) || isDigit(ch)
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
