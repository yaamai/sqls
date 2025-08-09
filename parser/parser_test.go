package parser

import (
	"reflect"
	"testing"

	"github.com/yaamai/sqls/ast"
	"github.com/yaamai/sqls/token"
)

func TestParseStatement(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		checkFn func(t *testing.T, stmts []*ast.Statement, input string)
	}{
		{
			name:  "single statement",
			input: "select 1;",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				if len(stmts) != 1 {
					t.Fatalf("Query does not contain 1 statements, got %d", len(stmts))
				}
				testStatement(t, stmts[0], 4, "select 1;")
			},
		},
		{
			name:  "single statement non semicolon",
			input: "select 1",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				if len(stmts) != 1 {
					t.Fatalf("Query does not contain 1 statements, got %d", len(stmts))
				}
				testStatement(t, stmts[0], 3, "select 1")
			},
		},
		{
			name:  "three statement",
			input: "select 1;select 2;select 3;",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				if len(stmts) != 3 {
					t.Fatalf("Query does not contain 3 statements, got %d", len(stmts))
				}
				testStatement(t, stmts[0], 4, "select 1;")
				testPos(t, stmts[0], genPosOneline(0), genPosOneline(9))
				testStatement(t, stmts[1], 4, "select 2;")
				testPos(t, stmts[1], genPosOneline(9), genPosOneline(18))
				testStatement(t, stmts[2], 4, "select 3;")
				testPos(t, stmts[2], genPosOneline(18), genPosOneline(27))
			},
		},
		{
			name:  "three statement non semicolon",
			input: "select 1;select 2;select 3",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				if len(stmts) != 3 {
					t.Fatalf("Query does not contain 3 statements, got %d", len(stmts))
				}
				testStatement(t, stmts[0], 4, "select 1;")
				testPos(t, stmts[0], genPosOneline(0), genPosOneline(9))
				testStatement(t, stmts[1], 4, "select 2;")
				testPos(t, stmts[1], genPosOneline(9), genPosOneline(18))
				testStatement(t, stmts[2], 3, "select 3")
				testPos(t, stmts[2], genPosOneline(18), genPosOneline(26))
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			stmts := parseInit(t, tt.input)
			tt.checkFn(t, stmts, tt.input)
		})
	}
}

func TestParseComments(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		checkFn func(t *testing.T, stmts []*ast.Statement, input string)
	}{
		{
			name:  "line comment with identiger",
			input: "-- foo\nbar",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 3, "-- foo\nbar")

				list := stmts[0].GetTokens()
				testItem(t, list[0], "-- foo")
				testIdentifier(t, list[2], "bar")
			},
		},
		{
			name:  "range comment with identiger",
			input: "/* foo */bar",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 2, "/* foo */bar")

				list := stmts[0].GetTokens()
				testIdentifier(t, list[1], "bar")
			},
		},
		{
			name:  "range comment with identiger list",
			input: "foo, /* foo */bar",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, "foo, /* foo */bar")

				list := stmts[0].GetTokens()
				testIdentifierList(t, list[0], "foo, /* foo */bar")
			},
		},
		{
			name:  "multi line range comment with identiger",
			input: "/*\n * foo\n */\nbar",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 3, "/*\n   foo\n */\nbar")

				list := stmts[0].GetTokens()
				testItem(t, list[0], "/*\n   foo\n */")
				testItem(t, list[1], "\n")
				testIdentifier(t, list[2], "bar")
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			stmts := parseInit(t, tt.input)
			tt.checkFn(t, stmts, tt.input)
		})
	}
}

func TestParseParenthesis(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		checkFn func(t *testing.T, stmts []*ast.Statement, input string)
	}{
		{
			name:  "single",
			input: "(3)",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testParenthesis(t, list[0], input)
				testPos(t, stmts[0], genPosOneline(0), genPosOneline(3))
			},
		},
		{
			name:  "with operator",
			input: "(3 - 4)",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testParenthesis(t, list[0], input)
				testPos(t, stmts[0], genPosOneline(0), genPosOneline(7))
			},
		},
		{
			name:  "inner parenthesis",
			input: "(1 * 2 + (3 - 4))",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testParenthesis(t, list[0], input)
				testPos(t, stmts[0], genPosOneline(0), genPosOneline(17))
			},
		},
		{
			name:  "with select",
			input: "select (select (x3) x2) and (y2) bar",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 7, input)

				list := stmts[0].GetTokens()
				testItem(t, list[0], "select")
				testItem(t, list[1], " ")
				testParenthesis(t, list[2], "(select (x3) x2)")
				testItem(t, list[3], " ")
				testItem(t, list[4], "and")
				testItem(t, list[5], " ")
				testAliased(t, list[6], "(y2) bar", "(y2)", "bar")

				parenthesis := testTokenList(t, list[2], 5).GetTokens()
				testItem(t, parenthesis[0], "(")
				testItem(t, parenthesis[1], "select")
				testItem(t, parenthesis[2], " ")
				testAliased(t, parenthesis[3], "(x3) x2", "(x3)", "x2")
				testItem(t, parenthesis[4], ")")
			},
		},
		{
			name:  "not close parenthesis",
			input: "select (select (x3) x2 and (y2) bar ",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 3, input)

				list := stmts[0].GetTokens()
				testItem(t, list[0], "select")
				testItem(t, list[1], " ")
				testParenthesis(t, list[2], "(select (x3) x2 and (y2) bar ")
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			stmts := parseInit(t, tt.input)
			tt.checkFn(t, stmts, tt.input)
		})
	}
}

func TestParseFunction(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		checkFn func(t *testing.T, stmts []*ast.Statement, input string)
	}{
		{
			name:  "function none args",
			input: "foo()",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				list := stmts[0].GetTokens()
				testFunction(t, list[0], "foo()")
			},
		},
		{
			name:  "function one args",
			input: "foo(a)",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				list := stmts[0].GetTokens()
				testFunction(t, list[0], "foo(a)")
			},
		},
		{
			name:  "function multiplue args",
			input: "foo(a, b, c)",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				list := stmts[0].GetTokens()
				testFunction(t, list[0], "foo(a, b, c)")
			},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			stmts := parseInit(t, tt.input)
			tt.checkFn(t, stmts, tt.input)
		})
	}

}

func TestParsePeriod_Double(t *testing.T) {
	input := `a.*, b.id`
	stmts := parseInit(t, input)

	testStatement(t, stmts[0], 1, input)

	list := stmts[0].GetTokens()
	testIdentifierList(t, list[0], input)

	il := testTokenList(t, list[0], 4).GetTokens()
	testMemberIdentifier(t, il[0], "a.*", "a", "*")
	testItem(t, il[1], ",")
	testItem(t, il[2], " ")
	testMemberIdentifier(t, il[3], "b.id", "b", "id")
}

func TestParsePeriod_WithWildcard(t *testing.T) {
	input := `a.*`
	stmts := parseInit(t, input)

	testStatement(t, stmts[0], 1, input)

	list := stmts[0].GetTokens()
	testMemberIdentifier(t, list[0], "a.*", "a", "*")
}

func TestParsePeriod_Invalid(t *testing.T) {
	input := `a.`
	stmts := parseInit(t, input)

	testStatement(t, stmts[0], 1, input)

	list := stmts[0].GetTokens()
	testMemberIdentifier(t, list[0], "a.", "a", "")
}

func TestParsePeriod_InvalidWithSelect(t *testing.T) {
	input := `SELECT foo. FROM foo`
	stmts := parseInit(t, input)

	testStatement(t, stmts[0], 7, input)

	list := stmts[0].GetTokens()
	testItem(t, list[0], "SELECT")
	testItem(t, list[1], " ")
	testMemberIdentifier(t, list[2], "foo.", "foo", "")
	testItem(t, list[3], " ")
	testItem(t, list[4], "FROM")
	testItem(t, list[5], " ")
	testIdentifier(t, list[6], "foo")
}

func TestParseIdentifier(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		checkFn func(t *testing.T, stmts []*ast.Statement, input string)
	}{
		{
			name:  "identifier",
			input: "abc",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifier(t, list[0], "abc")
			},
		},
		{
			name:  "double quote identifier",
			input: `"abc"`,
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifier(t, list[0], `"abc"`)
			},
		},
		{
			name:  "back quote identifier",
			input: "`abc`",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifier(t, list[0], input)
			},
		},
		{
			name:  "non close back quote identifier",
			input: "`abc",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifier(t, list[0], input)
			},
		},
		{
			name:  "wildcard",
			input: "*",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifier(t, list[0], "*")
			},
		},
		{
			name:  "select identifier",
			input: "select abc",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 3, input)
				list := stmts[0].GetTokens()
				testItem(t, list[0], "select")
				testItem(t, list[1], " ")
				testIdentifier(t, list[2], "abc")
			},
		},
		{
			name:  "from identifier",
			input: "select abc from def",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 7, input)
				list := stmts[0].GetTokens()
				testItem(t, list[0], "select")
				testItem(t, list[1], " ")
				testIdentifier(t, list[2], "abc")
				testItem(t, list[3], " ")
				testItem(t, list[4], "from")
				testItem(t, list[5], " ")
				testIdentifier(t, list[6], "def")
			},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			stmts := parseInit(t, tt.input)
			tt.checkFn(t, stmts, tt.input)
		})
	}
}

func TestMemberIdentifier(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		checkFn func(t *testing.T, stmts []*ast.Statement, input string)
	}{
		{
			name:  "member identifier",
			input: "a.b",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testMemberIdentifier(t, list[0], input, "a", "b")
			},
		},
		{
			name:  "double quote member identifier",
			input: `"abc"."def"`,
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testMemberIdentifier(t, list[0], input, `"abc"`, `"def"`)
			},
		},
		{
			name:  "back quote member identifier",
			input: "`abc`.`def`",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testMemberIdentifier(t, list[0], input, "`abc`", "`def`")
			},
		},
		{
			name:  "invalid member identifier",
			input: "a.",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testMemberIdentifier(t, list[0], input, "a", "")
			},
		},
		{
			name:  "member identifier wildcard",
			input: "a.*",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testMemberIdentifier(t, list[0], input, "a", "*")
			},
		},
		{
			name:  "member identifier select",
			input: "select foo.bar from abc",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 7, input)
				list := stmts[0].GetTokens()
				testItem(t, list[0], "select")
				testItem(t, list[1], " ")
				testMemberIdentifier(t, list[2], "foo.bar", "foo", "bar")
				testItem(t, list[3], " ")
				testItem(t, list[4], "from")
				testItem(t, list[5], " ")
				testIdentifier(t, list[6], "abc")
			},
		},
		{
			name:  "invalid member identifier select",
			input: "select foo. from abc",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 7, input)
				list := stmts[0].GetTokens()
				testItem(t, list[0], "select")
				testPos(t, list[0], genPosOneline(0), genPosOneline(6))
				testItem(t, list[1], " ")
				testPos(t, list[1], genPosOneline(6), genPosOneline(7))
				testMemberIdentifier(t, list[2], "foo.", "foo", "")
				testPos(t, list[2], genPosOneline(7), genPosOneline(11))
				testItem(t, list[3], " ")
				testPos(t, list[3], genPosOneline(11), genPosOneline(12))
				testItem(t, list[4], "from")
				testPos(t, list[4], genPosOneline(12), genPosOneline(16))
				testItem(t, list[5], " ")
				testIdentifier(t, list[6], "abc")
			},
		},
		{
			name:  "member identifier from",
			input: "select foo from myschema.abc",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 7, input)
				list := stmts[0].GetTokens()
				testItem(t, list[0], "select")
				testItem(t, list[1], " ")
				testIdentifier(t, list[2], "foo")
				testItem(t, list[3], " ")
				testItem(t, list[4], "from")
				testItem(t, list[5], " ")
				testMemberIdentifier(t, list[6], "myschema.abc", "myschema", "abc")
			},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			stmts := parseInit(t, tt.input)
			tt.checkFn(t, stmts, tt.input)
		})
	}
}

func TestParseMultiKeyword(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		checkFn func(t *testing.T, stmts []*ast.Statement, input string)
	}{
		{
			name:  "order keyword",
			input: "order by",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testMultiKeyword(t, list[0], input)
				testPos(t, stmts[0], genPosOneline(0), genPosOneline(8))
			},
		},
		{
			name:  "group keyword",
			input: "group by",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testMultiKeyword(t, list[0], input)
				testPos(t, stmts[0], genPosOneline(0), genPosOneline(8))
			},
		},
		{
			name:  "insert keyword",
			input: "insert into",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testMultiKeyword(t, list[0], input)
			},
		},
		{
			name:  "delete keyword",
			input: "delete from",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testMultiKeyword(t, list[0], input)
			},
		},
		{
			name:  "inner join keyword",
			input: "inner join",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testMultiKeyword(t, list[0], input)
			},
		},
		{
			name:  "left outer join keyword",
			input: "left outer join",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testMultiKeyword(t, list[0], input)
			},
		},
		{
			name:  "select with group keyword",
			input: "select a, b, c from abc group by d, e, f",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 11, input)
				list := stmts[0].GetTokens()
				testItem(t, list[0], "select")
				testItem(t, list[1], " ")
				testIdentifierList(t, list[2], "a, b, c")
				testItem(t, list[3], " ")
				testItem(t, list[4], "from")
				testItem(t, list[5], " ")
				testIdentifier(t, list[6], "abc")
				testItem(t, list[7], " ")
				testMultiKeyword(t, list[8], "group by")
				testItem(t, list[9], " ")
				testIdentifierList(t, list[10], "d, e, f")
			},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			stmts := parseInit(t, tt.input)
			tt.checkFn(t, stmts, tt.input)
		})
	}
}

func TestParseOperator(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		checkFn func(t *testing.T, stmts []*ast.Statement, input string)
	}{
		{
			name:  "plus operator",
			input: "foo+100",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testOperator(t, list[0], input, "foo", "+", "100")
			},
		},
		{
			name:  "minus operator",
			input: "foo-100",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testOperator(t, list[0], input, "foo", "-", "100")
			},
		},
		{
			name:  "mult operator",
			input: "foo*100",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testOperator(t, list[0], input, "foo", "*", "100")
			},
		},
		{
			name:  "div operator",
			input: "foo/100",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testOperator(t, list[0], input, "foo", "/", "100")
			},
		},
		{
			name:  "mod operator",
			input: "foo%100",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testOperator(t, list[0], input, "foo", "%", "100")
			},
		},
		{
			name:  "operator with whitespace",
			input: "foo + 100",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testOperator(t, list[0], input, "foo", "+", "100")

			},
		},
		{
			name:  "multiple operator",
			input: "1 + 2 - 3 / 4",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				operator1 := testOperator(t, list[0], "1 + 2 - 3 / 4", "1 + 2 - 3", "/", "4")
				operator2 := testOperator(t, operator1.GetLeft(), "1 + 2 - 3", "1 + 2", "-", "3")
				testOperator(t, operator2.GetLeft(), "1 + 2", "1", "+", "2")
			},
		},
		{
			name:  "in parenthesis",
			input: "(100+foo)",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				parenthesis := testParenthesis(t, list[0], input)
				testOperator(t, parenthesis.Inner().GetTokens()[0], "100+foo", "100", "+", "foo")
			},
		},
		{
			name:  "in double parenthesis",
			input: "((100+foo))",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				parenthesis1 := testParenthesis(t, list[0], input)
				parenthesis2 := testParenthesis(t, parenthesis1.Inner().GetTokens()[0], "(100+foo)")
				testOperator(t, parenthesis2.Inner().GetTokens()[0], "100+foo", "100", "+", "foo")
			},
		},
		{
			name:  "left parenthesis",
			input: "(100+foo)/100",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				operator := testOperator(t, list[0], input, "(100+foo)", "/", "100")
				parenthesis := testTokenList(t, operator.GetLeft(), 3).GetTokens()
				testOperator(t, parenthesis[1], "100+foo", "100", "+", "foo")
			},
		},
		{
			name:  "right parenthesis",
			input: "100/(100+foo)",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				operator := testOperator(t, list[0], input, "100", "/", "(100+foo)")
				parenthesis := testTokenList(t, operator.GetRight(), 3).GetTokens()
				testOperator(t, parenthesis[1], "100+foo", "100", "+", "foo")
			},
		},
		{
			name:  "invalid",
			input: "foo+",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testOperator(t, list[0], input, "foo", "+", "")
			},
		},
		{
			name:  "invalid with space",
			input: "foo + ",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testOperator(t, list[0], input, "foo", "+", "")
			},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			stmts := parseInit(t, tt.input)
			tt.checkFn(t, stmts, tt.input)
		})
	}
}

func TestParseComparison(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		checkFn func(t *testing.T, stmts []*ast.Statement, input string)
	}{
		{
			name:  "equal number",
			input: "foo = 25.5",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testComparison(t, list[0], input, "foo", "=", "25.5")
			},
		},
		{
			name:  "equal string",
			input: "foo = 'bar'",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testComparison(t, list[0], input, "foo", "=", "'bar'")
			},
		},
		{
			name:  "greater equal",
			input: "1 >= 2",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testComparison(t, list[0], input, "1", ">=", "2")
			},
		},
		{
			name:  "less equal",
			input: "1 <= 2",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testComparison(t, list[0], input, "1", "<=", "2")
			},
		},
		{
			name:  "equal left parenthesis",
			input: "(3 = 4) = 7",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				comparison := testComparison(t, list[0], input, "(3 = 4)", "=", "7")
				parenthesis := testTokenList(t, comparison.GetLeft(), 3).GetTokens()
				testComparison(t, parenthesis[1], "3 = 4", "3", "=", "4")
			},
		},
		{
			name:  "equal right parenthesis",
			input: "7 = (3 = 4)",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				comparison := testComparison(t, list[0], input, "7", "=", "(3 = 4)")
				parenthesis := testTokenList(t, comparison.GetRight(), 3).GetTokens()
				testComparison(t, parenthesis[1], "3 = 4", "3", "=", "4")
			},
		},
		{
			name:  "equal left function",
			input: "DATE(foo.bar) = bar.baz",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testComparison(t, list[0], input, "DATE(foo.bar)", "=", "bar.baz")
			},
		},
		{
			name:  "equal right function",
			input: "foo.bar = DATE(bar.baz)",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testComparison(t, list[0], input, "foo.bar", "=", "DATE(bar.baz)")
			},
		},
		{
			name:  "invalid",
			input: "foo=",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testComparison(t, list[0], input, "foo", "=", "")
			},
		},
		{
			name:  "invalid with space",
			input: "foo = ",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testComparison(t, list[0], input, "foo", "=", "")
			},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			stmts := parseInit(t, tt.input)
			tt.checkFn(t, stmts, tt.input)
		})
	}
}

func TestParseAliased(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		checkFn func(t *testing.T, stmts []*ast.Statement, input string)
	}{
		{
			name:  "alias",
			input: "foo AS bar",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testAliased(t, list[0], "foo AS bar", "foo", "bar")
			},
		},
		{
			name:  "alias without AS",
			input: "foo bar",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testAliased(t, list[0], "foo bar", "foo", "bar")
			},
		},
		{
			name:  "alias select identifier",
			input: "select foo as bar from mytable",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 7, input)
				list := stmts[0].GetTokens()
				testItem(t, list[0], "select")
				testItem(t, list[1], " ")
				testAliased(t, list[2], "foo as bar", "foo", "bar")
				testItem(t, list[3], " ")
				testItem(t, list[4], "from")
				testItem(t, list[5], " ")
				testIdentifier(t, list[6], "mytable")
			},
		},
		{
			name:  "alias from identifier",
			input: "select foo from mytable as mt",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 7, input)
				list := stmts[0].GetTokens()
				testItem(t, list[0], "select")
				testItem(t, list[1], " ")
				testIdentifier(t, list[2], "foo")
				testItem(t, list[3], " ")
				testItem(t, list[4], "from")
				testItem(t, list[5], " ")
				testAliased(t, list[6], "mytable as mt", "mytable", "mt")
			},
		},
		{
			name:  "alias join identifier",
			input: "select foo from abc inner join def as d",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 11, input)
				list := stmts[0].GetTokens()
				testItem(t, list[0], "select")
				testItem(t, list[1], " ")
				testIdentifier(t, list[2], "foo")
				testItem(t, list[3], " ")
				testItem(t, list[4], "from")
				testItem(t, list[5], " ")
				testIdentifier(t, list[6], "abc")
				testItem(t, list[7], " ")
				testMultiKeyword(t, list[8], "inner join")
				testItem(t, list[9], " ")
				testAliased(t, list[10], "def as d", "def", "d")
			},
		},
		{
			name:  "alias sub query",
			input: "select * from (select ci.ID, ci.Name from city as ci) as t",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 7, input)

				list := stmts[0].GetTokens()
				testAliased(t, list[6], "(select ci.ID, ci.Name from city as ci) as t", "(select ci.ID, ci.Name from city as ci)", "t")
				aliased := testTokenList(t, list[6], 5).GetTokens()
				testParenthesis(t, aliased[0], "(select ci.ID, ci.Name from city as ci)")

				parenthesis := testTokenList(t, aliased[0], 9).GetTokens()
				testItem(t, parenthesis[0], "(")
				testItem(t, parenthesis[1], "select")
				testItem(t, parenthesis[2], " ")
				testIdentifierList(t, parenthesis[3], "ci.ID, ci.Name")
				testItem(t, parenthesis[4], " ")
				testItem(t, parenthesis[5], "from")
				testItem(t, parenthesis[6], " ")
				testAliased(t, parenthesis[7], "city as ci", "city", "ci")
				testItem(t, parenthesis[8], ")")
			},
		},
		{
			name:  "alias in parenthesis",
			input: "(SELECT ci.ID AS city_id, ci.Name AS city_name FROM world.city AS ci)",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)

				list := stmts[0].GetTokens()
				testParenthesis(t, list[0], input)

				parenthesis := testTokenList(t, list[0], 9).GetTokens()
				testItem(t, parenthesis[0], "(")
				testItem(t, parenthesis[1], "SELECT")
				testItem(t, parenthesis[2], " ")
				testIdentifierList(t, parenthesis[3], "ci.ID AS city_id, ci.Name AS city_name")
				testItem(t, parenthesis[4], " ")
				testItem(t, parenthesis[5], "FROM")
				testItem(t, parenthesis[6], " ")
				testAliased(t, parenthesis[7], "world.city AS ci", "world.city", "ci")
				testItem(t, parenthesis[8], ")")
			},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			stmts := parseInit(t, tt.input)
			tt.checkFn(t, stmts, tt.input)
		})
	}
}

func TestParseIdentifierList(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		checkFn func(t *testing.T, stmts []*ast.Statement, input string)
	}{
		{
			name:  "simple",
			input: "foo, bar, foobar",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifierList(t, list[0], input)
			},
		},
		{
			name:  "with null",
			input: "foo, null, foobar",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifierList(t, list[0], input)
			},
		},
		{
			name:  "invalid single identifier without whitespace",
			input: "foo,",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifierList(t, list[0], input)
			},
		},
		{
			name:  "invalid single identifier include whitespace",
			input: "foo,  ",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifierList(t, list[0], input)
			},
		},
		{
			name:  "invalid multiple identifier without whitespace",
			input: "foo, bar,",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifierList(t, list[0], input)
			},
		},
		{
			name:  "invalid multiple identifier include whitespace",
			input: "foo, bar,  ",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifierList(t, list[0], input)
			},
		},
		{
			name:  "aliased member identiger",
			input: "`c`.`Name` as `country_name`, `cl`.`Language` as `country_language`",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifierList(t, list[0], input)
			},
		},
		{
			name:  "parenthesis",
			input: "(foo, bar, foobar)",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testParenthesis(t, list[0], input)
				parenthesis := list[0].(*ast.Parenthesis)
				tokens := parenthesis.Inner().GetTokens()
				il := testIdentifierList(t, tokens[0], "foo, bar, foobar")

				testIdentifierListGetIndex(t, il, genPosOneline(0), -1)
				testIdentifierListGetIndex(t, il, genPosOneline(1), 0)
				testIdentifierListGetIndex(t, il, genPosOneline(2), 0)
				testIdentifierListGetIndex(t, il, genPosOneline(3), 0)
				testIdentifierListGetIndex(t, il, genPosOneline(4), 0)
				testIdentifierListGetIndex(t, il, genPosOneline(5), 1)
				testIdentifierListGetIndex(t, il, genPosOneline(6), 1)
				testIdentifierListGetIndex(t, il, genPosOneline(7), 1)
				testIdentifierListGetIndex(t, il, genPosOneline(8), 1)
				testIdentifierListGetIndex(t, il, genPosOneline(9), 1)
				testIdentifierListGetIndex(t, il, genPosOneline(10), 2)
				testIdentifierListGetIndex(t, il, genPosOneline(11), 2)
				testIdentifierListGetIndex(t, il, genPosOneline(12), 2)
				testIdentifierListGetIndex(t, il, genPosOneline(13), 2)
				testIdentifierListGetIndex(t, il, genPosOneline(14), 2)
				testIdentifierListGetIndex(t, il, genPosOneline(15), 2)
				testIdentifierListGetIndex(t, il, genPosOneline(16), 2)
				testIdentifierListGetIndex(t, il, genPosOneline(17), 2)
				testIdentifierListGetIndex(t, il, genPosOneline(18), -1)
			},
		},
		{
			name:  "parenthesis list",
			input: "(foo, bar, foobar), (fooo, barr, fooobarr)",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 4, input)
				list := stmts[0].GetTokens()

				parenthesis1 := testParenthesis(t, list[0], "(foo, bar, foobar)")
				tokens1 := parenthesis1.Inner().GetTokens()
				testIdentifierList(t, tokens1[0], "foo, bar, foobar")

				testItem(t, list[1], ",")
				testItem(t, list[2], " ")

				parenthesis2 := testParenthesis(t, list[3], "(fooo, barr, fooobarr)")
				tokens2 := parenthesis2.Inner().GetTokens()
				testIdentifierList(t, tokens2[0], "fooo, barr, fooobarr")
			},
		},
		{
			name:  "invalid parenthesis",
			input: "(foo, bar,",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testParenthesis(t, list[0], input)
			},
		},
		{
			name:  "invalid single IdentifierList in select statement",
			input: "select foo,  from abc",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 6, input)
				list := stmts[0].GetTokens()
				testItem(t, list[0], "select")
				testItem(t, list[1], " ")
				testIdentifierList(t, list[2], "foo,  ")
				testItem(t, list[3], "from")
				testItem(t, list[4], " ")
				testIdentifier(t, list[5], "abc")
			},
		},
		{
			name:  "invalid multiplue IdentifierList in select statement",
			input: "select foo, bar, from abc",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 6, input)
				list := stmts[0].GetTokens()
				testItem(t, list[0], "select")
				testItem(t, list[1], " ")
				testIdentifierList(t, list[2], "foo, bar, ")
				testItem(t, list[3], "from")
				testItem(t, list[4], " ")
				testIdentifier(t, list[5], "abc")
			},
		},
		{
			name:  "IdentifierList function",
			input: "sum(a), sum(b)",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifierList(t, list[0], input)
			},
		},
		{
			name:  "IdentifierList Aliased",
			input: "sum(a) as x, b as y",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifierList(t, list[0], input)
			},
		},
		{
			name:  "IdentifierList comparison",
			input: "1 > 2, 3 < 4, 5 = 6",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifierList(t, list[0], input)
			},
		},
		{
			name:  "IdentifierList operator",
			input: "1 + 2, 3 - 4, 5 * 6",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifierList(t, list[0], input)
			},
		},
		{
			name:  "IdentifierList tokens",
			input: "1, 'aaa', 'N'",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifierList(t, list[0], input)
			},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			stmts := parseInit(t, tt.input)
			tt.checkFn(t, stmts, tt.input)
		})
	}
}

func TestParseCase(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		checkFn func(t *testing.T, stmts []*ast.Statement, input string)
	}{
		{
			name:  "case only",
			input: "CASE WHEN 1 THEN 2 ELSE 3 END",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testSwitchCase(t, list[0], input)
			},
		},
		{
			name:  "case alias with as",
			input: "CASE WHEN 1 THEN 2 ELSE 3 END as foo",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testAliased(t, list[0], input, "CASE WHEN 1 THEN 2 ELSE 3 END", "foo")
			},
		},
		{
			name:  "case alias without as",
			input: "CASE WHEN 1 THEN 2 ELSE 3 END foo",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testAliased(t, list[0], input, "CASE WHEN 1 THEN 2 ELSE 3 END", "foo")
			},
		},
		{
			name:  "case identifier list",
			input: "foo, CASE WHEN 1 THEN 2 ELSE 3 END as onetwothree, bar",
			checkFn: func(t *testing.T, stmts []*ast.Statement, input string) {
				testStatement(t, stmts[0], 1, input)
				list := stmts[0].GetTokens()
				testIdentifierList(t, list[0], input)
			},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			stmts := parseInit(t, tt.input)
			tt.checkFn(t, stmts, tt.input)
		})
	}
}

func parseInit(t *testing.T, input string) []*ast.Statement {
	t.Helper()
	parsed, err := Parse(input)
	if err != nil {
		t.Fatalf("error %+v\n", err)
	}

	var stmts []*ast.Statement
	for _, node := range parsed.GetTokens() {
		stmt, ok := node.(*ast.Statement)
		if !ok {
			t.Fatalf("invalid type want Statement parsed %T", stmt)
		}
		stmts = append(stmts, stmt)
	}
	return stmts
}

func testTokenList(t *testing.T, node ast.Node, length int) ast.TokenList {
	t.Helper()
	list, ok := node.(ast.TokenList)
	if !ok {
		t.Fatalf("invalid type want GetTokens got %T", node)
	}
	if length != len(list.GetTokens()) {
		t.Fatalf("Statements does not contain %d statements, got %d", length, len(list.GetTokens()))
	}
	return list
}

func testStatement(t *testing.T, stmt *ast.Statement, length int, expect string) {
	t.Helper()
	if length != len(stmt.GetTokens()) {
		t.Fatalf("Statement does not contain %d nodes, got %d, (expect %q got: %q)", length, len(stmt.GetTokens()), expect, stmt.String())
	}
	if expect != stmt.String() {
		t.Errorf("expected %q, got %q", expect, stmt.String())
	}
}

func testItem(t *testing.T, node ast.Node, expect string) {
	t.Helper()
	item, ok := node.(*ast.Item)
	if !ok {
		t.Fatalf("invalid type want Item got %T", node)
	}
	if item != nil {
		if expect != item.String() {
			t.Errorf("expected %q, got %q", expect, item.String())
		}
	} else {
		t.Errorf("item is null")
	}
}

func testMemberIdentifier(t *testing.T, node ast.Node, expect, parent, child string) {
	t.Helper()
	mi, ok := node.(*ast.MemberIdentifier)
	if !ok {
		t.Fatalf("invalid type want MemberIdentifier got %T", node)
	}
	if expect != node.String() {
		t.Errorf("expected %q, got %q", expect, node.String())
	}
	if parent != "" {
		if mi.Parent != nil {
			if parent != mi.Parent.String() {
				t.Errorf("parent expected %q , got %q", parent, mi.Parent.String())
			}
		} else {
			t.Errorf("parent is nil , got %q", parent)
		}
	}
	if child != mi.GetChild().String() {
		t.Errorf("child expected %q , got %q", child, mi.Parent.String())
	}
}

func testIdentifier(t *testing.T, node ast.Node, expect string) {
	t.Helper()
	_, ok := node.(*ast.Identifier)
	if !ok {
		t.Fatalf("invalid type want Identifier got %T", node)
	}
	if expect != node.String() {
		t.Errorf("expected %q, got %q", expect, node.String())
	}
}

func testMultiKeyword(t *testing.T, node ast.Node, expect string) {
	t.Helper()
	_, ok := node.(*ast.MultiKeyword)
	if !ok {
		t.Fatalf("invalid type want MultiKeyword got %T", node)
	}
	if expect != node.String() {
		t.Errorf("expected %q, got %q", expect, node.String())
	}
}

func testOperator(t *testing.T, node ast.Node, expect string, left, ope, right string) *ast.Operator {
	t.Helper()
	operator, ok := node.(*ast.Operator)
	if !ok {
		t.Fatalf("invalid type want Operator got %T", node)
	}
	if expect != node.String() {
		t.Errorf("expected %q, got %q", expect, node.String())
	}
	if left != operator.GetLeft().String() {
		t.Errorf("expected left %q, got %q", left, operator.Left.String())
	}
	if ope != operator.GetOperator().String() {
		t.Errorf("expected operator %q, got %q", ope, operator.GetOperator().String())
	}
	if right != operator.GetRight().String() {
		t.Errorf("expected right %q, got %q", right, operator.GetRight().String())
	}
	return operator
}

func testComparison(t *testing.T, node ast.Node, expect string, left, comp, right string) *ast.Comparison {
	t.Helper()
	comparison, ok := node.(*ast.Comparison)
	if !ok {
		t.Fatalf("invalid type want Comparison got %T", node)
	}
	if expect != node.String() {
		t.Errorf("expected %q, got %q", expect, node.String())
	}
	if left != comparison.GetLeft().String() {
		t.Errorf("expected left %q, got %q", left, comparison.GetLeft().String())
	}
	if comp != comparison.GetComparison().String() {
		t.Errorf("expected comparison %q, got %q", comp, comparison.GetComparison().String())
	}
	if right != comparison.GetRight().String() {
		t.Errorf("expected right %q, got %q", right, comparison.GetRight().String())
	}
	return comparison
}

func testParenthesis(t *testing.T, node ast.Node, expect string) *ast.Parenthesis {
	t.Helper()
	parenthesis, ok := node.(*ast.Parenthesis)
	if !ok {
		t.Errorf("invalid type want Parenthesis got %T", node)
	}
	if expect != node.String() {
		t.Errorf("expected %q, got %q", expect, node.String())
	}
	return parenthesis
}

func testFunction(t *testing.T, node ast.Node, expect string) {
	t.Helper()
	_, ok := node.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("invalid type want Function got %T", node)
	}
	if expect != node.String() {
		t.Errorf("expected %q, got %q", expect, node.String())
	}
}

func testAliased(t *testing.T, node ast.Node, expect string, realName, aliasedName string) {
	t.Helper()
	aliased, ok := node.(*ast.Aliased)
	if !ok {
		t.Fatalf("invalid type want Identifier got %T", node)
		return
	}
	if expect != node.String() {
		t.Errorf("expected %q, got %q", expect, node.String())
	}
	if aliased.RealName != nil {
		if realName != aliased.RealName.String() {
			t.Errorf("expected %q, got %q", realName, aliased.RealName.String())
		}
	} else {
		t.Errorf("RealName is null")
	}
	if aliased.AliasedName != nil {
		if aliasedName != aliased.AliasedName.String() {
			t.Errorf("expected %q, got %q", aliasedName, aliased.AliasedName.String())
		}
	} else {
		t.Errorf("AliasedName is null")
	}
}

func testIdentifierList(t *testing.T, node ast.Node, expect string) *ast.IdentifierList {
	t.Helper()
	if expect != node.String() {
		t.Errorf("expected %q, got %q", expect, node.String())
	}
	il, ok := node.(*ast.IdentifierList)
	if !ok {
		t.Fatalf("invalid type want IdentifierList got %T", node)
	}
	return il
}

func testIdentifierListGetIndex(t *testing.T, il *ast.IdentifierList, pos token.Pos, expect int) {
	t.Helper()
	got := il.GetIndex(pos)
	if expect != got {
		t.Errorf("expected %d, got %d", expect, got)
	}
}

func testSwitchCase(t *testing.T, node ast.Node, expect string) {
	t.Helper()
	_, ok := node.(*ast.SwitchCase)
	if !ok {
		t.Fatalf("invalid type want SwitchCase got %T", node)
	}
	if expect != node.String() {
		t.Errorf("expected %q, got %q", expect, node.String())
	}
}

func testPos(t *testing.T, node ast.Node, pos, end token.Pos) {
	t.Helper()
	if !reflect.DeepEqual(pos, node.Pos()) {
		t.Errorf("PosExpected %+v, got %+v", pos, node.Pos())
	}
	if !reflect.DeepEqual(end, node.End()) {
		t.Errorf("EndExpected %+v, got %+v", end, node.End())
	}
}

func genPosOneline(col int) token.Pos {
	return token.Pos{Line: 0, Col: col}
}
