package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"html/template"
	"os"
	"encoding/json"
	"path/filepath"
)

type Config struct {
	Listen      string `json:"listen"`
	WebRoot	    string `json:"webroot"`
	Path	    string `json:"path"`
}

type Files struct {
	Name	string
	Size	int64
	Path	string
}

type Folders struct {
	Name	string
	Size	int64
	Path	string
}

type Content struct {
	Name    string
	FolderList []Folders
	FileList []Files
	BackPath string
}

var config Config

type ErrorMessage struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

type SuccessMessage struct {
	Delkey string `json:"delkey"`
}

func readConfig() Config {
	file, _ := os.Open("server.conf")
	decoder := json.NewDecoder(file)
	config := Config{}
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error reading config: ", err)
	}
	return config
}


func f_isFile(path string) (isFile bool) {

	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	switch mode := fi.Mode(); {
		case mode.IsDir():
			isFile = false
		case mode.IsRegular():
			isFile = true
	}

	return
}

func f_genBreadcrumb (path string) (breadcrumb []string) {
	for path != "/" {
		breadcrumb=append(breadcrumb, filepath.Base(path))
		path=filepath.Dir(path)
	}

	return
}

func home(w http.ResponseWriter, r *http.Request) {
	config = readConfig()

	var content Content
	var path string
	files, _ := ioutil.ReadDir(config.Path + r.URL.Path)
	var backPath string
	var breadcrumb []string

	backPath = filepath.Dir(r.URL.Path)
	log.Printf("backPath %s", backPath)

	if r.URL.Path != "/" {
		path=r.URL.Path+"/"
	} else {
                path=""
	}

	breadcrumb = f_genBreadcrumb(backPath)
	for index, value := range breadcrumb {
		log.Printf("bread %d : %s", index, value)
	}
	if f_isFile(config.Path + r.URL.Path) {
		log.Printf("C'est un fichier")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(config.Path + r.URL.Path)+"\"")
		http.ServeFile(w, r, config.Path + r.URL.Path)
	} else {
		for _, f := range files {
			if f.Name()[0] != '.' {
				//log.Printf("Path %s : Base %s : Abs %s", path+f.Name(), filepath.Base(path+f.Name()), filepath.Dir(path+f.Name()))
				if f.IsDir() {
					folder := Folders{f.Name(), f.Size(), path+f.Name()}
					content.FolderList=append(content.FolderList, folder)
				} else {
					file := Files{f.Name(), f.Size(), path+f.Name()}
					content.FileList=append(content.FileList, file)
				}
			}
		}
		tmpl, err := template.ParseFiles("static/index.html")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf(err.Error())
		}
		content.Name=r.URL.Path
		content.BackPath=filepath.Dir(r.URL.Path)
		tmpl.Execute(w, content)
	}
}


func main() {
	config = readConfig()

	http.HandleFunc(config.WebRoot + "/", home)
	http.Handle(config.WebRoot + "/static/", http.StripPrefix(config.WebRoot + "/static", http.FileServer(http.Dir("static"))))
	//http.Handle(config.WebRoot + config.Path, http.StripPrefix(config.WebRoot + config.Path, http.FileServer(http.Dir(config.Path))))

	log.Printf("Starting HTTP server on %s\n", config.Listen)
	log.Println(http.ListenAndServe(config.Listen, nil))
}

