package orm

type RawExpr struct {
	raw  string
	args []any
}

func (r RawExpr) selectable() {}

func (RawExpr) expr() {}

func Raw(expr string, args ...any) RawExpr {
	return RawExpr{
		raw:  expr,
		args: args,
	}
}

func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}
