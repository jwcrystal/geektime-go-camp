package orm

import (
	"context"
	"geektime-go/orm/internal/errs"
	"geektime-go/orm/model"
	"strings"
)

//type Selector[T any] interface {
//	Form(table string) Selector[T]
//	Where(where string, args ...any) Selector[T]
//	OrderBy(order string) Selector[T]
//}

type Selector[T any] struct {
	table string
	where []Predicate
	model *model.Model
	sb    strings.Builder
	args  []any
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
	s.model, err = model.ParseModel(new(T))

	if err != nil {
		return nil, err
	}
	//var sb strings.Builder
	s.sb.WriteString("SELECT * FROM ")
	if s.table == "" {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.TableName)
		s.sb.WriteByte('`')
	} else {
		s.sb.WriteString(s.table)
	}

	//args := make([]any, 0, 4)
	// 構建 Where
	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		// 合併多個 predicate
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
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

		if err = s.buildExpression(p); err != nil {
			return nil, err
		}
	}

	s.sb.WriteString(";")

	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) buildExpression(e Expression) error {
	switch expr := e.(type) {
	// nil: 因為沒有 left predicate, e.g. NOT
	case nil:
	case Column:
		field, ok := s.model.FieldMap[expr.name]
		if !ok {
			return errs.NewErrUnknownField(expr.name)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(field.ColName)
		s.sb.WriteByte('`')
		// 剩下不考慮
	case value:
		s.sb.WriteByte('?')
		//s.args = append(s.args, expr.val)
		s.addArg(expr.val)
		// 剩下不考慮
	case RawExpr:
		s.sb.WriteByte('(')
		s.sb.WriteString(expr.raw)
		s.addArg(expr.args...)
		s.sb.WriteByte(')')
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

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Selector[T]) GetMulti(ctx context.Context) (*T, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Selector[T]) From(table string) *Selector[T] {
	s.table = table
	return s
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

func (s *Selector[T]) addArg(vals ...any) {
	if len(vals) == 0 {
		return
	}
	if s.args == nil {
		// 給予固定內存，避免 slice 不必要的自動擴展
		s.args = make([]any, 0, 8)
	}
	s.args = append(s.args, vals...)
}
