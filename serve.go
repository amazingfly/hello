package main

import (
	"database/sql"
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

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
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

//Model is a model for a gorm table
type Model struct {
	ID uint `gorm:"primary_key"`
}

//Product is for the itemDB
type Product struct {
	ID    int `sql:"AUTO_INCREMENT" gorm:"primary_key"`
	Name  string
	Qty   int
	Price float64
}

//Wiki is the wiki page table
type Wiki struct {
	ID          int    `sql:"AUTO_INCREMENT" gorm:"primary_key"`
	Title       string `sql:"size:255;unique;index"`
	FirstName   string
	LastName    string
	Body        string
	ViewCount   int
	DateCreated time.Time `sql:"DEFAULT:current_timestamp"`
}

//SQL stuff
func gormDB() gorm.DB {
	//var product []Product
	//var wiki Wiki
	//Opens DB connection
	db, err := gorm.Open("postgres", "user=postgres password=postgres dbname=wiki sslmode=disable")
	if err != nil {
		log.Printf("ERROR: %s\n", err)
	}
	//defer db.Close()
	//create tables
	//db.CreateTable(&wiki)

	//db.Create(&Wiki{Title: "derek", FirstName: "Derek", LastName: "Stich", Body: "empty", ViewCount: 0})
	//Find records
	/*
		db = db.Where("FirstName LIKE ?", "%ere%").Find(&wiki)
		fmt.Println("===========================")
		r, err := db.Rows()
		if err != nil {
			fmt.Println(err)
		}
		if r != nil {
			for r.Next() {
				var id int
				var name string
				var qty int
				var price float64

				err = r.Scan(&id, &name, &qty, &price)
				if err != nil {
					log.Println(err)
				}
				fmt.Println(id)
				fmt.Println(name)
				fmt.Println(qty)
				fmt.Println(price)
			}
		} else {
			log.Println("rows was nil")
		}
	*/
	return *db
}

func loadRecord(title string) (*Wiki, bool) {
	var db = gormDB()
	var notFound bool
	//var wiki Wiki

	db = *db.Where("Title = ?", title).Find(&Wiki{Title: title})
	//db = *db.FirstOrCreate(&Wiki{Title: title}, &Wiki{Title: title})
	notFound = db.RecordNotFound()

	fmt.Printf("a======== %v\n", notFound)
	r, err := db.Rows()
	if err != nil {
		fmt.Println(err)
	}
	defer r.Close()
	if r != nil {
		for r.Next() {
			var id int
			var title, firstName, lastName, body string
			var dateCreated time.Time
			var viewCount int

			err = r.Scan(&id, &title, &firstName, &lastName, &body, &viewCount, &dateCreated)
			if err != nil {
				log.Println(err)
			}
			fmt.Println(id)
			fmt.Println(title)
			fmt.Println(firstName)
			fmt.Println(lastName)
			fmt.Println(body)
			fmt.Println(viewCount)
			fmt.Println(dateCreated)
			fmt.Println("=== RETURNING ===")

			return (&Wiki{ID: id, Title: title, FirstName: firstName, LastName: lastName, Body: body, ViewCount: viewCount, DateCreated: dateCreated}), notFound
		}
	} else {
		log.Println("rows was nil")
	}
	return nil, notFound
}
func openDB() {
	//start DB connection
	db, err := sql.Open("postgres", "user=postgres password=postgres dbname=wiki sslmode=disable")
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	//ping to test the connection worked
	err = db.Ping()
	if err != nil {
		log.Println(err)
	}
	p, err := db.Prepare("SELECT * FROM products WHERE NAME LIKE '%' || $1 || '%';")
	if err != nil {
		log.Println(err)
	}
	r, err := p.Query("pep")
	if err != nil {
		log.Println(err)
	}
	defer r.Close()
	for r.Next() {
		var id int
		var name string
		var qty int
		var price float64

		err = r.Scan(&id, &name, &qty, &price)
		if err != nil {
			log.Println(err)
		}
		fmt.Println(id)
		fmt.Println(name)
		fmt.Println(qty)
		fmt.Println(price)
	}
}

func (p *Page) save() error {
	log.Println("saving page")
	filename := p.Title + ".wiki"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPageHTML(title string) (*Page, error) {
	log.Println("Loading page html")
	filename := title + ".html"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func loadPage(title string) (*Page, error) {
	log.Println("loading page")
	filename := title + ".wiki"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	log.Println("rendering template")
	err = templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("=================")
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
	/*
		p, err := loadPage(title)
		if err != nil {
			log.Printf("=======error: %s\n", err)
			http.Redirect(w, r, "/edit/"+title, http.StatusFound)
			return
		}
	*/
	p, notFound := loadRecord(title)
	if notFound == true {
		editHandler(w, r, title)
	} else {
		renderTemplate(w, "view", p)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	log.Println("edit handler loaded")
	/*
		p, err := loadPage(title)
		if err != nil {
			p = &Page{Title: title}
		}
	*/
	p, notFound := loadRecord(title)
	if notFound == true {
		var db = gormDB()
		db = *db.Create(&Wiki{Title: title})
		p, _ = loadRecord(title)
	}

	renderTemplate(w, "edit", p)
}
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	log.Println("Save handler loaded")
	/*
		body := r.FormValue("body")
		p := &Page{Title: title, Body: []byte(body)}
		err = p.save()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	*/
	var wiki Wiki
	p, notFound := loadRecord(title)
	_ = notFound
	db := gormDB()
	db = *db.First(&p)
	p.Body = r.FormValue("body")
	fmt.Printf("body= %s\n", wiki.Body)
	db.Save(p)
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}
func indexHandler(w http.ResponseWriter, r *http.Request, title string) {
	log.Println("index handler loaded")
	p, err := loadPageHTML(title)
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
	gormDB()
	fmt.Println("==============")
	fmt.Println("")
	//openDB()
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
