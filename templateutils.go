package main

import (
	"html/template"
)

func (b *Blog) Route(Route string, Params ...string) string {
	return UrlToPath(b.router.Get(Route).URL(Params...))
}

func (b *Blog) AbsRoute(Route string, Params ...string) string {
	return UrlToAbsPath(b.router.Get(Route).URL(Params...))
}

func SafeHTML(content string) template.HTML {
	return template.HTML(content)
}
