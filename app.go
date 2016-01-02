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
	"strconv"
)

type Config struct {
	Listen      string `json:"listen"`
	WebRoot	    string `json:"webroot"`
	Path	    string `json:"path"`
	HiddenFile  bool `json:"hiddenfile"`
}

type Files struct {
	Name	string
	Size	string
	Path	string
}

type Folders struct {
	Name	string
	Size	string
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
	file, _ := os.Open("app.conf")
	decoder := json.NewDecoder(file)
	config := Config{}
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error reading config: ", err)
	}
	return config
}

func convBytes(value int64) (result string) {
	var unit string
	unit="B"
	if value >= 1024 && value < 1024*1024 {
		value=value/1024
		unit="KB"
	} else if value >= 1024*1024 && value < 1024*1024*1024 {
		value=value/1024/1024
		unit="MB"
	} else if value >= 1024*1024*1024 && value < 1024*1024*1024*1024 {
		value=value/1024/1024/1024
		unit="GB"
	} else if value >= 1024*1024*1024*1024 {
		value=value/1024/1024/1024/1024
		unit="TB"
	}

	result=strconv.FormatInt(value, 10)+" "+unit

	return
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

func home(w http.ResponseWriter, r *http.Request) {
	config = readConfig()

	var content Content
	var path string
	var realPath string
	var backPath string

	path = r.URL.Path
	realPath = r.URL.Path[len(config.WebRoot):]
	files, _ := ioutil.ReadDir(config.Path + realPath)

	backPath = filepath.Dir(r.URL.Path)
	log.Printf("RealPath : %s, Url.path : %s, backPath : %s", realPath, r.URL.Path, backPath)
	if realPath != "/" {
		folder := Folders{"..", "4 KB", backPath}
		content.FolderList=append(content.FolderList, folder)
	}

	if f_isFile(config.Path + realPath) {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(config.Path + path)+"\"")
		http.ServeFile(w, r, config.Path + realPath)
	} else {
		if path != "/" {
			path=path+"/"
		}
		for _, f := range files {
			if ! config.HiddenFile {
				if f.Name()[0] != '.' {
					if f.IsDir() {
						folder := Folders{f.Name(), convBytes(f.Size()), path+f.Name()}
						content.FolderList=append(content.FolderList, folder)
					} else {
						file := Files{f.Name(), convBytes(f.Size()), path+f.Name()}
						content.FileList=append(content.FileList, file)
					}
				}
			} else {
				if f.IsDir() {
					folder := Folders{f.Name(), convBytes(f.Size()), path+f.Name()}
					content.FolderList=append(content.FolderList, folder)
				} else {
					file := Files{f.Name(), convBytes(f.Size()), path+f.Name()}
					content.FileList=append(content.FileList, file)
				}
			}
		}
		tmpl, err := template.ParseFiles("templates/index.html")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf(err.Error())
		}
		content.Name=realPath
		tmpl.Execute(w, content)
	}
}


func main() {
	config = readConfig()

	http.HandleFunc(config.WebRoot + "/", home)

	log.Printf("Starting goBrowser on %s\n", config.Listen)
	log.Println(http.ListenAndServe(config.Listen, nil))
}

