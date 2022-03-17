package main

import (
	"embed"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var config Config
var users map[string]User
var processMutex sync.Mutex

//go:embed login.html
var embedFS embed.FS

// $Env:GOOS = "linux"; $Env:GOARCH = "amd64"

func main() {
	config = parseConfig()

	err := pullImages()
	if err != nil {
		log.Fatal(err)
	}

	users = map[string]User{}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", serve)
	r.Post("/launcher/login", login)
	r.Get("/{user}/accounts/logout/", logout)
	r.NotFound(proxy)

	log.Println("Server Ready")

	err = http.ListenAndServe(config.Serve, r)
	if err != nil {
		log.Fatal(err)
	}
}

func proxy(w http.ResponseWriter, r *http.Request) {
	user := r.Header.Get(config.RemoteUserHeader)

	if data, ok := users[user]; ok {
		data.Timeout.Reset(config.SessionTimeout)

		proxy := httputil.ReverseProxy{Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Host = "127.0.0.1:" + strconv.Itoa(data.Port)
			r.Host = r.URL.Host
		}, ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			time.Sleep(3 * time.Second)
			http.Redirect(w, r, r.URL.Path, 302)
		}}
		proxy.ServeHTTP(w, r)
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	user := r.Header.Get(config.RemoteUserHeader)
	email := r.Header.Get(config.RemoteEmailHeader)
	password := r.Form.Get("password")
	log.Println("login")
	// if volume is not mounted
	if _, ok := users[user]; !ok && user != "" && password != "" {
		processMutex.Lock()
		err := mount(user, password)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			http.Redirect(w, r, "/", 302)
			processMutex.Unlock()
			return
		}
		userData, err := spawnPaperless(getMountPath(user), user, password, email)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			unmount(user)
			http.Redirect(w, r, "/", 302)
			processMutex.Unlock()
			return
		}
		users[user] = userData
		processMutex.Unlock()
	}
	// redirect to /user
	http.Redirect(w, r, "/"+user, 302)
}

func serve(w http.ResponseWriter, r *http.Request) {
	user := r.Header.Get(config.RemoteUserHeader)
	// if volume is mounted (user logged in)
	if _, ok := users[user]; ok {
		// redirect to /user
		http.Redirect(w, r, "/"+user, 302)
	} else {
		// show login form
		p, err := embedFS.ReadFile("login.html")
		if err == nil {
			w.Write(p)
		}
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	user := r.Header.Get(config.RemoteUserHeader)
	logoutUser(user)
	// redirect to /
	http.Redirect(w, r, "/", 302)
}

func logoutUser(user string) {
	// if volume is mounted
	processMutex.Lock()
	if data, ok := users[user]; ok {
		err := killPaperless(user)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
		}
		err = unmount(user)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
		}
		data.Timeout.Stop()
		delete(users, user)
	}
	processMutex.Unlock()
}
