package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"time"

	"github.com/andyleap/microformats"
)

type PostData struct {
	Type    string
	Content *json.RawMessage
}

type Post interface {
	Render(t *template.Template) template.HTML
	Slug() string
}

type HEntry struct {
	Published time.Time
	Mentions  []*Mention
}

type Note struct {
	HEntry
	Message string
	Draft   bool
}

type Article struct {
	HEntry
	Title   string
	Content string
	Draft   bool
}

type Mention struct {
	Source    *url.URL
	Data      *microformats.MicroFormat
	Published time.Time
}

const (
	TypeNote    = "note"
	TypeArticle = "article"
)

func UnmarshalPost(data []byte) Post {
	var post PostData
	json.Unmarshal(data, &post)
	switch post.Type {
	case TypeNote:
		var note Note
		json.Unmarshal(*post.Content, &note)
		return note
	case TypeArticle:
		var article Article
		json.Unmarshal(*post.Content, &article)
		return article
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
	case Article:
		data.Type = TypeArticle
	}
	datadata, _ := json.Marshal(data)
	return datadata
}

func (n Note) Render(t *template.Template) template.HTML {
	buf := &bytes.Buffer{}

	if err := t.ExecuteTemplate(buf, "note.tpl", n); err != nil {
		fmt.Println(err)
	}
	return template.HTML(buf.String())
}

func (a Article) Render(t *template.Template) template.HTML {
	buf := &bytes.Buffer{}

	if err := t.ExecuteTemplate(buf, "article.tpl", a); err != nil {
		fmt.Println(err)
	}
	return template.HTML(buf.String())
}

func (e HEntry) MentionItems() []struct {
	Content string
	URL     string
	Mention *Mention
} {
	mentions := make([]struct {
		Content string
		URL     string
		Mention *Mention
	}, 0)
	for _, m := range e.Mentions {
		renContent := ""
		renURL := ""
		if content, ok := m.Data.Properties["content"]; ok && renContent == "" {
			if content, ok := content[0].(*microformats.MicroFormat); ok {
				renContent = content.Value
			}
		}
		if summary, ok := m.Data.Properties["summary"]; ok && renContent == "" {
			if summary, ok := summary[0].(string); ok {
				renContent = summary
			}
		}
		if name, ok := m.Data.Properties["name"]; ok && renContent == "" {
			if name, ok := name[0].(string); ok {
				renContent = name
			}
		}
		if url, ok := m.Data.Properties["url"]; ok {
			if url, ok := url[0].(string); ok {
				renURL = url
			}
		}
		mentions = append(mentions, struct {
			Content string
			URL     string
			Mention *Mention
		}{
			renContent,
			renURL,
			m,
		})
	}
	return mentions
}

func (h HEntry) Render(t *template.Template) template.HTML {
	return template.HTML("")
}

func (e HEntry) Slug() string {
	return TimeToSlug(e.Published)
}
