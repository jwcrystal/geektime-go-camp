package orm

type Column struct {
	name string
}

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

type Predicate struct {
	left  Expression
	op    op
	right Expression
}

func (o op) String() string {
	return string(o)
}

// Eq("id", 12)
//func Eq(left string, right ...any) *Predicate {
//	return &Predicate{
//		Col: left,
//		Op:  "=",
//		Arg: right,
//	}
//}

func C(name string) Column {
	return Column{name: name}
}

// Col("id").Eq(12)
func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEq,
		right: value{val: arg},
	}
}

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

func (Predicate) expr() {}

func (Column) expr() {}

type value struct {
	val any
}

func (value) expr() {}
