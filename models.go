package main

import (
	"bytes"
	"encoding/json"
	"html/template"
)

type Profile struct {
	Github string
}

type PostData struct {
	Type    string
	Content *json.RawMessage
}

type Post interface {
	Render(t *template.Template) template.HTML
}

type HEntry struct {
}

type Note struct {
	HEntry
	Message string
	Draft   bool
}

const (
	TypeNote = "note"
)

func UnmarshalPost(data []byte) Post {
	var post PostData
	json.Unmarshal(data, &post)
	switch post.Type {
	case TypeNote:
		var note Note
		json.Unmarshal(*post.Content, &note)
		return note
	}
	return nil
}

func MarshalPost(post Post) []byte {
	contentdata, _ := json.Marshal(post)
	var data PostData
	contentjson := json.RawMessage(contentdata)
	data.Content = &contentjson
	switch post.(type) {
	case Note:
		data.Type = TypeNote
	}
	datadata, _ := json.Marshal(data)
	return datadata
}

func (n Note) Render(t *template.Template) template.HTML {
	buf := &bytes.Buffer{}
	t.ExecuteTemplate(buf, "note.tpl", n)
	return template.HTML(buf.String())
}
