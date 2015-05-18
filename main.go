// indieserv project main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/andyleap/boltinspect"
	"github.com/andyleap/cookiestore"
	"github.com/andyleap/formbuilder"
	"github.com/andyleap/goindieauth"
	"github.com/andyleap/gopub"
	"github.com/andyleap/microformats"
	"github.com/andyleap/tartheme"
	"github.com/andyleap/webmention"

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
	ia             *goindieauth.IndieAuth
	profile        *Profile
	wm             *webmention.WebMention
	gp             *gopub.GoPub
	bi             *boltinspect.BoltInspect
}

type LoginInfo struct {
	Me       string
	Password string
}

var (
	Port = flag.Int("Port", 3000, "Specifies the port to listen on")
)

type Profile struct {
	Name     string
	Password string
	Host     string
	Github   string
}

func main() {
	flag.Parse()

	theme, _ := tartheme.LoadDir("theme")
	db, _ := bolt.Open("blog.db", 0666, nil)

	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("config"))
		tx.CreateBucketIfNotExists([]byte("posts"))
		tx.CreateBucketIfNotExists([]byte("subs"))
		tx.CreateBucketIfNotExists([]byte("mentions"))
		return nil
	})

	blog := &Blog{}

	data, _ := ioutil.ReadFile("profile.json")
	json.Unmarshal(data, &blog.profile)

	blog.templates = theme.Prefix("templates/").AddTemplates(template.New("default").Funcs(template.FuncMap{
		"Route":    blog.Route,
		"AbsRoute": blog.AbsRoute,
		"AutoLink": AutoLink,
		"SafeHTML": SafeHTML,
	}))
	blog.microtemplates = theme.Prefix("templates/microformats/").AddTemplates(template.New("default").Funcs(template.FuncMap{
		"Route":    blog.Route,
		"AbsRoute": blog.AbsRoute,
		"AutoLink": AutoLink,
		"SafeHTML": SafeHTML,
	}))
	mainrouter := mux.NewRouter()
	blog.router = mainrouter.Host(blog.profile.Host).Subrouter()
	blog.db = db

	blog.fb = formbuilder.New(theme.Prefix("templates/form/").Templates())
	blog.c = cookiestore.New("IndieServe")
	loginform := blog.fb.NewForm("login")
	loginform.NewHidden("me")
	loginform.NewHidden("token")
	loginform.NewPassword("Password", "password", "Password")
	btns := loginform.NewButtons()
	btns.AddButton("login", "Login", "primary")

	profileform := blog.fb.NewForm("profile")
	profileform.NewString("Name", "Name", "Name", "")
	profileform.NewString("HomeURL", "HomeURL", "HomeURL", "")
	profileform.NewString("Github", "Github", "Github", "")
	btns = profileform.NewButtons()
	btns.AddButton("Save", "Save", "primary")

	blog.static = theme.Prefix("static/")

	blog.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", blog.static))
	blog.router.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", theme.Prefix("templates/")))

	mainchain := alice.New()
	authchain := mainchain.Append(blog.RequireLogin)
	blog.router.Handle("/", mainchain.ThenFunc(blog.Index)).Methods("GET").Name("Home")
	blog.router.Handle("/post/{id}", mainchain.ThenFunc(blog.Post)).Methods("GET").Name("Post")
	blog.router.Handle("/admin/profile", authchain.ThenFunc(blog.AdminProfile)).Methods("GET").Name("AdminProfile")
	blog.router.Handle("/admin/profile", authchain.ThenFunc(blog.AdminProfilePost)).Methods("POST").Name("AdminProfilePost")

	blog.ia = goindieauth.New()
	blog.ia.InfoPage = blog.IAInfoPage
	blog.ia.LoginPage = blog.IALoginPage
	blog.ia.CheckLogin = blog.IACheckLogin

	blog.wm = webmention.New()
	blog.wm.Mention = blog.WMMention

	blog.gp = gopub.New(&SubStorage{blog})

	blog.bi = boltinspect.New(blog.db)
	blog.router.HandleFunc("/boltdb", blog.bi.InspectEndpoint).Name("BoltInspectEndpoint")

	blog.router.HandleFunc("/indieauth", blog.ia.AuthEndpoint).Name("IndieAuthEndpoint")
	blog.router.HandleFunc("/token", blog.ia.TokenEndpoint).Name("TokenEndpoint")
	blog.router.HandleFunc("/micropub", blog.MicroPubEndpoint).Name("MicroPubEndpoint")
	blog.router.HandleFunc("/webmention", blog.wm.WebMentionEndpoint).Name("WebMentionEndpoint")
	pubhubroute := blog.router.HandleFunc("/hub", blog.gp.HubEndpoint).Name("PubHubEndpoint")

	blog.gp.Hub, _ = pubhubroute.URL()

	http.ListenAndServe(fmt.Sprintf(":%d", *Port), blog.router)
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
			case Article:
				if loggedin || !post.Draft {
					posts = append(posts, post)
				}
			}
		}
		return nil
	})

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
	}{
		"Index",
		profile,
		postsrendered,
	}
	err := b.templates.ExecuteTemplate(rw, "index.tpl", data)
	if err != nil {
		fmt.Println(err)
	}
}

func (b *Blog) Post(rw http.ResponseWriter, req *http.Request) {
	var profile Profile
	var post Post
	s := b.c.GetSession(req)
	_, loggedin := s.Values["user"]

	b.db.View(func(tx *bolt.Tx) error {
		profiledata := tx.Bucket([]byte("config")).Get([]byte("Profile"))
		json.Unmarshal(profiledata, &profile)
		postbucket := tx.Bucket([]byte("posts"))
		postdata := postbucket.Get(TimeToID(SlugToTime(mux.Vars(req)["id"])))
		post = UnmarshalPost(postdata)
		return nil
	})
	switch post := post.(type) {
	case Note:
		if !loggedin && post.Draft {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
	case Article:
		if !loggedin && post.Draft {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
	case nil:
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	postrendered := post.Render(b.microtemplates)

	data := struct {
		Name    string
		Profile Profile
		Post    template.HTML
		Slug    string
	}{
		"Post",
		profile,
		postrendered,
		post.Slug(),
	}
	err := b.templates.ExecuteTemplate(rw, "post.tpl", data)
	if err != nil {
		fmt.Println(err)
	}
}

func (b *Blog) AdminProfile(rw http.ResponseWriter, req *http.Request) {
	profileform := b.fb.GetForm("profile")
	data := struct {
		Name     string
		FormName string
		Form     template.HTML
	}{
		"Profile",
		"Profile",
		profileform.Render(b.profile, UrlToPath(b.router.Get("AdminProfilePost").URL()), "POST"),
	}
	if err := b.templates.ExecuteTemplate(rw, "form.tpl", data); err != nil {
		fmt.Println(err)
	}
}

func (b *Blog) AdminProfilePost(rw http.ResponseWriter, req *http.Request) {
	form := b.fb.GetForm("profile")
	form.Parse(req.FormValue, &b.profile)
	data, _ := json.Marshal(b.profile)
	ioutil.WriteFile("profile.json", data, 0600)
	http.Redirect(rw, req, UrlToPath(b.router.Get("AdminProfile").URL()), http.StatusSeeOther)
}

func (b *Blog) IALoginPage(rw http.ResponseWriter, req *http.Request, user, token, client_id string) {
	loginform := b.fb.GetForm("login")
	formdata := struct {
		me    string
		token string
	}{
		user,
		token,
	}
	data := struct {
		Name     string
		FormName string
		Form     template.HTML
	}{
		"Login",
		"Login for " + client_id,
		loginform.Render(formdata, "", "POST"),
	}

	if err := b.templates.ExecuteTemplate(rw, "form.tpl", data); err != nil {
		fmt.Println(err)
	}
}

func (b *Blog) IAInfoPage(rw http.ResponseWriter, req *http.Request) {

}

func (b *Blog) IACheckLogin(rw http.ResponseWriter, req *http.Request, user, password string) bool {
	s := b.c.GetSession(req)
	_, ok := s.Values["user"]
	s.Save(rw)
	if strings.EqualFold(user, b.profile.Host) && (ok || password == b.profile.Password) {
		s.Values["user"] = b.AbsRoute("Home")
		return true
	}
	return false
}

func (b *Blog) WMMention(source, target *url.URL, data *microformats.Data) {
	req, _ := http.NewRequest("GET", target.String(), nil)
	log.Printf("WebMention from %s, to %s", source.String(), target.String())
	b.db.Update(func(tx *bolt.Tx) error {
		mentionbucket := tx.Bucket([]byte("mentions"))
		jsondata, _ := json.Marshal(data)
		mentionbucket.Put([]byte(fmt.Sprintf("%s-%s", source.String(), target.String())), jsondata)
		return nil
	})
	rm := &mux.RouteMatch{}
	b.router.Match(req, rm)
	originentry := getEntry(data)
	if rm.Route != nil {
		switch rm.Route.GetName() {
		case "Post":
			id := rm.Vars["id"]
			b.db.Update(func(tx *bolt.Tx) error {
				postbucket := tx.Bucket([]byte("posts"))
				postdata := postbucket.Get(TimeToID(SlugToTime(id)))
				post := UnmarshalPost(postdata)
				switch tpost := post.(type) {
				case Note:
					tpost.Mentions = append(tpost.Mentions, &Mention{
						Source:    source,
						Published: time.Now(),
						Data:      originentry,
					})
					post = tpost
				}
				postbucket.Put(TimeToID(SlugToTime(id)), MarshalPost(post))
				return nil
			})
		}
	}
}

func getEntry(data *microformats.Data) *microformats.MicroFormat {
	for _, item := range data.Items {
		entry := getEntryRecurse(item)
		if entry != nil {
			return entry
		}
	}
	return nil
}

func getEntryRecurse(item *microformats.MicroFormat) *microformats.MicroFormat {
	if stringInSlice("h-entry", item.Type) {
		return item
	}
	for _, subitem := range item.Children {
		entry := getEntryRecurse(subitem)
		if entry != nil {
			return entry
		}
	}
	for _, prop := range item.Properties {
		for _, propitem := range prop {
			if subitem, ok := propitem.(*microformats.MicroFormat); ok {
				entry := getEntryRecurse(subitem)
				if entry != nil {
					return entry
				}
			}
		}
	}
	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
