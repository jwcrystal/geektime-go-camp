package web

import (
	"bytes"
	"context"
	"html/template"
)

type TemplateEngine interface {
	// Render 渲染頁面
	// tplName 模板名，按名索引
	// data 渲染頁面用的數據
	Render(ctx context.Context, tplName string, data any) ([]byte, error)

	// 新增模板，不需要
	// 讓用戶自己管理 具體實現的模板
	// AddTemplate(tplName string, tpl []byte) error
}

type GoTemplateEngine struct {
	T *template.Template
}

func (g *GoTemplateEngine) Render(ctx context.Context, tplName string, data any) ([]byte, error) {
	bs := &bytes.Buffer{}
	err := g.T.ExecuteTemplate(bs, tplName, data)
	return bs.Bytes(), err

}

func (g *GoTemplateEngine) ParseGlob(pattern string) error {
	var err error
	g.T, err = template.ParseGlob(pattern)
	return err
}
