package v1

import (
	"bytes"
	"fmt"
	"strconv"
)

var textFormat = "%s" // Changed to "%q" in tests for better error messages.

type NodeType int

func (t NodeType) Type() NodeType {
	return t
}

type Node interface {
	Type() NodeType
	String() string
	// Copy does a deep copy of the Node and all its components.
	Copy() Node
	Position() Pos // byte position of start of node in full original input string
}

const (
	NodeList NodeType = iota
	NodeText
	NodeVar
	NodeLookup
	NodeUnary
	NodeFloat
	NodeInteger
	NodeString
	NodeBool
	NodeAdd
	NodeMul
	NodeMapExpr
	NodeMapElem
	NodeIndexExpr
	NodeSet
	NodeIf
	NodeElseIf
	NodeFor
)

// This is a stack of nodes starting at a position.  It has the default NodeType
// but should never end up in the AST;  it's use is in implementing order of
// operations for expressions
type nodeStack struct {
	NodeType
	Pos
	Nodes []Node
}

func newStack(pos Pos) *nodeStack {
	return &nodeStack{Pos: pos}
}

func (n *nodeStack) len() int       { return len(n.Nodes) }
func (n *nodeStack) push(node Node) { n.Nodes = append(n.Nodes, node) }
func (n *nodeStack) pop() Node {
	var r Node
	if len(n.Nodes) > 0 {
		r = n.Nodes[len(n.Nodes)-1]
		n.Nodes = n.Nodes[:len(n.Nodes)-1]
	}
	return r
}

// ListNode holds a sequence of nodes.
type ListNode struct {
	NodeType
	Pos
	Nodes []Node // The element nodes in lexical order.
}

func newList(pos Pos) *ListNode {
	return &ListNode{NodeType: NodeList, Pos: pos}
}

func (l *ListNode) append(n Node) { l.Nodes = append(l.Nodes, n) }
func (l *ListNode) len() int      { return len(l.Nodes) }

func (l *ListNode) String() string {
	b := new(bytes.Buffer)
	for _, n := range l.Nodes {
		fmt.Fprint(b, n)
	}
	return b.String()
}

func (l *ListNode) CopyList() *ListNode {
	if l == nil {
		return l
	}
	n := newList(l.Pos)
	for _, elem := range l.Nodes {
		n.append(elem.Copy())
	}
	return n
}

func (l *ListNode) Copy() Node { return l.CopyList() }

// TextNode holds plain text.
type TextNode struct {
	NodeType
	Pos
	Text []byte // The text; may span newlines.
}

func newText(pos Pos, text string) *TextNode {
	return &TextNode{NodeType: NodeText, Pos: pos, Text: []byte(text)}
}

func (t *TextNode) String() string { return fmt.Sprintf(textFormat, t.Text) }
func (t *TextNode) Copy() Node     { return &TextNode{NodeText, t.Pos, append([]byte{}, t.Text...)} }

// VarNode represents a var print expr, ie {{ ... }}.
// It is represented as a sequence of expressions.
type VarNode struct {
	NodeType
	Pos
	Node Node
}

func newVar(pos Pos) *VarNode {
	return &VarNode{NodeType: NodeVar, Pos: pos}
}

func (v *VarNode) String() string { return "{{ " + v.Node.String() + " }}" }
func (v *VarNode) Copy() Node     { return &VarNode{v.NodeType, v.Pos, v.Node} }

// A LookupNode is a variable lookup.
type LookupNode struct {
	NodeType
	Pos
	Name string
}

func newLookup(pos Pos, name string) *LookupNode {
	return &LookupNode{NodeType: NodeLookup, Pos: pos, Name: name}
}
func (l *LookupNode) String() string { return l.Name }
func (l *LookupNode) Copy() Node     { return newLookup(l.Pos, l.Name) }

type StringNode struct {
	NodeType
	Pos
	Value string
}

func (s *StringNode) Copy() Node     { return &StringNode{s.NodeType, s.Pos, s.Value} }
func (s *StringNode) String() string { return fmt.Sprintf(`"%s"`, s.Value) }

type BoolNode struct {
	NodeType
	Pos
	Value bool
}

func (s *BoolNode) Copy() Node     { return &BoolNode{s.NodeType, s.Pos, s.Value} }
func (s *BoolNode) String() string { return fmt.Sprintf("%v", s.Value) }

type IntegerNode struct {
	NodeType
	Pos
	Value int64
}

func (i *IntegerNode) Copy() Node     { return &IntegerNode{i.NodeType, i.Pos, i.Value} }
func (i *IntegerNode) String() string { return strconv.FormatInt(i.Value, 10) }

type FloatNode struct {
	NodeType
	Pos
	Value float64
}

func (f *FloatNode) Copy() Node     { return &FloatNode{f.NodeType, f.Pos, f.Value} }
func (f *FloatNode) String() string { return fmt.Sprint(f.Value) }

type UnaryNode struct {
	NodeType
	Pos
	Value Node
	Unary item
}

func newUnaryNode(val Node, unary item) *UnaryNode {
	return &UnaryNode{NodeUnary, val.Position(), val, unary}
}

func (u *UnaryNode) Copy() Node     { return &UnaryNode{u.NodeType, u.Pos, u.Value, u.Unary} }
func (u *UnaryNode) String() string { return fmt.Sprintf("%s%s", u.Unary.val, u.Value) }

// newLiteral creates a new string, integer, or float node depending on itemType
func newLiteral(pos Pos, typ itemType, val string) Node {
	switch typ {
	case tokenFloat:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			panic(err)
		}
		return &FloatNode{NodeFloat, pos, v}
	case tokenInteger:
		// FIXME: complex integer types?  hex, octal, etc?
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			panic(err)
		}
		return &IntegerNode{NodeInteger, pos, v}
	case tokenString:
		return &StringNode{NodeString, pos, val}
	case tokenBool:
		var v bool
		if val == "true" {
			v = true
		}
		return &BoolNode{NodeBool, pos, v}
	}
	panic(fmt.Sprint("unexpected literal type ", typ))
}

type AddExpr struct {
	NodeType
	Pos
	lhs      Node
	rhs      Node
	operator item
}

func newAddExpr(lhs, rhs Node, operator item) *AddExpr {
	return &AddExpr{NodeAdd, lhs.Position(), lhs, rhs, operator}
}

func (a *AddExpr) String() string {
	return fmt.Sprintf("%s %s %s", a.lhs, a.operator.val, a.rhs)
}

func (a *AddExpr) Copy() Node {
	return newAddExpr(a.lhs, a.rhs, a.operator)
}

type MulExpr struct {
	NodeType
	Pos
	lhs      Node
	rhs      Node
	operator item
}

func newMulExpr(lhs, rhs Node, operator item) *MulExpr {
	return &MulExpr{NodeMul, lhs.Position(), lhs, rhs, operator}
}

func (m *MulExpr) String() string {
	return fmt.Sprintf("%s %s %s", m.lhs, m.operator.val, m.rhs)
}

func (m *MulExpr) Copy() Node {
	return newMulExpr(m.lhs, m.rhs, m.operator)
}

// complex literals

type MapExpr struct {
	NodeType
	Pos
	Elems []*MapElem
}

func newMapExpr(pos Pos) *MapExpr {
	return &MapExpr{NodeType: NodeMapExpr, Pos: pos}
}

func (m *MapExpr) len() int          { return len(m.Elems) }
func (m *MapExpr) append(n *MapElem) { m.Elems = append(m.Elems, n) }

func (m *MapExpr) String() string {
	b := new(bytes.Buffer)
	b.WriteString("{")
	for i, n := range m.Elems {
		fmt.Fprint(b, n)
		if i != len(m.Elems)-1 {
			b.WriteString(", ")
		}
	}
	b.WriteString("}")
	return b.String()
}

func (m *MapExpr) Copy() Node {
	if m == nil {
		return m
	}
	n := newMapExpr(m.Pos)
	for _, elem := range m.Elems {
		n.append(elem.Copy().(*MapElem))
	}
	return n
}

type MapElem struct {
	NodeType
	Pos
	Key   Node
	Value Node
}

func newMapElem(lhs, rhs Node) *MapElem {
	return &MapElem{NodeMapElem, lhs.Position(), lhs, rhs}
}

func (m *MapElem) String() string {
	return fmt.Sprintf("%s: %s", m.Key, m.Value)
}

func (m *MapElem) Copy() Node {
	return newMapElem(m.Key, m.Value)
}

type IndexExpr struct {
	NodeType
	Pos
	Value Node
	Index Node
}

func newIndexExpr(val, idx Node) *IndexExpr {
	return &IndexExpr{NodeIndexExpr, val.Position(), val, idx}
}

func (i *IndexExpr) String() string {
	return fmt.Sprintf("%s[%s]", i.Value, i.Index)
}

func (i *IndexExpr) Copy() Node {
	return newIndexExpr(i.Value, i.Index)
}

// block types
type SetNode struct {
	NodeType
	Pos
	lhs Node
	rhs Node
}

func newSet(pos Pos, lhs, rhs Node) *SetNode {
	return &SetNode{NodeSet, pos, lhs, rhs}
}

// FIXME: environment needed to really recreate this as it requires block
// begin and end tags, which we don't technically know
func (s *SetNode) String() string { return fmt.Sprintf("{%% set %s = %s %%}", s.lhs, s.rhs) }
func (s *SetNode) Copy() Node {
	return newSet(s.Pos, s.lhs.Copy(), s.rhs.Copy())
}

// A ConditionalNode is a node that has a guard and a body.  If the guard evals
// as True, then the body is rendered.  Otherwise, it's a Noop.  If's and ElseIf's
// are modeled this way.
type ConditionalNode struct {
	NodeType
	Pos
	Guard Node
	Body  Node
}

func newIfCond(pos Pos) *ConditionalNode {
	return newConditional(pos, NodeIf)
}

func newElifCond(pos Pos) *ConditionalNode {
	return newConditional(pos, NodeElseIf)
}

func newConditional(pos Pos, typ NodeType) *ConditionalNode {
	return &ConditionalNode{NodeType: typ, Pos: pos}
}

func (c *ConditionalNode) Copy() Node {
	n := newConditional(c.Pos, c.NodeType)
	n.Guard = c.Guard.Copy()
	n.Body = c.Body.Copy()
	return n
}

func (c *ConditionalNode) String() string {
	b := new(bytes.Buffer)
	switch c.NodeType {
	case NodeIf:
		fmt.Fprintf(b, "{%% if %s %%}", c.Guard)
	case NodeElseIf:
		fmt.Fprintf(b, "{%% elif %s %%}", c.Guard)
	}
	fmt.Fprint(b, c.Body)
	return b.String()
}

// IfBlockNode represents a full {% if %}... block.  The if and elif bodies are
// modeled using the ConditionalNode, and are evaluated in order to determine
// which body to render.  `Else` will be a ListNode, but can be nil if no such
// clause is present.
type IfBlockNode struct {
	NodeType
	Pos
	Conditionals []Node
	Else         Node
}

func newIf(pos Pos) *IfBlockNode {
	return &IfBlockNode{NodeType: NodeIf, Pos: pos, Conditionals: make([]Node, 0, 1)}
}

func (i *IfBlockNode) String() string {
	b := new(bytes.Buffer)
	for _, c := range i.Conditionals {
		fmt.Fprint(b, c)
	}
	if i.Else != nil {
		fmt.Fprintf(b, "{%% else %%}%s", i.Else)
	}
	fmt.Fprint(b, "{% endif %}")
	return b.String()
}
func (i *IfBlockNode) Copy() Node {
	n := newIf(i.Pos)
	n.Conditionals = make([]Node, len(i.Conditionals))
	for _, e := range i.Conditionals {
		n.Conditionals = append(n.Conditionals, e.Copy())
	}
	if i.Else != nil {
		n.Else = i.Else.Copy()
	}
	return n
}

type ForNode struct {
	NodeType
	Pos
	ForExpr Node
	InExpr  Node
	Body    Node
}

func newFor(pos Pos) *ForNode {
	return &ForNode{NodeType: NodeFor, Pos: pos}
}

// FIXME: This should use the environment's begin and end tags, which we
// don't have down at this level...
func (f *ForNode) String() string {
	return fmt.Sprintf("{%% for %s in %s %%}%s{%% endfor %%}", f.ForExpr, f.InExpr, f.Body)
}
func (f *ForNode) Copy() Node {
	n := newFor(f.Pos)
	n.ForExpr = f.ForExpr.Copy()
	n.InExpr = f.InExpr.Copy()
	n.Body = f.Body.Copy()
	return n
}

type BlockNode struct {
	NodeType
	Pos
	Name string
	Body Node
}

func (b *BlockNode) String() string {
	return fmt.Sprintf("{%% block %v %%}%v{%% endblock %%}", b.Name, b.Body)
}

func (b *BlockNode) Copy() Node {
	return &BlockNode{b.NodeType, b.Pos, b.Name, b.Body.Copy()}
}

type Import struct {
	Name string
	As   string
}

type ExtendsNode struct {
	NodeType
	Pos
}
type PrintNode struct {
	NodeType
	Pos
}
type MacroNode struct {
	NodeType
	Pos
}
type IncludeNode struct {
	NodeType
	Pos
}

type FromNode struct {
	NodeType
	Pos
	Module  string
	Imports []Import
}
type ImportNode struct {
	NodeType
	Pos
	Module string
	Body   Import
}
type CallNode struct {
	NodeType
	Pos
}
