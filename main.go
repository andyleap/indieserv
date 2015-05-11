// indieserv project main.go
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/andyleap/cookiestore"
	"github.com/andyleap/formbuilder"
	"github.com/andyleap/tartheme"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

type Blog struct {
	templates      *template.Template
	microtemplates *template.Template
	static         tartheme.Assets
	router         *mux.Router
	db             *bolt.DB
	fb             *formbuilder.FormBuilder
	c              *cookiestore.CookieStore
}

func UrlToPath(url *url.URL, err error) string {
	if err != nil {
		panic(err)
	}
	return url.Path
}

func main() {
	theme, _ := tartheme.LoadDir("theme")
	db, _ := bolt.Open("blog.db", 0666, nil)

	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("config"))
		tx.CreateBucketIfNotExists([]byte("posts"))

		return nil
	})

	blog := &Blog{}
	blog.templates = theme.Prefix("templates/").Templates()
	blog.microtemplates = theme.Prefix("templates/microformats/").Templates()
	blog.router = mux.NewRouter()
	blog.db = db

	blog.fb = formbuilder.New(theme.Prefix("templates/form/").Templates())
	blog.c = cookiestore.New("IndieServe")
	loginform := blog.fb.NewForm("login")
	loginform.NewString("Username", "Username", "Username", "")
	loginform.NewPassword("Password", "Password", "Password")
	btns := loginform.NewButtons()
	btns.AddButton("login", "Login", "primary")

	profileform := blog.fb.NewForm("profile")
	profileform.NewString("Github", "Github", "Github", "")
	btns = profileform.NewButtons()
	btns.AddButton("Save", "Save", "primary")

	postform := blog.fb.NewForm("post")
	postform.NewString("Message", "Message", "Message", "")
	postform.NewBool("Draft", "Draft")
	btns = postform.NewButtons()
	btns.AddButton("Post", "Post", "primary")

	blog.static = theme.Prefix("static/")

	blog.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", blog.static))
	blog.router.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", theme.Prefix("templates/")))

	mainchain := alice.New()
	authchain := mainchain.Append(blog.RequireLogin)
	blog.router.Handle("/", mainchain.ThenFunc(blog.Index)).Methods("GET").Name("Home")
	blog.router.Handle("/", authchain.ThenFunc(blog.ContentPost)).Methods("POST").Name("ContentPost")
	blog.router.Handle("/login", mainchain.ThenFunc(blog.Login)).Methods("GET").Name("Login")
	blog.router.Handle("/login", mainchain.ThenFunc(blog.LoginPost)).Methods("POST").Name("LoginPost")
	blog.router.Handle("/admin/profile", authchain.ThenFunc(blog.AdminProfile)).Methods("GET").Name("AdminProfile")
	blog.router.Handle("/admin/profile", authchain.ThenFunc(blog.AdminProfilePost)).Methods("POST").Name("AdminProfilePost")

	http.ListenAndServe(":3000", blog.router)
}

func (b *Blog) RequireLogin(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		s := b.c.GetSession(req)
		_, ok := s.Values["user"]
		if !ok {
			http.Redirect(rw, req, UrlToPath(b.router.Get("Login").URL()), http.StatusSeeOther)
			return
		}
		handler.ServeHTTP(rw, req)
	})
}

func (b *Blog) Index(rw http.ResponseWriter, req *http.Request) {
	var profile Profile
	var posts []Post
	s := b.c.GetSession(req)
	_, loggedin := s.Values["user"]

	b.db.View(func(tx *bolt.Tx) error {
		profiledata := tx.Bucket([]byte("config")).Get([]byte("Profile"))
		json.Unmarshal(profiledata, &profile)
		c := tx.Bucket([]byte("posts")).Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			post := UnmarshalPost(v)
			switch post := post.(type) {
			case Note:
				if loggedin || !post.Draft {
					posts = append(posts, post)
				}
			}
		}
		return nil
	})

	var postformrender template.HTML
	if loggedin {
		postform := b.fb.GetForm("post")
		postformrender = postform.Render(nil, UrlToPath(b.router.Get("ContentPost").URL()), "POST")
	}

	postsrendered := make([]struct {
		Rendered template.HTML
	}, 0, len(posts))

	for _, post := range posts {
		postsrendered = append(postsrendered, struct {
			Rendered template.HTML
		}{
			post.Render(b.microtemplates),
		})
	}

	data := struct {
		Name    string
		Profile Profile
		Posts   []struct {
			Rendered template.HTML
		}
		PostForm template.HTML
	}{
		"Index",
		profile,
		postsrendered,
		postformrender,
	}
	err := b.templates.ExecuteTemplate(rw, "index.tpl", data)
	if err != nil {
		fmt.Println(err)
	}
}

func (b *Blog) ContentPost(rw http.ResponseWriter, req *http.Request) {
	form := b.fb.GetForm("post")
	var data Note
	form.Parse(req.FormValue, &data)

	b.db.Update(func(tx *bolt.Tx) error {
		posts := tx.Bucket([]byte("posts"))
		id := NextKey(posts)
		posts.Put(id, MarshalPost(data))
		return nil
	})
	http.Redirect(rw, req, UrlToPath(b.router.Get("Home").URL()), http.StatusSeeOther)
}

func (b *Blog) Login(rw http.ResponseWriter, req *http.Request) {
	loginform := b.fb.GetForm("login")
	data := struct {
		Name     string
		FormName string
		Form     template.HTML
	}{
		"Login",
		"Login",
		loginform.Render(nil, "/login", "POST"),
	}

	if err := b.templates.ExecuteTemplate(rw, "form.tpl", data); err != nil {
		fmt.Println(err)
	}
}

func (b *Blog) LoginPost(rw http.ResponseWriter, req *http.Request) {
	form := b.fb.GetForm("login")
	data := struct {
		Username string
		Password string
	}{}
	form.Parse(req.FormValue, &data)

	if data.Username == "Vendan" && data.Password == "password" {
		s := b.c.GetSession(req)
		s.Values["user"] = "Vendan"
		s.Save(rw)
		http.Redirect(rw, req, UrlToPath(b.router.Get("AdminProfile").URL()), http.StatusSeeOther)
	}
}

func (b *Blog) AdminProfile(rw http.ResponseWriter, req *http.Request) {
	var profile *Profile

	b.db.View(func(tx *bolt.Tx) error {
		profiledata := tx.Bucket([]byte("config")).Get([]byte("Profile"))
		json.Unmarshal(profiledata, profile)
		return nil
	})
	profileform := b.fb.GetForm("profile")
	data := struct {
		Name     string
		FormName string
		Form     template.HTML
	}{
		"Profile",
		"Profile",
		profileform.Render(nil, UrlToPath(b.router.Get("AdminProfilePost").URL()), "POST"),
	}
	if err := b.templates.ExecuteTemplate(rw, "form.tpl", data); err != nil {
		fmt.Println(err)
	}
}

func (b *Blog) AdminProfilePost(rw http.ResponseWriter, req *http.Request) {
	form := b.fb.GetForm("profile")
	data := &Profile{}
	form.Parse(req.FormValue, data)
	b.db.Update(func(tx *bolt.Tx) error {
		config := tx.Bucket([]byte("config"))
		jsondata, _ := json.Marshal(data)
		config.Put([]byte("Profile"), jsondata)
		return nil
	})
	http.Redirect(rw, req, UrlToPath(b.router.Get("AdminProfile").URL()), http.StatusSeeOther)
}

func NextKey(b *bolt.Bucket) []byte {
	i, _ := b.NextSequence()
	key, _ := json.Marshal(i)
	return key
}
