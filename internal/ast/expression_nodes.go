package ast

import (
	"strconv"

	"github.com/thetangentline/interpreter/internal/token"
)

type Identifier struct {
	Token token.Token
	Value string
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

type StringLiteral struct {
	Token token.Token
	Value string
}

type Boolean struct {
	Token token.Token
	Value bool
}

type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

type FunctionLiteral struct {
	Token      token.Token
	Parameters []*Identifier
	Body       *BlockStatement
}

type CallExpression struct {
	Token     token.Token
	Function  Expression
	Arguments []Expression
}

type ForExpression struct {
	Token     token.Token
	Init      Statement
	Condition Expression
	Update    Statement
	Body      *BlockStatement
}

func (i *Identifier) expressionNode()         {}
func (il *IntegerLiteral) expressionNode()    {}
func (sl *StringLiteral) expressionNode()     {}
func (b *Boolean) expressionNode()            {}
func (pfe *PrefixExpression) expressionNode() {}
func (ife *InfixExpression) expressionNode()  {}
func (ie *IfExpression) expressionNode()      {}
func (fl *FunctionLiteral) expressionNode()   {}
func (ce *CallExpression) expressionNode()    {}
func (fe *ForExpression) expressionNode()     {}

func (i *Identifier) TokenLiteral() string         { return i.Token.Literal }
func (il *IntegerLiteral) TokenLiteral() string    { return il.Token.Literal }
func (sl *StringLiteral) TokenLiteral() string     { return sl.Token.Literal }
func (b *Boolean) TokenLiteral() string            { return b.Token.Literal }
func (pfe *PrefixExpression) TokenLiteral() string { return pfe.Token.Literal }
func (ife *InfixExpression) TokenLiteral() string  { return ife.Token.Literal }
func (ie *IfExpression) TokenLiteral() string      { return ie.Token.Literal }
func (fl *FunctionLiteral) TokenLiteral() string   { return fl.Token.Literal }
func (ce *CallExpression) TokenLiteral() string    { return ce.Token.Literal }
func (fe *ForExpression) TokenLiteral() string     { return fe.Token.Literal }

func (i *Identifier) String() string      { return i.Value }
func (il *IntegerLiteral) String() string { return strconv.Itoa(int(il.Value)) }
func (sl *StringLiteral) String() string  { return sl.Value }
func (b *Boolean) String() string         { return strconv.FormatBool(b.Value) }
func (pfe *PrefixExpression) String() string {
	if pfe.Right == nil {
		return pfe.Operator
	}
	return pfe.Operator + pfe.Right.String()
}
func (ife *InfixExpression) String() string {
	left := "<nil>"
	right := "<nil>"
	if ife.Left != nil {
		left = ife.Left.String()
	}
	if ife.Right != nil {
		right = ife.Right.String()
	}
	return left + " " + ife.Operator + " " + right
}
func (ie *IfExpression) String() string {
	var elseBranch string
	if ie.Alternative != nil {
		elseBranch = "else {\n" + ie.Alternative.String() + "}"
	}
	return "if " + ie.Condition.String() + " {\n" + ie.Consequence.String() + "}" + elseBranch
}
func (fl *FunctionLiteral) String() string {
	result := "fn("
	for i, p := range fl.Parameters {
		result += p.String()
		if i != len(fl.Parameters)-1 {
			result += ","
		}
	}
	result += ") {\n" + fl.Body.String() + "}"
	return result
}
func (ce *CallExpression) String() string {
	result := ce.Function.String() + "("
	for i, p := range ce.Arguments {
		result += p.String()
		if i != len(ce.Arguments)-1 {
			result += ","
		}
	}
	result += ")"
	return result
}
func (fe *ForExpression) String() string {
	return "for(" + fe.Init.String() + ";" + fe.Condition.String() + ";" + fe.Update.String() + ") {\n" + fe.Body.String() + "\n}"
}
