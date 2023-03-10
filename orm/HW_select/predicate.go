package orm

// 衍生類型
type op string

const (
	opAnd op = "AND"
	opOr  op = "OR"
	opNot op = "NOT"
	opEq  op = "="
	opLt  op = "<"
	opGt  op = ">"
)

func (o op) String() string {
	return string(o)
}

// Predicate 表示 一個 查詢條件
// 可以通過和 Predicate 組合成複雜的查詢條件
type Predicate struct {
	left  Expression
	op    op
	right Expression
}

func (Predicate) expr() {}

// Eq("id", 12)
//func Eq(left string, right ...any) *Predicate {
//	return &Predicate{
//		Col: left,
//		Op:  "=",
//		Arg: right,
//	}
//}

func Not(p Predicate) Predicate {
	return Predicate{
		op:    opNot,
		right: p,
	}
}

// Col("id").Eq(12).And(Col("name").Eq("Tom"))
func (left Predicate) And(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opAnd,
		right: right,
	}
}

// Col("id").Eq(12).Or(Col("name").Eq("Tom"))
func (left Predicate) Or(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opOr,
		right: right,
	}
}

// Expression 標記接口，代表表達式
type Expression interface {
	expr()
}

func exprOf(e any) Expression {
	switch expr := e.(type) {
	case Expression:
		return expr
	default:
		return valueOf(expr)
	}
}
