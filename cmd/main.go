package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type ShortenLink struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

func startDatabase() {
	dir, _ := os.Getwd()
	currentData, err := ioutil.ReadFile(filepath.Join(dir, "database/db.json"))
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(currentData, &link)
}
func writeToDatabase() {
	dir, _ := os.Getwd()
	byteData, _ := json.MarshalIndent(link, "", " ")
	ioutil.WriteFile(filepath.Join(dir, "database/db.json"), byteData, 0644)
}

var link []ShortenLink

func createShortenUrl(w http.ResponseWriter, r *http.Request) {

	// make sure it is a post request
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "No OK method")
		return
	}

	// read parameter from body
	result := map[string]string{}
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "No OK body")
		return
	}

	// check if it is a valid url
	u, err := url.ParseRequestURI(result["url"])
	if err != nil || u.Scheme == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Not a URL")
		return
	}

	// if shorten link already existed, return it
	for _, v := range link {
		if v.URL == result["url"] {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Shorten link: http://localhost:8000/%s\n", v.ID)
			return
		}
	}

	// create new shorten link
	id := RandStringRunes(5)
	url := result["url"]
	shortenlink := ShortenLink{ID: id, URL: url}
	link = append(link, shortenlink)
	writeToDatabase()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Shorten link: http://localhost:8000/%s\n", id)

}

func accessURL(w http.ResponseWriter, r *http.Request) {

	// make sure it is a get request
	if r.Method != "GET" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "No OK method")
		return
	}

	// search for the link in db
	vars := mux.Vars(r)
	for _, v := range link {
		if v.ID == vars["url"] {
			http.Redirect(w, r, v.URL, 301)
			return
		}
	}
	w.WriteHeader(404)
	fmt.Fprintln(w, "Not found")
}

func main() {
	startDatabase()
	r := mux.NewRouter()
	s := r.PathPrefix("/api/v1").Subrouter()
	s.HandleFunc("/new", createShortenUrl)
	s.HandleFunc("/{url}", accessURL)

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	err := srv.ListenAndServe()

	if err != nil {
		fmt.Println(err)
	}
}
