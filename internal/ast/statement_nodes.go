package ast

import "github.com/thetangentline/interpreter/internal/token"

type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (ls *LetStatement) statementNode()        {}
func (rs *ReturnStatement) statementNode()     {}
func (es *ExpressionStatement) statementNode() {}
func (bs *BlockStatement) statementNode()      {}

func (ls *LetStatement) TokenLiteral() string        { return ls.Token.Literal }
func (rs *ReturnStatement) TokenLiteral() string     { return rs.Token.Literal }
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (bs *BlockStatement) TokenLiteral() string      { return bs.Token.Literal }

func (ls *LetStatement) String() string {
	val := "<nil>"
	if ls.Value != nil {
		val = ls.Value.String()
	}
	return "let " + ls.Name.String() + " = " + val
}
func (rs *ReturnStatement) String() string {
	val := "<nil>"
	if rs.ReturnValue != nil {
		val = rs.ReturnValue.String()
	}
	return "return " + val
}
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}
func (bs *BlockStatement) String() string {
	var block string
	for _, s := range bs.Statements {
		block += s.String() + ";\n"
	}
	return block
}
