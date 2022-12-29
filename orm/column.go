package orm

type Column struct {
	name  string
	alias string
}

func (c Column) selectable() {}

func (c Column) expr() {}

func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
	}
}

type value struct {
	val any
}

func (value) expr() {}

func valueOf(v any) value {
	return value{val: v}
}

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

func (c Column) Lt(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLt,
		right: value{val: arg},
	}
}

func (c Column) Gt(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opGt,
		right: value{val: arg},
	}
}
