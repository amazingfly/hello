package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		fn(w, r)

	}
}

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
	//title := r.URL.Path[len("/"):]
	title := "outlets"
	p, _ := loadPage(title)
	fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
}

func outletHandler(w http.ResponseWriter, r *http.Request) {
	signal := strings.SplitAfter(fmt.Sprint(r.URL), "=")
	isON := signal[2]
	signalSplit := strings.Split(signal[1], "&")
	outletNum := signalSplit[0]
	fmt.Println(signal)
	fmt.Println(isON)
	fmt.Println(signalSplit)
	fmt.Println(outletNum)

	cmd := exec.Command("on5outlet")
	err := cmd.Start()
	if err != nil {
		fmt.Println(err)
	}
}

func Serve() {
	http.HandleFunc("/", makeHandler(viewHandler))
	http.HandleFunc("/outlet", makeHandler(outletHandler))
	/*
		http.HandleFunc("/edit/", makeHandler(editHandler))
		http.HandleFunc("/save/", makeHandler(saveHandler))
		http.HandleFunc("/", makeHandler(indexHandler))
		http.HandleFunc("/searchFor", searchHandler)
	*/
	s := &http.Server{
		Addr:           ":8181",
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
func main() {
	fmt.Println("hello")
	Serve()
}
