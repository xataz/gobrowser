package main

import (
    	"fmt"
//    	"io"
	"io/ioutil"
	"log"
//	"strconv"
	"net/http"
	"html/template"
//	"net/url"
	"os"
//	"path"
	"encoding/json"
)

type Config struct {
	Listen      string `json:"listen"`
	WebRoot	    string `json:"webroot"`
	Path	    string `json:"path"`
}

type Files struct {
	Name 	string
	Size	int64
	Path	string
}

type Folders struct {
	Name 	string
	Size	int64
	Path	string
}

type Content struct {
	Name    string
	FolderList []Folders
	FileList []Files
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

func home(w http.ResponseWriter, r *http.Request) {
	config = readConfig()

	var content Content
	var path string
	files, _ := ioutil.ReadDir(config.Path + r.URL.Path)
	if r.URL.Path != "/" {
		path=r.URL.Path+"/"
	} else {
                path=""
	}
	for _, f := range files {
		if f.Name()[0] != '.' {
			if f.IsDir() {
				folder := Folders{f.Name(), f.Size(), path+f.Name()}
				content.FolderList=append(content.FolderList, folder)
				log.Printf("%s", path)
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

	log.Printf("%s", content)
	tmpl.Execute(w, content)

}


func main() {
	config = readConfig()

	http.HandleFunc(config.WebRoot + "/", home)
	http.Handle(config.WebRoot + "/static/", http.StripPrefix(config.WebRoot + "/static", http.FileServer(http.Dir("static"))))
	//http.Handle(config.WebRoot + config.Path, http.StripPrefix(config.WebRoot + config.Path, http.FileServer(http.Dir(config.Path))))

	log.Printf("Starting HTTP server on %s\n", config.Listen)
	log.Println(http.ListenAndServe(config.Listen, nil))
}

