package main

import (
	"encoding/json"
	"html/template"
)

type Profile struct {
	Github string
}

type PostData struct {
	Type    string
	Content json.RawMessage
}

type Post interface {
	Render() template.HTML
}

type HEntry struct {
}

type Note struct {
	HEntry
	Content string
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
		json.Unmarshal(post.Content, note)
		return note
	}
	return nil
}

func MarshalPost(post Post) []byte {
	contentdata, _ := json.Marshal(post)
	var data PostData
	data.Content = contentdata
	switch post.(type) {
	case Note:
		data.Type = TypeNote
	}
	datadata, _ := json.Marshal(data)
	return datadata
}

func (n Note) Render() template.HTML {
	return template.HTML(n.Content)
}
