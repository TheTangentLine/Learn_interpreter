package parser

import (
	"strconv"

	"github.com/thetangentline/interpreter/internal/ast"
	"github.com/thetangentline/interpreter/internal/token"
)

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	value, err := strconv.Atoi(p.curToken.Literal)
	if err != nil {
		return nil
	}
	return &ast.IntegerLiteral{Token: p.curToken, Value: int64(value)}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curToken.Type == token.TRUE}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expr.Right = p.parseExpression(PREFIX)
	return expr
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek([]token.TokenType{token.RPAREN}) {
		return nil
	}
	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expr := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek([]token.TokenType{token.LPAREN}) {
		return nil
	}
	p.nextToken()
	expr.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek([]token.TokenType{token.RPAREN}) {
		return nil
	}
	if !p.expectPeek([]token.TokenType{token.LBRACE}) {
		return nil
	}
	expr.Consequence = p.parseBlockStatement()

	if p.peekToken.Type == token.ELSE {
		p.nextToken()
		if !p.expectPeek([]token.TokenType{token.LBRACE}) {
			return nil
		}
		expr.Alternative = p.parseBlockStatement()
	}

	return expr
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek([]token.TokenType{token.LPAREN}) {
		return nil
	}
	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek([]token.TokenType{token.LBRACE}) {
		return nil
	}
	lit.Body = p.parseBlockStatement()

	return lit
}
