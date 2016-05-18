package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// Page is a struct to hold web pages
type Page struct {
	Title string
	Body  []byte
}

func loadPage(title string) (*Page, error) {
	filename := title + ".html"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/"):]
	p, _ := loadPage(title)
	fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
}

func outletOnHandler(w http.ResponseWriter, r *http.Request) {
	outletNum := r.URL.Path[len("/outletOn/"):]
	fmt.Println(outletNum)
}

func main() {
	http.HandleFunc("/", viewHandler)
	http.HandleFunc("/outletOn", outletOnHandler)
	http.ListenAndServe(":9873", nil)
}
