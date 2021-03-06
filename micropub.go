package main

import (
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
)

func (b *Blog) MicroPubEndpoint(rw http.ResponseWriter, req *http.Request) {
	token := b.ia.GetReqAccessToken(req)
	if token == nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}
	hasScope := false
	for _, scope := range token.Scope {
		if scope == "post" {
			hasScope = true
		}
	}
	if !hasScope {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}
	switch req.FormValue("h") {
	case "entry":
		name := req.FormValue("name")
		//summary := req.FormValue("summary")
		content := req.FormValue("content")
		//published := req.FormValue("published")
		//updated := req.FormValue("updated")
		//category := req.FormValue("category")
		//slug := req.FormValue("slug")
		//location := req.FormValue("location")
		//in_reply_to := req.FormValue("in-reply-to")
		//repost_of := req.FormValue("repost-of")
		//syndication := req.FormValue("syndication")
		//mp_syndicate_to := req.FormValue("mp-syndicate-to")
		if name != "" {
			var entry Article
			entry.Title = name
			entry.Content = content
			entry.Published = time.Now()
			b.db.Update(func(tx *bolt.Tx) error {
				posts := tx.Bucket([]byte("posts"))
				posts.Put(TimeToID(entry.Published), MarshalPost(entry))
				return nil
			})
			links := ScanLinks(content)
			source, _ := b.router.Get("Post").URL("id", entry.Slug())
			for _, link := range links {
				log.Printf("Sending notification to %s", link.String())
				b.wm.SendNotification(link, source)
			}
			rw.Header().Set("Location", b.AbsRoute("Post", "id", entry.Slug()))
			rw.WriteHeader(http.StatusCreated)
			b.gp.Publish(source, []byte(""))
			index, _ := b.router.Get("Home").URL()
			b.gp.Publish(index, []byte(""))
			return
		} else if content != "" {
			var entry Note
			entry.Message = content
			entry.Published = time.Now()
			b.db.Update(func(tx *bolt.Tx) error {
				posts := tx.Bucket([]byte("posts"))
				posts.Put(TimeToID(entry.Published), MarshalPost(entry))
				return nil
			})
			links := ScanLinks(content)
			source, _ := b.router.Get("Post").URL("id", entry.Slug())
			for _, link := range links {
				log.Printf("Sending notification to %s", link.String())
				b.wm.SendNotification(link, source)
			}
			rw.Header().Set("Location", b.AbsRoute("Post", "id", entry.Slug()))
			rw.WriteHeader(http.StatusCreated)
			b.gp.Publish(source, []byte(""))
			index, _ := b.router.Get("Home").URL()
			b.gp.Publish(index, []byte(""))
			return
		}

	}
}
