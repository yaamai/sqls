package parseutil

import (
	"github.com/yaamai/sqls/ast"
	"github.com/yaamai/sqls/ast/astutil"
	"github.com/yaamai/sqls/token"
)

func ExtractIdenfiers(parsed ast.TokenList, pos token.Pos) ([]ast.Node, error) {
	stmt, err := extractFocusedStatement(parsed, pos)
	if err != nil {
		return nil, err
	}

	identifierMatcher := astutil.NodeMatcher{
		NodeTypes: []ast.NodeType{
			ast.TypeIdentifier,
		},
	}
	return parsePrefix(astutil.NewNodeReader(stmt), identifierMatcher, parseIdentifier), nil
}

func parseIdentifier(reader *astutil.NodeReader) []ast.Node {
	return []ast.Node{reader.CurNode}
}
