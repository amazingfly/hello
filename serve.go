package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var err error
var fileCache []string
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

var name = make(chan string)
var response = make(chan string)

//Page is a struct to hold page data
type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	log.Println("saving page")
	filename := p.Title + ".wiki"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	log.Println("loading page")
	filename := title + ".html"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	log.Println("rendering template")
	err := templates.ExecuteTemplate(w, tmpl+".wiki", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("making handler")
		log.Printf("r.URL.Path= %s\n", r.URL.Path)
		if r.URL.Path != "/" {
			m := validPath.FindStringSubmatch(r.URL.Path)
			if m == nil {
				log.Println("m is nil")
				http.NotFound(w, r)
				return
			}
			log.Printf("m[2m= %s\n", m[2])
			fn(w, r, m[2])
		} else {
			fn(w, r, "index")
		}
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	log.Println("view handler loaded")
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	log.Println("edit handler loaded")
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	log.Println("Save handler loaded")
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}
func indexHandler(w http.ResponseWriter, r *http.Request, title string) {
	log.Println("index handler loaded")
	p, err := loadPage(title)
	if err != nil {
		log.Println(err)
	}
	fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
}
func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	log.Println("getting title")
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	name <- r.FormValue("searchKey")

	w.Header().Set("Content-Type", "application/JSON")
	w.Write([]byte(<-response))

	log.Println("response sent")
	if err != nil {
		fmt.Println(err)
	}
}

func cacheFileNames() {
	fmt.Println("starting cache")
	fileCache, err = filepath.Glob("*.wiki")
	if err != nil {
		log.Println(err)
	}
	var json = `{"Result":{ "Pages":[{`

	for a, b := range fileCache {
		fmt.Printf("a= %d, len= %d", a, len(fileCache))
		if a+1 < len(fileCache) {
			json = json + fmt.Sprintf(`"Name": "%s"},{`, strings.Split(b, ".")[0])
		} else {
			json = json + fmt.Sprintf(`"Name": "%s"}]}}`, strings.Split(b, ".")[0])
		}
	}
	fmt.Println(json)
	fmt.Println("sedning rsponse channel")
	response <- json
}

func search() {
	for {
		var n = <-name
		if n != "" {
			log.Printf("n= %s\n", n)
			for _, fileName := range fileCache {
				if n == strings.Split(fileName, ".")[0] {
					log.Printf("fileName[:5}= %s\n", fileName[:5])
					response <- fmt.Sprintf(`{"Result":{"Pages":[{"Name": "%s"}]}}`, n)
				}
				fmt.Println(n)
			}
		} else {
			fmt.Println("this happense")
			cacheFileNames()
		}
	}
}

func main() {
	go search()
	log.Println("Server started")
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/", makeHandler(indexHandler))
	http.HandleFunc("/searchFor", searchHandler)

	s := &http.Server{
		Addr:           ":8181",
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
	/*
		var err error
		var req *http.Request

		go func() {
			for {
				var searchJSON = fmt.Sprintf(`{"searchField": "%s"}`, <-name)

				req, err = http.NewRequest("POST", url, bytes.NewBuffer(PrepHMAC(searchJSON)))

				//adds record to db
				//no frontend access
				//req, err = http.NewRequest("POST", url, bytes.NewBuffer(PrepHMAC(AddItem())))

				if err != nil {
					fmt.Println(err)
				}
				//req.SetBasicAuth(authName, authPass)
				//creates client and sends request then gathers the response
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
				} else {
					defer resp.Body.Close()
					//Reads the body of the result then converts it into bytes then into gabs.Container
					body, err := ioutil.ReadAll(resp.Body)
					bodyByte := []byte(body)

					result, _ := gabs.ParseJSON(bodyByte)
					if err != nil {
						fmt.Printf("%s", err)
						os.Exit(1)
					}
					//Search for the results and send them as a string to func Handler
					response <- result.S("results").String()
				}
			}
		}()
		forever := make(chan bool)
		<-forever */
}
