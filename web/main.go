package main

import (
	"fmt"
	"github.com/azomio/courses/lesson2/pkg/render"
	"github.com/azomio/courses/lesson2/pkg/requester"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

var cfg = struct {
	Port        int
	UserAddr    string
	MovieAddr   string
	PaymentAddr string
}{
	Port:        8080,
	MovieAddr:   "http://localhost:8081",
	UserAddr:    "http://localhost:8082",
	PaymentAddr: "http://localhost:8083",
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", MainHandler)

	// Обработчик статических файлов
	fs := http.FileServer(http.Dir("assets"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Настройка шаблонизатора
	render.SetTemplateDir(".")
	render.SetTemplateLayout("layout.html")
	render.AddTemplate("main", "main.html")
	err := render.ParseTemplates()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting on port %d", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r))
}

type MainPage struct {
	Movies *[]Movie
	User   User
	PayURL string
}

type User struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	IsPaid bool   `json:"is_paid"`
}

type Movie struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Poster   string `json:"poster"`
	MovieUrl string `json:"movie_url"`
	IsPaid   bool   `json:"is_paid"`
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	page := MainPage{}

	var err error
	page.Movies, err = getMovies()
	if err != nil {
		log.Printf("Get movie error: %v", err)
	}

	page.User, err = getUser(r)
	if err != nil {
		log.Printf("Get user error: %v", err)
	} else {
		page.PayURL = cfg.PaymentAddr + "/checkout?uid=" + strconv.Itoa(page.User.ID)
	}

	render.RenderTemplate(w, "main", page)
}

func getMovies() (*[]Movie, error) {
	mm := &[]Movie{}
	err := requester.GetJSON(cfg.MovieAddr+"/movies", mm)
	if err != nil {
		return nil, err
	}

	return mm, nil
}

func getUser(r *http.Request) (usr User, err error) {
	ses, err := r.Cookie("session")
	if ses == nil {
		return usr, err
	}

	res := &struct {
		User
		Error string
	}{}
	err = requester.GetJSON(cfg.UserAddr+"/user?token="+ses.Value, res)
	if err != nil {
		return usr, err
	}

	if res.Error != "" {
		return usr, fmt.Errorf(res.Error)
	}

	usr.ID = res.ID
	usr.Name = res.Name
	usr.IsPaid = res.IsPaid

	return usr, nil
}
