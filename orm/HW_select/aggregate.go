package orm

// Aggregate 表示聚合函數
// e.g. AVG, MAX, MIN...etc
type Aggregate struct {
	fn    string
	arg   string
	alias string
}

// selectable 標記接口
func (a Aggregate) selectable() {}
func (a Aggregate) expr()       {}

func Avg(c string) Aggregate {
	return Aggregate{
		fn:  "AVG",
		arg: c,
	}
}

// Col("id").Eq(12)
func (a Aggregate) Eq(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opEq,
		right: exprOf(arg),
	}
}

func (a Aggregate) Gt(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opGt,
		right: exprOf(arg),
	}
}

func (a Aggregate) Lt(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opLt,
		right: exprOf(arg),
	}
}
