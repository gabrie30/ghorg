package text

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

type wrapOpts struct {
	indent string
	pad    string
	align  Alignment
}

// WrapOption is a functional option for the Wrap() function
type WrapOption func(opts *wrapOpts)

// WrapPad configure the padding with a string for Wrap()
func WrapPad(pad string) WrapOption {
	return func(opts *wrapOpts) {
		opts.pad = pad
	}
}

// WrapPadded configure the padding with a number of space characters for Wrap()
func WrapPadded(padLen int) WrapOption {
	return func(opts *wrapOpts) {
		opts.pad = strings.Repeat(" ", padLen)
	}
}

// WrapPad configure the indentation on the first line for Wrap()
func WrapIndent(indent string) WrapOption {
	return func(opts *wrapOpts) {
		opts.indent = indent
	}
}

// WrapAlign configure the text alignment for Wrap()
func WrapAlign(align Alignment) WrapOption {
	return func(opts *wrapOpts) {
		opts.align = align
	}
}

// allWrapOpts compile the set of WrapOption into a final wrapOpts
// from the default values.
func allWrapOpts(opts []WrapOption) *wrapOpts {
	wrapOpts := &wrapOpts{
		indent: "",
		pad:    "",
		align:  NoAlign,
	}
	for _, opt := range opts {
		opt(wrapOpts)
	}
	if wrapOpts.indent == "" {
		wrapOpts.indent = wrapOpts.pad
	}
	return wrapOpts
}

// Wrap a text for a given line size.
// Handle properly terminal color escape code
// Options are accepted to configure things like indent, padding or alignment.
// Return the wrapped text and the number of lines
func Wrap(text string, lineWidth int, opts ...WrapOption) (string, int) {
	wrapOpts := allWrapOpts(opts)

	if lineWidth <= 0 {
		return "", 1
	}

	var result strings.Builder
	var state EscapeState
	nbLine := 0

	// output function to:
	// - set the endlines (same as strings.Join())
	// - reset and set again the escape state around the padding/indent
	output := func(padding string, content string) {
		zeroState := state.IsZero()
		if !zeroState && len(padding) > 0 {
			result.WriteString("\x1b[0m")
		}
		if nbLine > 0 {
			result.WriteString("\n")
		}
		result.WriteString(padding)
		if !zeroState && len(padding) > 0 {
			result.WriteString(state.FormatString())
		}
		result.WriteString(content)
		nbLine++
		state.Witness(content)
	}

	if Len(wrapOpts.indent) >= lineWidth {
		// indent is too wide, fallback rendering
		output(strings.Repeat("⭬", lineWidth), "")
		wrapOpts.indent = wrapOpts.pad
	}
	if Len(wrapOpts.pad) >= lineWidth {
		// padding is too wide, fallback rendering
		line := strings.Repeat("⭬", lineWidth)
		return strings.Repeat(line+"\n", 5), 6
	}

	// Start with the indent
	padStr := wrapOpts.indent
	padLen := Len(wrapOpts.indent)

	// tabs are formatted as 4 spaces
	text = strings.Replace(text, "\t", "    ", -1)

	// NOTE: text is first segmented into lines so that softwrapLine can handle individually
	for i, line := range strings.Split(text, "\n") {
		// on the second line, switch to use the padding instead
		if i == 1 {
			padStr = wrapOpts.pad
			padLen = Len(wrapOpts.pad)
		}

		if line == "" || strings.TrimSpace(line) == "" {
			// nothing in the line, we just add the non-empty part of the padding
			output(strings.TrimRight(padStr, " "), "")
			continue
		}

		wrapped := softwrapLine(line, lineWidth-padLen)
		split := strings.Split(wrapped, "\n")

		if i == 0 && len(split) > 1 {
			// the very first line got wrapped.
			// that means we need to use the indent, use the first wrapped line, discard the rest
			// switch to the normal padding, do the softwrap again with the remainder,
			// and fallback to the normal wrapping flow

			content := LineAlign(strings.TrimRight(split[0], " "), lineWidth-padLen, wrapOpts.align)
			output(padStr, content)

			line = strings.TrimPrefix(line, split[0])
			line = strings.TrimLeft(line, " ")

			padStr = wrapOpts.pad
			padLen = Len(wrapOpts.pad)

			wrapped = softwrapLine(line, lineWidth-padLen)
			split = strings.Split(wrapped, "\n")
		}

		for j, seg := range split {
			if j == 0 {
				// keep the left padding of the wrapped line
				content := LineAlign(strings.TrimRight(seg, " "), lineWidth-padLen, wrapOpts.align)
				output(padStr, content)
			} else {
				content := LineAlign(strings.TrimSpace(seg), lineWidth-padLen, wrapOpts.align)
				output(padStr, content)
			}
		}
	}

	return result.String(), nbLine
}

// WrapLeftPadded wrap a text for a given line size with a left padding.
// Handle properly terminal color escape code
func WrapLeftPadded(text string, lineWidth int, leftPad int) (string, int) {
	return Wrap(text, lineWidth, WrapPadded(leftPad))
}

// WrapWithPad wrap a text for a given line size with a custom left padding
// Handle properly terminal color escape code
func WrapWithPad(text string, lineWidth int, pad string) (string, int) {
	return Wrap(text, lineWidth, WrapPad(pad))
}

// WrapWithPad wrap a text for a given line size with a custom left padding
// This function also align the result depending on the requested alignment.
// Handle properly terminal color escape code
func WrapWithPadAlign(text string, lineWidth int, pad string, align Alignment) (string, int) {
	return Wrap(text, lineWidth, WrapPad(pad), WrapAlign(align))
}

// WrapWithPadIndent wrap a text for a given line size with a custom left padding
// and a first line indent. The padding is not effective on the first line, indent
// is used instead, which allow to implement indents and outdents.
// Handle properly terminal color escape code
func WrapWithPadIndent(text string, lineWidth int, indent string, pad string) (string, int) {
	return Wrap(text, lineWidth, WrapIndent(indent), WrapPad(pad))
}

// WrapWithPadIndentAlign wrap a text for a given line size with a custom left padding
// and a first line indent. The padding is not effective on the first line, indent
// is used instead, which allow to implement indents and outdents.
// This function also align the result depending on the requested alignment.
// Handle properly terminal color escape code
func WrapWithPadIndentAlign(text string, lineWidth int, indent string, pad string, align Alignment) (string, int) {
	return Wrap(text, lineWidth, WrapIndent(indent), WrapPad(pad), WrapAlign(align))
}

// Break a line into several lines so that each line consumes at most
// 'lineWidth' cells.  Lines break at groups of white spaces and multicell
// chars. Nothing is removed from the original text so that it behaves like a
// softwrap.
//
// Required: The line shall not contain '\n'
//
// WRAPPING ALGORITHM: The line is broken into non-breakable chunks, then line
// breaks ("\n") are inserted between these groups so that the total length
// between breaks does not exceed the required width. Words that are longer than
// the textWidth are broken into pieces no longer than textWidth.
func softwrapLine(line string, lineWidth int) string {
	escaped, escapes := ExtractTermEscapes(line)

	chunks := segmentLine(escaped)
	// Reverse the chunk array so we can use it as a stack.
	for i, j := 0, len(chunks)-1; i < j; i, j = i+1, j-1 {
		chunks[i], chunks[j] = chunks[j], chunks[i]
	}

	// for readability, minimal implementation of a stack:

	pop := func() string {
		result := chunks[len(chunks)-1]
		chunks = chunks[:len(chunks)-1]
		return result
	}

	push := func(chunk string) {
		chunks = append(chunks, chunk)
	}

	peek := func() string {
		return chunks[len(chunks)-1]
	}

	empty := func() bool {
		return len(chunks) == 0
	}

	var out strings.Builder

	// helper to write in the output while interleaving the escape
	// sequence at the correct places.
	// note: the final algorithm will add additional line break in the original
	// text. Those line break are *not* fed to this helper so the positions don't
	// need to be offset, which make the whole thing much easier.
	currPos := 0
	currItem := 0
	outputString := func(s string) {
		for _, r := range s {
			for currItem < len(escapes) && currPos == escapes[currItem].Pos {
				out.WriteString(escapes[currItem].Item)
				currItem++
			}
			out.WriteRune(r)
			currPos++
		}
	}

	width := 0

	for !empty() {
		wl := Len(peek())

		if width+wl <= lineWidth {
			// the chunk fit in the available space
			outputString(pop())
			width += wl
			if width == lineWidth && !empty() {
				// only add line break when there is more chunk to come
				out.WriteRune('\n')
				width = 0
			}
		} else if wl > lineWidth {
			// words too long for a full line are split to fill the remaining space.
			// But if the long words is the first non-space word in the middle of the
			// line, preceding spaces shall not be counted in word splitting.
			splitWidth := lineWidth - width
			if strings.HasSuffix(out.String(), "\n"+strings.Repeat(" ", width)) {
				splitWidth += width
			}
			left, right := splitWord(pop(), splitWidth)
			// remainder is pushed back to the stack for next round
			push(right)
			outputString(left)
			out.WriteRune('\n')
			width = 0
		} else {
			// normal line overflow, we add a line break and try again
			out.WriteRune('\n')
			width = 0
		}
	}

	// Don't forget the trailing escapes, if any.
	for currItem < len(escapes) && currPos >= escapes[currItem].Pos {
		out.WriteString(escapes[currItem].Item)
		currItem++
	}

	return out.String()
}

// Segment a line into chunks, where each chunk consists of chars with the same
// type and is not breakable.
func segmentLine(s string) []string {
	var chunks []string

	var word string
	wordType := none
	flushWord := func() {
		chunks = append(chunks, word)
		word = ""
		wordType = none
	}

	for _, r := range s {
		// A WIDE_CHAR itself constitutes a chunk.
		thisType := runeType(r)
		if thisType == wideChar {
			if wordType != none {
				flushWord()
			}
			chunks = append(chunks, string(r))
			continue
		}
		// Other type of chunks starts with a char of that type, and ends with a
		// char with different type or end of string.
		if thisType != wordType {
			if wordType != none {
				flushWord()
			}
			word = string(r)
			wordType = thisType
		} else {
			word += string(r)
		}
	}
	if word != "" {
		flushWord()
	}

	return chunks
}

type RuneType int

// Rune categories
//
// These categories are so defined that each category forms a non-breakable
// chunk. It IS NOT the same as unicode code point categories.
const (
	none RuneType = iota
	wideChar
	invisible
	shortUnicode
	space
	visibleAscii
)

// Determine the category of a rune.
func runeType(r rune) RuneType {
	rw := runewidth.RuneWidth(r)
	if rw > 1 {
		return wideChar
	} else if rw == 0 {
		return invisible
	} else if r > 127 {
		return shortUnicode
	} else if r == ' ' {
		return space
	} else {
		return visibleAscii
	}
}

// splitWord split a word at the given length, while ignoring the terminal escape sequences
func splitWord(word string, length int) (string, string) {
	runes := []rune(word)
	var result []rune
	added := 0
	escape := false

	if length == 0 {
		return "", word
	}

	for _, r := range runes {
		if r == '\x1b' {
			escape = true
		}

		width := runewidth.RuneWidth(r)
		if width+added > length {
			// wide character made the length overflow
			break
		}

		result = append(result, r)

		if !escape {
			added += width
			if added >= length {
				break
			}
		}

		if r == 'm' {
			escape = false
		}
	}

	leftover := runes[len(result):]

	return string(result), string(leftover)
}
