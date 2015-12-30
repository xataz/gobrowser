package main

import (
    	"fmt"
//    	"io"
	"io/ioutil"
	"log"
	"strconv"
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

type List struct {
    Name    string
    Content []string
//    Size []int
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

	var filename []string
//	var size []int
	var i int = 0

	files, _ := ioutil.ReadDir(config.Path + r.URL.Path)
	for _, f := range files {
		if f.IsDir() {
			filename=append(filename,f.Name()+"/")
//			size[i]=f.Size()
		} else {
			filename=append(filename,f.Name()+"\t\t("+strconv.FormatInt(f.Size(), 10)+")")
       //                 size[i]=f.Size()
		}
		i+=1
	}
	list := List{"Welcome", filename}
	tmpl, err := template.ParseFiles("static/index.html")

    	if err != nil {
        	http.Error(w, err.Error(), http.StatusInternalServerError)
        	log.Printf(err.Error())
    	}


	log.Printf("%s", list)
	tmpl.Execute(w, list)

}


func main() {
	config = readConfig()

	http.HandleFunc(config.WebRoot + "/", home)
	http.Handle(config.WebRoot + "/static/", http.StripPrefix(config.WebRoot + "/static", http.FileServer(http.Dir("static"))))

	log.Printf("Starting HTTP server on %s\n", config.Listen)
	log.Println(http.ListenAndServe(config.Listen, nil))
}

