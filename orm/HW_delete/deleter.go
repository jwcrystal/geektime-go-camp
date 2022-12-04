package orm

import (
	"geektime-go/orm/HW_delete/model"
	"geektime-go/orm/internal/errs"
	"strings"
)

type Deleter[T any] struct {
	table string
	where []Predicate
	model *model.Model
	sb    strings.Builder
	args  []any
}

func (d *Deleter[T]) From(table string) *Deleter[T] {
	d.table = table
	return d
}

func (d *Deleter[T]) Where(p ...Predicate) *Deleter[T] {
	d.where = p
	return d
}

func (d *Deleter[T]) Build() (*Query, error) {
	var err error
	d.model, err = model.ParseModel(new(T))
	if err != nil {
		return nil, err
	}
	d.sb.WriteString("DELETE FROM ")
	if d.table == "" {
		d.sb.WriteByte('`')
		d.sb.WriteString(d.model.TableName)
		d.sb.WriteByte('`')
	} else {
		d.sb.WriteString(d.table)
	}
	// 構建 Where
	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")
		// 合併多個 predicate
		p := d.where[0]
		for i := 1; i < len(d.where); i++ {
			p = p.And(d.where[i])
		}
		if err := d.buildExpression(p); err != nil {
			return nil, err
		}
	}
	d.sb.WriteByte(';')
	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, nil
}

func (d *Deleter[T]) buildExpression(e Expression) error {
	switch expr := e.(type) {
	case nil:
	case Column:
		field, ok := d.model.FieldMap[expr.name]
		if !ok {
			return errs.NewErrUnknownField(expr.name)
		}
		d.sb.WriteByte('`')
		d.sb.WriteString(field.ColName)
		d.sb.WriteByte('`')
	case value:
		d.sb.WriteByte('?')
		d.addArg(expr.val)
	case Predicate:
		_, lp := expr.left.(Predicate)
		if lp {
			d.sb.WriteByte('(')
		}
		if err := d.buildExpression(expr.left); err != nil {
			return err
		}
		if lp {
			d.sb.WriteByte(')')
		}

		if expr.op != "" {
			d.sb.WriteByte(' ')
			d.sb.WriteString(expr.op.String())
			d.sb.WriteByte(' ')
		}

		_, rp := expr.right.(Predicate)
		if rp {
			d.sb.WriteByte('(')
		}
		if err := d.buildExpression(expr.right); err != nil {
			return err
		}
		if rp {
			d.sb.WriteByte(')')
		}
	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

func (d *Deleter[T]) addArg(vals ...any) {
	if len(vals) == 0 {
		return
	}
	if d.args == nil {
		d.args = make([]any, 0, 8)
	}
	d.args = append(d.args, vals...)
}
