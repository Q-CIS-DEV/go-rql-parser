package rqlParser

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

var IsValueError error = fmt.Errorf("Bloc is a value")

type TokenBloc []TokenString

func (tb TokenBloc) String() (s string) {
	for _, t := range tb {
		s = s + fmt.Sprintf("'%s' ", t.string)
	}
	return
}

type RqlNode struct {
	Op   string
	Args []interface{}
}

type Sort struct {
	By   string
	Desc bool
}

type RqlRootNode struct {
	Node   *RqlNode
	limit  string
	offset string
	sorts  []Sort
}

func (r *RqlRootNode) Limit() string {
	return r.limit
}

func (r *RqlRootNode) Offset() string {
	return r.offset
}

func (r *RqlRootNode) OffsetInt() int {
	i, err := strconv.Atoi(r.offset)
	if err != nil {
		return 0
	}
	return i
}

func (r *RqlRootNode) Sort() []Sort {
	return r.sorts
}

func parseLimit(n *RqlNode, root *RqlRootNode) (isLimitOp bool) {
	if n == nil {
		return false
	}
	if strings.ToUpper(n.Op) == "LIMIT" {
		root.offset = n.Args[0].(string)
		if len(n.Args) > 1 {
			root.limit = n.Args[1].(string)
		}
		isLimitOp = true
	}
	return
}
func parseSort(n *RqlNode, root *RqlRootNode) (isSortOp bool) {
	if n == nil {
		return false
	}
	if strings.ToUpper(n.Op) == "SORT" {
		for _, s := range n.Args {
			property := s.(string)
			desc := false

			if property[0] == '+' {
				property = property[1:]
			} else if property[0] == '-' {
				desc = true
				property = property[1:]
			}
			root.sorts = append(root.sorts, Sort{By: property, Desc: desc})
		}

		isSortOp = true
	}
	return
}

func (r *RqlRootNode) ParseSpecialOps() (err error) {
	if parseLimit(r.Node, r) || parseSort(r.Node, r) {
		r.Node = nil
	} else if r.Node != nil {
		if strings.ToUpper(r.Node.Op) == "AND" {
			limitIndex := -1
			sortIndex := -1
			for i, c := range r.Node.Args {
				switch n := c.(type) {
				case *RqlNode:
					if parseLimit(n, r) {
						limitIndex = i
					} else if parseSort(n, r) {
						sortIndex = i
					}
				}
			}
			if limitIndex >= 0 {
				if sortIndex > limitIndex {
					sortIndex = sortIndex - 1
				}
				if len(r.Node.Args) == 2 {
					keepIndex := 0
					if limitIndex == 0 {
						keepIndex = 1
					}
					r.Node = r.Node.Args[keepIndex].(*RqlNode)
				} else {
					r.Node.Args = append(r.Node.Args[:limitIndex], r.Node.Args[limitIndex+1:]...)
				}
			}
			if sortIndex >= 0 {
				if len(r.Node.Args) == 2 {
					keepIndex := 0
					if sortIndex == 0 {
						keepIndex = 1
					}
					r.Node = r.Node.Args[keepIndex].(*RqlNode)
				} else {
					r.Node.Args = append(r.Node.Args[:sortIndex], r.Node.Args[sortIndex+1:]...)
				}
			}
			if len(r.Node.Args) == 0 {
				r.Node = nil
			}
		}
	}

	return
}

type Parser struct {
	s *Scanner
}

func NewParser() *Parser {
	return &Parser{s: NewScanner()}
}

func (parser *Parser) Parse(r io.Reader) (root *RqlRootNode, err error) {
	var tokenStrings []TokenString
	if tokenStrings, err = parser.s.Scan(r); err != nil {
		return nil, err
	}

	root = &RqlRootNode{}

	root.Node, err = parse(tokenStrings)
	if err != nil {
		return nil, err
	}

	root.ParseSpecialOps()

	return
}

func getTokenOp(t Token) string {
	switch t {
	case AMPERSAND, COMMA:
		return "AND"
	case PIPE, SEMI_COLON:
		return "OR"
	}
	return ""
}

func parse(tokenStrings []TokenString) (node *RqlNode, err error) {
	var childNode *RqlNode

	node = &RqlNode{}

	if len(tokenStrings) == 0 {
		return nil, nil
	}
	//remove surrounding parenthesises
	if isParenthesisBloc(tokenStrings) && findClosingIndex(tokenStrings[1:]) == len(tokenStrings)-2 {
		tokenStrings = tokenStrings[1: len(tokenStrings)-1]
	}

	var childTokens [][]TokenString
	node.Op, childTokens = splitByBasisOp(tokenStrings)

	if node.Op == "" || len(childTokens) == 1 {
		return getBlocNode(tokenStrings)
	}

	for _, childToken := range childTokens {
		childNode, err = parse(childToken)
		if err != nil {
			if err == IsValueError {
				node.Args = append(node.Args, childToken[0].string)
			} else {
				return nil, err
			}
		} else {
			node.Args = append(node.Args, childNode)
		}
	}

	return node, nil
}

func isTokenInSlice(tokens []Token, tok Token) bool {
	for _, t := range tokens {
		if t == tok {
			return true
		}
	}
	return false
}

func splitByBasisOp(tb []TokenString) (op string, tbs [][]TokenString) {
	matchingToken := ILLEGAL

	prof := 0
	lastIndex := 0

	basisTokenGroups := [][]Token{
		[]Token{AMPERSAND, COMMA},
		[]Token{PIPE, SEMI_COLON},
	}
	for _, bt := range basisTokenGroups {
		btExtended := append(bt, ILLEGAL)
		for i, ts := range tb {
			if ts.token == OPENING_PARENTHESIS {
				prof++
			} else if ts.token == CLOSING_PARENTHESIS {
				prof--
			} else if prof == 0 {
				if isTokenInSlice(bt, ts.token) && isTokenInSlice(btExtended, matchingToken) {
					tbs = append(tbs, tb[lastIndex:i])
					matchingToken = ts.token
					lastIndex = i + 1
				}
			}
		}
		if lastIndex != 0 {
			break
		}
	}

	tbs = append(tbs, tb[lastIndex:])

	op = getTokenOp(matchingToken)

	return
}

func getBlocNode(tokenStrings []TokenString) (*RqlNode, error) {
	rqlNode := &RqlNode{}

	if len(tokenStrings) == 0 {
		rqlNode.Args = []interface{}{}
		rqlNode.Op = "single-value"
		return rqlNode, IsValueError
	} else if isValue(tokenStrings) {
		rqlNode.Args = []interface{}{tokenStrings[0].string}
		rqlNode.Op = "single-value"
		return rqlNode, IsValueError
	} else if isFuncStyleBloc(tokenStrings) {
		var err error

		rqlNode.Op = tokenStrings[0].string
		rqlNode.Args, err = parseFuncArgs(tokenStrings[2: findClosingIndex(tokenStrings[2:])+2])
		if err != nil {
			return nil, err
		}
	} else if isSimpleEqualBloc(tokenStrings) {
		rqlNode.Op = "eq"
		rqlNode.Args = []interface{}{tokenStrings[0].string, tokenStrings[2].string}

	} else if isDoubleEqualBloc(tokenStrings) {

		rqlNode.Op = tokenStrings[2].string
		rqlNode.Args = []interface{}{tokenStrings[0].string}
		if len(tokenStrings) == 5 {
			rqlNode.Args = append(rqlNode.Args, tokenStrings[4].string)
		} else if isParenthesisBloc(tokenStrings[4:]) && findClosingIndex(tokenStrings[5:]) == len(tokenStrings)-6 {
			args, err := parseFuncArgs(tokenStrings[5: len(tokenStrings)-1])
			if err != nil {
				return nil, err
			}
			rqlNode.Args = append(rqlNode.Args, args...)
		} else {
			return nil, fmt.Errorf("Unrecognized DoubleEqual bloc : " + TokenBloc(tokenStrings).String())
		}

	} else {
		return nil, fmt.Errorf("Unrecognized bloc : " + TokenBloc(tokenStrings).String())
	}

	return rqlNode, nil
}

func isValue(tb []TokenString) bool {
	return len(tb) == 0 || len(tb) == 1 && tb[0].token == IDENT
}

func isParenthesisBloc(tb []TokenString) bool {
	return tb[0].token == OPENING_PARENTHESIS
}

func isFuncStyleBloc(tb []TokenString) bool {
	return (tb[0].token == IDENT) && (tb[1].token == OPENING_PARENTHESIS)
}

func isSimpleEqualBloc(tb []TokenString) bool {
	isSimple := (tb[0].token == IDENT && tb[1].token == EQUAL_SIGN)
	if len(tb) > 3 {
		isSimple = isSimple && tb[3].token != EQUAL_SIGN
	}

	return isSimple
}

func isDoubleEqualBloc(tb []TokenString) bool {
	return tb[0].token == IDENT && tb[1].token == EQUAL_SIGN && tb[2].token == IDENT && tb[3].token == EQUAL_SIGN
}

func parseFuncArgs(tb []TokenString) (args []interface{}, err error) {
	var tokenGroups [][]TokenString

	indexes := findAllTokenIndexes(tb, COMMA)

	//split tokens by groups
	if len(indexes) == 0 {
		tokenGroups = append(tokenGroups, tb)
	} else {
		lastIndex := 0
		for _, i := range indexes {
			tokenGroups = append(tokenGroups, tb[lastIndex:i])
			lastIndex = i + 1
		}
		tokenGroups = append(tokenGroups, tb[lastIndex:])
	}

	for _, tokenGroup := range tokenGroups {
		rqlNode, err := parse(tokenGroup)
		if err != nil {
			if err == IsValueError {
				//use clean value
				if len(rqlNode.Args) > 0 {
					args = append(args, rqlNode.Args[0])
				}
			} else {
				return args, err
			}
		} else {
			args = append(args, rqlNode)
		}
	}

	return
}

func findClosingIndex(tb []TokenString) int {
	i := findTokenIndex(tb, CLOSING_PARENTHESIS)
	return i
}

func findTokenIndex(tb []TokenString, token Token) int {
	prof := 0
	for i, ts := range tb {
		if ts.token == OPENING_PARENTHESIS {
			prof++
		} else if ts.token == CLOSING_PARENTHESIS {
			if prof == 0 && token == CLOSING_PARENTHESIS {
				return i
			}
			prof--
		} else if token == ts.token && prof == 0 {
			return i
		}
	}
	return -1
}

func findAllTokenIndexes(tb []TokenString, token Token) (indexes []int) {
	prof := 0
	for i, ts := range tb {
		if ts.token == OPENING_PARENTHESIS {
			prof++
		} else if ts.token == CLOSING_PARENTHESIS {
			if prof == 0 && token == CLOSING_PARENTHESIS {
				indexes = append(indexes, i)
			}
			prof--
		} else if token == ts.token && prof == 0 {
			indexes = append(indexes, i)
		}
	}
	return
}

func __printTB(s string, tb []TokenString) {
	fmt.Printf("%s%s\n", s, TokenBloc(tb).String())
}
