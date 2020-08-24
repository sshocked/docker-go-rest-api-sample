package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type Article struct {
	Title   string `json:"Title"`
	Desc    string `json:"desc"`
	Content string `json:"content"`
}

type Articles []Article

func allArticles(w http.ResponseWriter, r *http.Request) {
	articles := getArticles()

	fmt.Println("endpoint hit articles")

	json.NewEncoder(w).Encode(articles)
}

func getArticles() []Article {
	articles := Articles{
		Article{Title: "title of article", Desc: "descprition full", Content: "hi there"},
	}

	return articles
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "homepage endpont hit")
}

func handleRequest() {
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", homePage).Methods(http.MethodGet)
	r.HandleFunc("/articles", allArticles)
	log.Fatal(http.ListenAndServe(":10000", r))
}

func main() {
	handleRequest()
}
