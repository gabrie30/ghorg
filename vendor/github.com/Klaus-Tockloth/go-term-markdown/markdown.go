package markdown

import (
	md "github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
)

// Extensions returns the bitmask of extensions supported by this renderer.
// The output of this function can be used to instantiate a new markdown
// parser using the `NewWithExtensions` function.
func Extensions() parser.Extensions {
	extensions := parser.NoIntraEmphasis        // Ignore emphasis markers inside words
	extensions |= parser.Tables                 // Parse tables
	extensions |= parser.FencedCode             // Parse fenced code blocks
	extensions |= parser.Autolink               // Detect embedded URLs that are not explicitly marked
	extensions |= parser.Strikethrough          // Strikethrough text using ~~test~~
	extensions |= parser.SpaceHeadings          // Be strict about prefix heading rules
	extensions |= parser.HeadingIDs             // specify heading IDs  with {#id}
	extensions |= parser.BackslashLineBreak     // Translate trailing backslashes into line breaks
	extensions |= parser.DefinitionLists        // Parse definition lists
	extensions |= parser.LaxHTMLBlocks          // more in HTMLBlock, less in HTMLSpan
	extensions |= parser.NoEmptyLineBeforeBlock // no need for new line before a list

	return extensions
}

func Render(source string, lineWidth int, leftPad int, opts ...Options) []byte {
	p := parser.NewWithExtensions(Extensions())
	nodes := md.Parse([]byte(source), p)
	renderer := NewRenderer(lineWidth, leftPad, opts...)

	return md.Render(nodes, renderer)
}
