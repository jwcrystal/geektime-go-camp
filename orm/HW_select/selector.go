package orm

import (
	"context"
	"geektime-go/orm/HW_select/model"
	"geektime-go/orm/internal/errs"
	"strings"
)

//type Selector[T any] interface {
//	Form(table string) Selector[T]
//	Where(where string, args ...any) Selector[T]
//	OrderBy(order string) Selector[T]
//}

type Selector[T any] struct {
	table   string
	columns []Selectable
	where   []Predicate
	model   *model.Model
	sb      strings.Builder
	args    []any
	db      *DB
	groupBy []Column
	having  []Predicate // 同 where
	orderBy []OrderBy
	offset  int
	limit   int
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
	}
}

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
	return s
}

func (s *Selector[T]) From(table string) *Selector[T] {
	s.table = table
	return s
}

func (s *Selector[T]) Build() (*Query, error) {
	// get table name
	//var (
	//	t   T
	//	err error
	//)
	//m, err := model.ParseModel(&t)
	// or
	var err error
	// s.model, err = model.ParseModel(new(T))

	// Refactor:  be with Registry
	s.model, err = s.db.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	s.sb.WriteString("SELECT ")
	if err = s.buildColumns(); err != nil {
		return nil, err
	}
	s.sb.WriteString(" FROM ")
	if s.table == "" {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.TableName)
		s.sb.WriteByte('`')
	} else {
		s.sb.WriteString(s.table)
	}

	// 構建 Where
	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		if err = s.buildPredicates(s.where); err != nil {
			return nil, err
		}
	}
	// 構建 GroupBy
	if len(s.groupBy) > 0 {
		s.sb.WriteString(" GROUP BY ")
		if err = s.buildGroupBy(s.groupBy); err != nil {
			return nil, err
		}
	}
	// 構建 Having
	if len(s.having) > 0 {
		s.sb.WriteString(" HAVING ")
		if err = s.buildPredicates(s.having); err != nil {
			return nil, err
		}
	}
	// 構建 OrderBy
	if len(s.orderBy) > 0 {
		s.sb.WriteString(" ORDER BY ")
		if err = s.buildOrderBy(); err != nil {
			return nil, err
		}
	}
	// 構建 Offset x limit y
	// 不考慮 負數 校驗
	if s.offset > 0 {
		s.sb.WriteString(" OFFSET ?")
		s.addArgs(s.offset)
	}

	if s.limit > 0 {
		s.sb.WriteString(" LIMIT ?")
		s.addArgs(s.limit)
	}
	s.sb.WriteString(";")

	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) buildPredicates(ps []Predicate) error {
	// 合併多個 predicate
	p := ps[0]
	for i := 1; i < len(ps); i++ {
		p = p.And(ps[i])
	}
	// 構建 predicate
	// p.left
	// op
	// p.right

	//switch left := p.left.(type) {
	//case Column:
	//	sb.WriteByte('`')
	//	sb.WriteString(left.name)
	//	sb.WriteByte('`')
	//}
	//sb.WriteByte(' ')
	//sb.WriteString(p.op.String())
	//sb.WriteByte(' ')
	//
	//switch right := p.right.(type) {
	//case value:
	//	sb.WriteByte('?')
	//	args = append(args, right.val)
	//}

	if err := s.buildExpression(p); err != nil {
		return err
	}
	return nil
}

func (s *Selector[T]) buildGroupBy(groupBy []Column) error {
	for i, c := range groupBy {
		if i > 0 {
			s.sb.WriteByte(',')
		}
		if err := s.buildColumn(c.name, ""); err != nil {
			return err
		}
	}
	return nil
}

func (s *Selector[T]) buildOrderBy() error {
	// e.g. age, id ...etc, 表示先排序 age，再排序 id
	for idx, val := range s.orderBy {
		if idx > 0 {
			// 表示一個排序以上
			s.sb.WriteByte(',')
		}
		err := s.buildColumn(val.col, "")
		if err != nil {
			return err
		}
		s.sb.WriteByte(' ')
		s.sb.WriteString(val.order)
	}
	return nil
}

func (s *Selector[T]) buildColumns() error {
	if len(s.columns) == 0 {
		s.sb.WriteByte('*')
		return nil
	}
	for idx, col := range s.columns {
		if idx > 0 {
			s.sb.WriteByte(',')
		}
		switch val := col.(type) {
		case Column:
			if err := s.buildColumn(val.name, val.alias); err != nil {
				return err
			}
		case Aggregate:
			if err := s.buildAggregate(val, true); err != nil {
				return err
			}
		case RawExpr:
			s.sb.WriteString(val.raw)
			if len(val.args) != 0 {
				s.addArgs(val.args...)
			}
		default:
			return errs.NewErrUnsupportedSelectable(col)
		}
	}
	return nil
}

func (s *Selector[T]) buildColumn(c string, alias string) error {
	field, ok := s.model.FieldMap[c]
	if !ok {
		return errs.NewErrUnknownField(c)
	}
	s.sb.WriteByte('`')
	s.sb.WriteString(field.ColName)
	s.sb.WriteByte('`')
	if alias != "" {
		s.buildAs(alias)
	}
	return nil
}

func (s *Selector[T]) buildAggregate(expr Aggregate, aliasUsed bool) error {
	s.sb.WriteString(expr.fn)
	s.sb.WriteString("(`")
	field, ok := s.model.FieldMap[expr.arg]
	if !ok {
		return errs.NewErrUnknownField(expr.arg)
	}
	s.sb.WriteString(field.ColName)
	s.sb.WriteString("`)")
	if aliasUsed {
		s.buildAs(expr.alias)
	}
	return nil
}

func (s *Selector[T]) buildExpression(e Expression) error {
	switch expr := e.(type) {
	// nil: 因為沒有 left predicate, e.g. NOT
	case nil:
	case Column:
		// 純 column，沒有使用 alias
		return s.buildColumn(expr.name, "")
	case Aggregate:
		return s.buildAggregate(expr, false)
	case value:
		s.sb.WriteByte('?')
		//s.args = append(s.args, expr.val)
		s.addArgs(expr.val)
		// 剩下不考慮
	case RawExpr:
		//s.sb.WriteByte('(')
		//s.sb.WriteString(expr.raw)
		//s.addArgs(expr.args...)
		//s.sb.WriteByte(')')
		// 用戶自己管
		s.sb.WriteString(expr.raw)
		if len(expr.args) != 0 {
			s.addArgs(expr.args...)
		}
	case Predicate:
		_, lp := expr.left.(Predicate)
		if lp {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(expr.left); err != nil {
			return err
		}
		if lp {
			s.sb.WriteByte(')')
		}

		if expr.op != "" {
			s.sb.WriteByte(' ')
			s.sb.WriteString(expr.op.String())
			s.sb.WriteByte(' ')
		}

		_, rp := expr.right.(Predicate)
		if rp {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(expr.right); err != nil {
			return err
		}
		if rp {
			s.sb.WriteByte(')')
		}
	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

func (s *Selector[T]) buildAs(alias string) {
	if alias != "" {
		s.sb.WriteString(" AS ")
		s.sb.WriteByte('`')
		s.sb.WriteString(alias)
		s.sb.WriteByte('`')
	}
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	query, err := s.Build()
	if err != nil {
		return nil, err
	}
	// s.db 是 Selector 結構體 定義的 DB
	// s.db.db 是 結構體裡面使用的 sql.DB
	// 採用 QueryContext，可以跟 GetMulti 復用同一份處理結果集代碼
	rows, err := s.db.db.QueryContext(ctx, query.SQL, query.Args...)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, errs.ErrNoRows
	}
	tp := new(T)
	// TODO: 讀取元數據
	s.model, err = s.db.r.Get(tp)
	//if err != nil {
	//	return nil, err
	//}
	return tp, nil
}

func (s *Selector[T]) GetMulti(ctx context.Context) (*T, error) {
	//TODO implement me
	panic("implement me")
}

// ids := []int{1,2,3}
// Where("id in (?,?,?)", ids)
// Where("id in (?,?,?)", ids...)

//func (s *Selector[T]) Where(where string, args ...any) *Selector[T] {
//
//}

// cols 用於 WHERE的列，無法支持多種條件
// e.g. AND、OR、NOT...
//func (s *Selector[T]) Where(cols []string, args ...any) *Selector[T] {
//
//}

// 為了支持多樣化, 引入 Predicate 結構體
// e.g. AND、OR、NOT...

func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	s.where = ps
	return s
}

func (s *Selector[T]) addArgs(vals ...any) {
	if len(vals) == 0 {
		return
	}
	if s.args == nil {
		// 給予固定內存，避免 slice 不必要的自動擴展
		s.args = make([]any, 0, 8)
	}

	// Where("id in (?,?,?)", ids)
	// Where("id in (?,?,?)", ids...)
	s.args = append(s.args, vals...)
}

func (s *Selector[T]) GroupBy(cols ...Column) *Selector[T] {
	s.groupBy = cols
	return s
}

func (s *Selector[T]) Having(ps ...Predicate) *Selector[T] {
	s.having = ps
	return s
}

func (s *Selector[T]) OrderBy(orders ...OrderBy) *Selector[T] {
	s.orderBy = orders
	return s
}

func (s *Selector[T]) Offset(offset int) *Selector[T] {
	s.offset = offset
	return s
}

func (s *Selector[T]) Limit(limit int) *Selector[T] {
	s.limit = limit
	return s
}

type Selectable interface {
	selectable()
}

type OrderBy struct {
	col   string
	order string
}

func Asc(col string) OrderBy {
	return OrderBy{
		col:   col,
		order: "ASC",
	}
}

func Desc(col string) OrderBy {
	return OrderBy{
		col:   col,
		order: "DESC",
	}
}
