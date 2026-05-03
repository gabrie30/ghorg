package html

import . "golang.org/x/net/html"

// WalkStatus allows NodeVisitor to have some control over the tree traversal.
// It is returned from NodeVisitor and different values allow Node.Walk to
// decide which node to go to next.
type WalkStatus int

const (
	// GoToNext is the default traversal of every node.
	GoToNext WalkStatus = iota
	// SkipChildren tells walker to skip all children of current node.
	SkipChildren
	// Terminate tells walker to terminate the traversal.
	Terminate
)

// NodeVisitor is a callback to be called when traversing the syntax tree.
// Called twice for every node: once with entering=true when the branch is
// first visited, then with entering=false after all the children are done.
type NodeVisitor interface {
	Visit(node *Node, entering bool) WalkStatus
}

func Walk(n *Node, visitor NodeVisitor) WalkStatus {
	isContainer := n.FirstChild != nil

	// some special case that are container but can be self closing
	if n.Type == ElementNode {
		switch n.Data {
		case "td":
			isContainer = true
		}
	}

	status := visitor.Visit(n, true)

	if status == Terminate {
		// even if terminating, close container node
		if isContainer {
			visitor.Visit(n, false)
		}
	}

	if isContainer && status != SkipChildren {
		child := n.FirstChild
		for child != nil {
			status = Walk(child, visitor)
			if status == Terminate {
				return status
			}
			child = child.NextSibling
		}
	}

	if isContainer {
		status = visitor.Visit(n, false)
		if status == Terminate {
			return status
		}
	}

	return GoToNext
}

// NodeVisitorFunc casts a function to match NodeVisitor interface
type NodeVisitorFunc func(node *Node, entering bool) WalkStatus

// Visit calls visitor function
func (f NodeVisitorFunc) Visit(node *Node, entering bool) WalkStatus {
	return f(node, entering)
}

// WalkFunc is like Walk but accepts just a callback function
func WalkFunc(n *Node, f NodeVisitorFunc) {
	Walk(n, f)
}
