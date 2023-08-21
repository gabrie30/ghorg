package markdown

import (
	"bytes"
	"fmt"
	stdcolor "image/color"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/MichaelMure/go-term-text"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/eliukblau/pixterm/pkg/ansimage"
	"github.com/fatih/color"
	md "github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/kyokomi/emoji/v2"
	"golang.org/x/net/html"

	htmlWalker "github.com/MichaelMure/go-term-markdown/html"
)

/*

Here are the possible cases for the AST. You can render it using PlantUML.

@startuml

(*) --> Document
BlockQuote --> BlockQuote
BlockQuote --> CodeBlock
BlockQuote --> List
BlockQuote --> Paragraph
Del --> Emph
Del --> Strong
Del --> Text
Document --> BlockQuote
Document --> CodeBlock
Document --> Heading
Document --> HorizontalRule
Document --> HTMLBlock
Document --> List
Document --> Paragraph
Document --> Table
Emph --> Text
Heading --> Code
Heading --> Del
Heading --> Emph
Heading --> HTMLSpan
Heading --> Image
Heading --> Link
Heading --> Strong
Heading --> Text
Image --> Text
Link --> Image
Link --> Text
ListItem --> List
ListItem --> Paragraph
List --> ListItem
Paragraph --> Code
Paragraph --> Del
Paragraph --> Emph
Paragraph --> Hardbreak
Paragraph --> HTMLSpan
Paragraph --> Image
Paragraph --> Link
Paragraph --> Strong
Paragraph --> Text
Strong --> Emph
Strong --> Text
TableBody --> TableRow
TableCell --> Code
TableCell --> Del
TableCell --> Emph
TableCell --> HTMLSpan
TableCell --> Image
TableCell --> Link
TableCell --> Strong
TableCell --> Text
TableHeader --> TableRow
TableRow --> TableCell
Table --> TableBody
Table --> TableHeader

@enduml

*/

var _ md.Renderer = &renderer{}

type renderer struct {
	// maximum line width allowed
	lineWidth int
	// constant left padding to apply
	leftPad int

	// Dithering mode for ansimage
	// Default is fine directly through a terminal
	// DitheringWithBlocks is recommended if a terminal UI library is used
	imageDithering ansimage.DitheringMode

	// all the custom left paddings, without the fixed space from leftPad
	padAccumulator []string

	// one-shot indent for the first line of the inline content
	indent string

	// for Heading, Paragraph, HTMLBlock and TableCell, accumulate the content of
	// the child nodes (Link, Text, Image, formatting ...). The result
	// is then rendered appropriately when exiting the node.
	inlineAccumulator strings.Builder

	// record and render the heading numbering
	headingNumbering headingNumbering
	headingShade     levelShadeFmt

	blockQuoteLevel int
	blockQuoteShade levelShadeFmt

	table *tableRenderer
}

/// NewRenderer creates a new instance of the console renderer
func NewRenderer(lineWidth int, leftPad int, opts ...Options) *renderer {
	r := &renderer{
		lineWidth:       lineWidth,
		leftPad:         leftPad,
		padAccumulator:  make([]string, 0, 10),
		headingShade:    shade(defaultHeadingShades),
		blockQuoteShade: shade(defaultQuoteShades),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *renderer) pad() string {
	return strings.Repeat(" ", r.leftPad) + strings.Join(r.padAccumulator, "")
}

func (r *renderer) addPad(pad string) {
	r.padAccumulator = append(r.padAccumulator, pad)
}

func (r *renderer) popPad() {
	r.padAccumulator = r.padAccumulator[:len(r.padAccumulator)-1]
}

func (r *renderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	// TODO: remove
	// if node.AsLeaf() != nil {
	// 	fmt.Printf("%T, %v (%s)\n", node, entering, string(node.AsLeaf().Literal))
	// } else {
	// 	fmt.Printf("%T, %v\n", node, entering)
	// }

	switch node := node.(type) {
	case *ast.Document:
		// Nothing to do

	case *ast.BlockQuote:
		// set and remove a colored bar on the left
		if entering {
			r.blockQuoteLevel++
			r.addPad(r.blockQuoteShade(r.blockQuoteLevel)("┃ "))
		} else {
			r.blockQuoteLevel--
			r.popPad()
		}

	case *ast.List:
		// extra new line at the end of a list *if* next is not a list
		if next := ast.GetNextNode(node); !entering && next != nil {
			_, parentIsListItem := node.GetParent().(*ast.ListItem)
			_, nextIsList := next.(*ast.List)
			if !nextIsList && !parentIsListItem {
				_, _ = fmt.Fprintln(w)
			}
		}

	case *ast.ListItem:
		// write the prefix, add a padding if needed, and let Paragraph handle the rest
		if entering {
			switch {
			// numbered list
			case node.ListFlags&ast.ListTypeOrdered != 0:
				itemNumber := 1
				siblings := node.GetParent().GetChildren()
				for _, sibling := range siblings {
					if sibling == node {
						break
					}
					itemNumber++
				}
				prefix := fmt.Sprintf("%d. ", itemNumber)
				r.indent = r.pad() + Green(prefix)
				r.addPad(strings.Repeat(" ", text.Len(prefix)))

			// header of a definition
			case node.ListFlags&ast.ListTypeTerm != 0:
				r.inlineAccumulator.WriteString(greenOn)

			// content of a definition
			case node.ListFlags&ast.ListTypeDefinition != 0:
				r.addPad("  ")

			// no flags means it's the normal bullet point list
			default:
				r.indent = r.pad() + Green("• ")
				r.addPad("  ")
			}
		} else {
			switch {
			// numbered list
			case node.ListFlags&ast.ListTypeOrdered != 0:
				r.popPad()

			// header of a definition
			case node.ListFlags&ast.ListTypeTerm != 0:
				r.inlineAccumulator.WriteString(colorOff)

			// content of a definition
			case node.ListFlags&ast.ListTypeDefinition != 0:
				r.popPad()
				_, _ = fmt.Fprintln(w)

			// no flags means it's the normal bullet point list
			default:
				r.popPad()
			}
		}

	case *ast.Paragraph:
		// on exiting, collect and format the accumulated content
		if !entering {
			content := r.inlineAccumulator.String()
			r.inlineAccumulator.Reset()

			var out string
			if r.indent != "" {
				out, _ = text.WrapWithPadIndent(content, r.lineWidth, r.indent, r.pad())
				r.indent = ""
			} else {
				out, _ = text.WrapWithPad(content, r.lineWidth, r.pad())
			}
			_, _ = fmt.Fprint(w, out, "\n")

			// extra line break in some cases
			if next := ast.GetNextNode(node); next != nil {
				switch next.(type) {
				case *ast.Paragraph, *ast.Heading, *ast.HorizontalRule,
					*ast.CodeBlock, *ast.HTMLBlock:
					_, _ = fmt.Fprintln(w)
				}
			}
		}

	case *ast.Heading:
		if !entering {
			r.renderHeading(w, node.Level)
		}

	case *ast.HorizontalRule:
		r.renderHorizontalRule(w)

	case *ast.Emph:
		if entering {
			r.inlineAccumulator.WriteString(italicOn)
		} else {
			r.inlineAccumulator.WriteString(italicOff)
		}

	case *ast.Strong:
		if entering {
			r.inlineAccumulator.WriteString(boldOn)
		} else {
			// This is super silly but some terminals, instead of having
			// the ANSI code SGR 21 do "bold off" like the logic would guide,
			// do "double underline" instead. This is madness.

			// To resolve that problem, we take a snapshot of the escape state,
			// remove the bold, then output "reset all" + snapshot
			es := text.EscapeState{}
			es.Witness(r.inlineAccumulator.String())
			es.Bold = false
			r.inlineAccumulator.WriteString(resetAll)
			r.inlineAccumulator.WriteString(es.FormatString())
		}

	case *ast.Del:
		if entering {
			r.inlineAccumulator.WriteString(crossedOutOn)
		} else {
			r.inlineAccumulator.WriteString(crossedOutOff)
		}

	case *ast.Link:
		if entering {
			r.inlineAccumulator.WriteString("[")
			r.inlineAccumulator.WriteString(string(ast.GetFirstChild(node).AsLeaf().Literal))
			r.inlineAccumulator.WriteString("](")
			r.inlineAccumulator.WriteString(Blue(string(node.Destination)))
			if len(node.Title) > 0 {
				r.inlineAccumulator.WriteString(" ")
				r.inlineAccumulator.WriteString(string(node.Title))
			}
			r.inlineAccumulator.WriteString(")")
			return ast.SkipChildren
		}

	case *ast.Image:
		if entering {
			// the alt text/title is weirdly parsed and is actually
			// a child text of this node
			var title string
			if len(node.Children) == 1 {
				if t, ok := node.Children[0].(*ast.Text); ok {
					title = string(t.Literal)
				}
			}

			str, rendered := r.renderImage(
				string(node.Destination), title,
				r.lineWidth-r.leftPad,
			)

			if rendered {
				r.inlineAccumulator.WriteString("\n")
				r.inlineAccumulator.WriteString(str)
				r.inlineAccumulator.WriteString("\n\n")
			} else {
				r.inlineAccumulator.WriteString(str)
				r.inlineAccumulator.WriteString("\n")
			}

			return ast.SkipChildren
		}

	case *ast.Text:
		if string(node.Literal) == "\n" {
			break
		}
		content := string(node.Literal)
		if shouldCleanText(node) {
			content = removeLineBreak(content)
		}
		// emoji support !
		emojed := emoji.Sprint(content)
		r.inlineAccumulator.WriteString(emojed)

	case *ast.HTMLBlock:
		r.renderHTMLBlock(w, node)

	case *ast.CodeBlock:
		r.renderCodeBlock(w, node)

	case *ast.Softbreak:
		// not actually implemented in gomarkdown
		r.inlineAccumulator.WriteString("\n")

	case *ast.Hardbreak:
		r.inlineAccumulator.WriteString("\n")

	case *ast.Code:
		r.inlineAccumulator.WriteString(BlueBgItalic(string(node.Literal)))

	case *ast.HTMLSpan:
		r.inlineAccumulator.WriteString(Red(string(node.Literal)))

	case *ast.Table:
		if entering {
			r.table = newTableRenderer()
		} else {
			r.table.Render(w, r.leftPad, r.lineWidth)
			r.table = nil
		}

	case *ast.TableCell:
		if !entering {
			content := r.inlineAccumulator.String()
			r.inlineAccumulator.Reset()

			align := CellAlignLeft
			switch node.Align {
			case ast.TableAlignmentRight:
				align = CellAlignRight
			case ast.TableAlignmentCenter:
				align = CellAlignCenter
			}

			if node.IsHeader {
				r.table.AddHeaderCell(content, align)
			} else {
				r.table.AddBodyCell(content, CellAlignCopyHeader)
			}
		}

	case *ast.TableHeader, *ast.TableBody, *ast.TableFooter:
		// nothing to do

	case *ast.TableRow:
		if _, ok := node.Parent.(*ast.TableBody); ok && entering {
			r.table.NextBodyRow()
		}
		if _, ok := node.Parent.(*ast.TableFooter); ok && entering {
			r.table.NextBodyRow()
		}

	default:
		panic(fmt.Sprintf("Unknown node type %T", node))
	}

	return ast.GoToNext
}

func (*renderer) RenderHeader(w io.Writer, node ast.Node) {}

func (*renderer) RenderFooter(w io.Writer, node ast.Node) {}

func (r *renderer) renderHorizontalRule(w io.Writer) {
	_, _ = fmt.Fprintf(w, "%s%s\n\n", r.pad(), strings.Repeat("─", r.lineWidth-r.leftPad))
}

func (r *renderer) renderHeading(w io.Writer, level int) {
	content := r.inlineAccumulator.String()
	r.inlineAccumulator.Reset()

	// render the full line with the headingNumbering
	r.headingNumbering.Observe(level)
	content = fmt.Sprintf("%s %s", r.headingNumbering.Render(), content)
	content = r.headingShade(level)(content)

	// wrap if needed
	wrapped, _ := text.WrapWithPad(content, r.lineWidth, r.pad())
	_, _ = fmt.Fprintln(w, wrapped)

	// render the underline, if any
	if level == 1 {
		_, _ = fmt.Fprintf(w, "%s%s\n", r.pad(), strings.Repeat("─", r.lineWidth-r.leftPad))
	}

	_, _ = fmt.Fprintln(w)
}

func (r *renderer) renderCodeBlock(w io.Writer, node *ast.CodeBlock) {
	code := string(node.Literal)
	var lexer chroma.Lexer
	// try to get the lexer from the language tag if any
	if len(node.Info) > 0 {
		lexer = lexers.Get(string(node.Info))
	}
	// fallback on detection
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}
	// all failed :-(
	if lexer == nil {
		lexer = lexers.Fallback
	}
	// simplify the lexer output
	lexer = chroma.Coalesce(lexer)

	var formatter chroma.Formatter
	if color.NoColor {
		formatter = formatters.Fallback
	} else {
		formatter = formatters.TTY8
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		// Something failed, falling back to no highlight render
		r.renderFormattedCodeBlock(w, code)
		return
	}

	buf := &bytes.Buffer{}

	err = formatter.Format(buf, styles.Pygments, iterator)
	if err != nil {
		// Something failed, falling back to no highlight render
		r.renderFormattedCodeBlock(w, code)
		return
	}

	r.renderFormattedCodeBlock(w, buf.String())
}

func (r *renderer) renderFormattedCodeBlock(w io.Writer, code string) {
	// remove the trailing line break
	code = strings.TrimRight(code, "\n")

	r.addPad(GreenBold("┃ "))
	output, _ := text.WrapWithPad(code, r.lineWidth, r.pad())
	r.popPad()

	_, _ = fmt.Fprint(w, output)

	_, _ = fmt.Fprintf(w, "\n\n")
}

func (r *renderer) renderHTMLBlock(w io.Writer, node *ast.HTMLBlock) {
	var buf bytes.Buffer

	flushInline := func() {
		if r.inlineAccumulator.Len() <= 0 {
			return
		}
		content := r.inlineAccumulator.String()
		r.inlineAccumulator.Reset()
		out, _ := text.WrapWithPad(content, r.lineWidth, r.pad())
		_, _ = fmt.Fprint(&buf, out, "\n\n")
	}

	doc, err := html.Parse(bytes.NewReader(node.Literal))
	if err != nil {
		// if there is a parsing error, fallback to a simple render
		r.inlineAccumulator.Reset()
		content := Red(string(node.Literal))
		out, _ := text.WrapWithPad(content, r.lineWidth, r.pad())
		_, _ = fmt.Fprint(w, out, "\n\n")
		return
	}

	htmlWalker.WalkFunc(doc, func(node *html.Node, entering bool) htmlWalker.WalkStatus {
		// if node.Type != html.TextNode {
		// 	fmt.Println(node.Type, "(", node.Data, ")", entering)
		// }

		switch node.Type {
		case html.CommentNode, html.DoctypeNode:
			// Not rendered

		case html.DocumentNode:

		case html.ElementNode:
			switch node.Data {
			case "html", "body":
				return htmlWalker.GoToNext

			case "head":
				return htmlWalker.SkipChildren

			case "div", "p":
				if entering {
					flushInline()
				} else {
					content := r.inlineAccumulator.String()
					r.inlineAccumulator.Reset()
					if len(content) == 0 {
						return htmlWalker.GoToNext
					}
					// remove all line breaks, those are fully managed in HTML
					content = strings.Replace(content, "\n", "", -1)
					align := getDivHTMLAttr(node.Attr)
					content, _ = text.WrapWithPadAlign(content, r.lineWidth, r.pad(), align)
					_, _ = fmt.Fprint(&buf, content, "\n\n")
				}

			case "h1":
				if !entering {
					r.renderHeading(&buf, 1)
				}
			case "h2":
				if !entering {
					r.renderHeading(&buf, 2)
				}
			case "h3":
				if !entering {
					r.renderHeading(&buf, 3)
				}
			case "h4":
				if !entering {
					r.renderHeading(&buf, 4)
				}
			case "h5":
				if !entering {
					r.renderHeading(&buf, 5)
				}
			case "h6":
				if !entering {
					r.renderHeading(&buf, 6)
				}

			case "img":
				flushInline()
				src, title := getImgHTMLAttr(node.Attr)
				str, _ := r.renderImage(src, title, r.lineWidth-len(r.pad()))
				r.inlineAccumulator.WriteString(str)

			case "hr":
				flushInline()
				r.renderHorizontalRule(&buf)

			case "ul", "ol":
				if !entering {
					if node.NextSibling == nil {
						_, _ = fmt.Fprint(&buf, "\n")
						return htmlWalker.GoToNext
					}
					switch node.NextSibling.Data {
					case "ul", "ol":
					default:
						_, _ = fmt.Fprint(&buf, "\n")
					}
				}

			case "li":
				if entering {
					switch node.Parent.Data {
					case "ul":
						r.indent = r.pad() + Green("• ")
						r.addPad("  ")

					case "ol":
						itemNumber := 1
						previous := node.PrevSibling
						for previous != nil {
							itemNumber++
							previous = previous.PrevSibling
						}
						prefix := fmt.Sprintf("%d. ", itemNumber)
						r.indent = r.pad() + Green(prefix)
						r.addPad(strings.Repeat(" ", text.Len(prefix)))

					default:
						r.inlineAccumulator.WriteString(Red(renderRawHtml(node)))
						return htmlWalker.SkipChildren
					}
				} else {
					switch node.Parent.Data {
					case "ul", "ol":
						content := r.inlineAccumulator.String()
						r.inlineAccumulator.Reset()
						out, _ := text.WrapWithPadIndent(content, r.lineWidth, r.indent, r.pad())
						r.indent = ""
						_, _ = fmt.Fprint(&buf, out, "\n")
						r.popPad()
					}
				}

			case "a":
				if entering {
					r.inlineAccumulator.WriteString("[")
				} else {
					href, alt := getAHTMLAttr(node.Attr)
					r.inlineAccumulator.WriteString("](")
					r.inlineAccumulator.WriteString(Blue(href))
					if len(alt) > 0 {
						r.inlineAccumulator.WriteString(" ")
						r.inlineAccumulator.WriteString(alt)
					}
					r.inlineAccumulator.WriteString(")")
				}

			case "br":
				if entering {
					r.inlineAccumulator.WriteString("\n")
				}

			case "table":
				if entering {
					flushInline()
					r.table = newTableRenderer()
				} else {
					r.table.Render(&buf, r.leftPad, r.lineWidth)
					r.table = nil
				}

			case "thead", "tbody":
				// nothing to do

			case "tr":
				if entering && node.Parent.Data != "thead" {
					r.table.NextBodyRow()
				}

			case "th":
				if !entering {
					content := r.inlineAccumulator.String()
					r.inlineAccumulator.Reset()

					align := getTdHTMLAttr(node.Attr)
					r.table.AddHeaderCell(content, align)
				}

			case "td":
				if !entering {
					content := r.inlineAccumulator.String()
					r.inlineAccumulator.Reset()

					align := getTdHTMLAttr(node.Attr)
					r.table.AddBodyCell(content, align)
				}

			case "strong", "b":
				if entering {
					r.inlineAccumulator.WriteString(boldOn)
				} else {
					// This is super silly but some terminals, instead of having
					// the ANSI code SGR 21 do "bold off" like the logic would guide,
					// do "double underline" instead. This is madness.

					// To resolve that problem, we take a snapshot of the escape state,
					// remove the bold, then output "reset all" + snapshot
					es := text.EscapeState{}
					es.Witness(r.inlineAccumulator.String())
					es.Bold = false
					r.inlineAccumulator.WriteString(resetAll)
					r.inlineAccumulator.WriteString(es.FormatString())
				}

			case "i", "em":
				if entering {
					r.inlineAccumulator.WriteString(italicOn)
				} else {
					r.inlineAccumulator.WriteString(italicOff)
				}

			case "s":
				if entering {
					r.inlineAccumulator.WriteString(crossedOutOn)
				} else {
					r.inlineAccumulator.WriteString(crossedOutOff)
				}

			default:
				r.inlineAccumulator.WriteString(Red(renderRawHtml(node)))
			}

		case html.TextNode:
			t := strings.TrimSpace(node.Data)
			t = strings.ReplaceAll(t, "\n", "")
			r.inlineAccumulator.WriteString(t)

		default:
			panic("unhandled case")
		}

		return htmlWalker.GoToNext
	})

	flushInline()
	_, _ = fmt.Fprint(w, buf.String())
	r.inlineAccumulator.Reset()

	// 		// dl + (dt+dd)
	//
	// 		// details
	// 		// summary
	//
}

func getDivHTMLAttr(attrs []html.Attribute) text.Alignment {
	for _, attr := range attrs {
		switch attr.Key {
		case "align":
			switch attr.Val {
			case "left":
				return text.AlignLeft
			case "center":
				return text.AlignCenter
			case "right":
				return text.AlignRight
			}
		}
	}
	return text.AlignLeft
}

func getImgHTMLAttr(attrs []html.Attribute) (src, title string) {
	for _, attr := range attrs {
		switch attr.Key {
		case "src":
			src = attr.Val
		case "alt":
			title = attr.Val
		}
	}
	return
}

func getAHTMLAttr(attrs []html.Attribute) (href, alt string) {
	for _, attr := range attrs {
		switch attr.Key {
		case "href":
			href = attr.Val
		case "alt":
			alt = attr.Val
		}
	}
	return
}

func getTdHTMLAttr(attrs []html.Attribute) CellAlign {
	for _, attr := range attrs {
		switch attr.Key {
		case "align":
			switch attr.Val {
			case "right":
				return CellAlignRight
			case "left":
				return CellAlignLeft
			case "center":
				return CellAlignCenter
			}

		case "style":
			for _, pair := range strings.Split(attr.Val, " ") {
				split := strings.Split(pair, ":")
				if split[0] != "text-align" || len(split) != 2 {
					continue
				}
				switch split[1] {
				case "right":
					return CellAlignRight
				case "left":
					return CellAlignLeft
				case "center":
					return CellAlignCenter
				}
			}
		}
	}
	return CellAlignLeft
}

func renderRawHtml(node *html.Node) string {
	var result strings.Builder
	openContent := make([]string, 0, 8)

	openContent = append(openContent, node.Data)
	for _, attr := range node.Attr {
		openContent = append(openContent, fmt.Sprintf("%s=\"%s\"", attr.Key, attr.Val))
	}

	result.WriteString("<")
	result.WriteString(strings.Join(openContent, " "))

	if node.FirstChild == nil {
		result.WriteString("/>")
		return result.String()
	}

	result.WriteString(">")

	child := node.FirstChild
	for child != nil {
		if child.Type == html.TextNode {
			t := strings.TrimSpace(child.Data)
			result.WriteString(t)
			child = child.NextSibling
			continue
		}

		switch node.Data {
		case "ul", "p":
			result.WriteString("\n  ")
		}

		result.WriteString(renderRawHtml(child))
		child = child.NextSibling
	}

	switch node.Data {
	case "ul", "p":
		result.WriteString("\n")
	}

	result.WriteString("</")
	result.WriteString(node.Data)
	result.WriteString(">")

	return result.String()
}

func (r *renderer) renderImage(dest string, title string, lineWidth int) (result string, rendered bool) {
	title = strings.ReplaceAll(title, "\n", "")
	title = strings.TrimSpace(title)
	dest = strings.ReplaceAll(dest, "\n", "")
	dest = strings.TrimSpace(dest)

	fallback := func() (string, bool) {
		return fmt.Sprintf("![%s](%s)", title, Blue(dest)), false
	}

	reader, err := imageFromDestination(dest)
	if err != nil {
		return fallback()
	}

	x := lineWidth

	if r.imageDithering == ansimage.DitheringWithChars || r.imageDithering == ansimage.DitheringWithBlocks {
		// not sure why this is needed by ansimage
		// x *= 4
	}

	img, err := ansimage.NewScaledFromReader(reader, math.MaxInt32, x,
		stdcolor.Black, ansimage.ScaleModeFit, r.imageDithering)

	if err != nil {
		return fallback()
	}

	if title != "" {
		return fmt.Sprintf("%s%s: %s", img.Render(), title, Blue(dest)), true
	}
	return fmt.Sprintf("%s%s", img.Render(), Blue(dest)), true
}

func imageFromDestination(dest string) (io.ReadCloser, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	if strings.HasPrefix(dest, "http://") || strings.HasPrefix(dest, "https://") {
		res, err := client.Get(dest)
		if err != nil {
			return nil, err
		}
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("http: %v", http.StatusText(res.StatusCode))
		}

		return res.Body, nil
	}

	return os.Open(dest)
}

func removeLineBreak(text string) string {
	lines := strings.Split(text, "\n")

	if len(lines) <= 1 {
		return text
	}

	for i, l := range lines {
		switch i {
		case 0:
			lines[i] = strings.TrimRightFunc(l, unicode.IsSpace)
		case len(lines) - 1:
			lines[i] = strings.TrimLeftFunc(l, unicode.IsSpace)
		default:
			lines[i] = strings.TrimFunc(l, unicode.IsSpace)
		}
	}
	return strings.Join(lines, " ")
}

func shouldCleanText(node ast.Node) bool {
	for node != nil {
		switch node.(type) {
		case *ast.BlockQuote:
			return false

		case *ast.Heading, *ast.Image, *ast.Link,
			*ast.TableCell, *ast.Document, *ast.ListItem:
			return true
		}

		node = node.GetParent()
	}

	panic("bad markdown document or missing case")
}
