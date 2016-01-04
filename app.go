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
    "crypto/rand"
)

type Config struct {
    Listen      string `json:"listen"`
    WebRoot     string `json:"webroot"`
    Path        string `json:"path"`
    HiddenFile  bool `json:"hiddenfile"`
}

type Files struct {
    Name    string
    Size    string
    Path    string
}

type Folders struct {
    Name    string
    Path    string
}

type Content struct {
    Name        string
    WebRoot     string
    FolderList  []Folders
    FileList    []Files
}

type Share struct {
    Name        string
    Url        string
}

var config Config

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

func convBytes(value float64) (result string) {
    var unit string
    unit="B"

    if value >= 1024.0 && value < 1024.0*1024.0 {
        value=value/1024.0
    unit="KB"
    } else if value >= 1024.0*1024.0 && value < 1024.0*1024.0*1024.0 {
        value=value/1024.0/1024.0
        unit="MB"
    } else if value >= 1024.0*1024.0*1024.0 && value < 1024.0*1024.0*1024.0*1024.0 {
        value=value/1024.0/1024.0/1024.0
        unit="GB"
    } else if value >= 1024.0*1024.0*1024.0*1024.0 {
        value=value/1024.0/1024.0/1024.0/1024.0
        unit="TB"
    }

    result=strconv.FormatFloat(value, 'f', 2, 64)+" "+unit

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

    path := r.URL.Path
    urlPath := r.URL.Path[len(config.WebRoot):]
    realPath := config.Path + urlPath
    backPath := filepath.Dir(r.URL.Path)
    files, _ := ioutil.ReadDir(realPath)
    
    if _, err := os.Stat(realPath); os.IsNotExist(err) {
        tmpl, err := template.ParseFiles("templates/notfound.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            log.Printf(err.Error())
        }
        tmpl.Execute(w, urlPath)
        return
    }
    if urlPath != "/" {
        folder := Folders{"..", backPath}
        content.FolderList=append(content.FolderList, folder)
    }

    if f_isFile(realPath) {
        w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(realPath)+"\"")
        http.ServeFile(w, r, realPath)
    } else {
        if path != "/" {
            path=path+"/"
        }
        for _, f := range files {
            if ! config.HiddenFile {
                if f.Name()[0] != '.' {
                    if f.IsDir() {
                        folder := Folders{f.Name(), path+f.Name()}
                        content.FolderList=append(content.FolderList, folder)
                    } else {
                        file := Files{f.Name(), convBytes(float64(f.Size())), path+f.Name()}
                        content.FileList=append(content.FileList, file)
                    }
                }
            } else {
                if f.IsDir() {
                    folder := Folders{f.Name(), path+f.Name()}
                    content.FolderList=append(content.FolderList, folder)
                } else {
                    file := Files{f.Name(), convBytes(float64(f.Size())), path+f.Name()}
                    content.FileList=append(content.FileList, file)
                }
            }
        }
        tmpl, err := template.ParseFiles("templates/index.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            log.Printf(err.Error())
        }
        content.Name=urlPath
        content.WebRoot=config.WebRoot
        tmpl.Execute(w, content)
    }
}

func getshare(w http.ResponseWriter, r *http.Request) {
    config = readConfig()
    //path := r.URL.Path
    fileShare := r.URL.Path[len(config.WebRoot+"/getshare"):]
    
    dat, err := ioutil.ReadFile("share"+fileShare)
    if err != nil {
        fmt.Println(err)
        tmpl, err := template.ParseFiles("templates/notfound.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            log.Printf(err.Error())
        }
        tmpl.Execute(w, fileShare)
        return
    }
    filePath := string(dat)
    
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        tmpl, err := template.ParseFiles("templates/notfound.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            log.Printf(err.Error())
        }
        tmpl.Execute(w, fileShare)
        return
    }
    
    w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(filePath)+"\"")
    http.ServeFile(w, r, filePath)
    log.Printf(filePath)
}

func createshare(w http.ResponseWriter, r *http.Request) {
    config = readConfig()

    var Chars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
    newLink := make([]byte, 32)
    rand.Read(newLink)
    for i, b := range newLink {
        newLink[i] = Chars[b % byte(len(Chars))]
    }
    s := string(newLink[:])
    
    fileShare := r.URL.Path[len(config.WebRoot+"/createshare"+config.WebRoot):]
    fileShare = config.Path + fileShare
    _, errRead := ioutil.ReadFile(fileShare)
    if errRead != nil {
        w.Write([]byte("Error"))
    }
    
    err := ioutil.WriteFile("share/"+s, []byte(fileShare), 0644)
    if err != nil {
        fmt.Println(err)
    }
    
    w.Write([]byte("Share Url : http://"+r.Host+config.WebRoot+"/share/"+s)) 
    log.Printf("createshare %s", s) 
}

func showshare(w http.ResponseWriter, r *http.Request) {
    config = readConfig()
    
    
    shares, _ := ioutil.ReadDir("share")
    
    for _, f := range shares {
        dat, _ := ioutil.ReadFile("share/"+f.Name())
        fmt.Println(r)
        log.Printf("%s : %s", r.Host+config.WebRoot+"/getshare/"+f.Name(), string(dat))
    }
}

func delshare(w http.ResponseWriter, r *http.Request) {
    config = readConfig()
    //var content Content
    //var path string
    //var urlPath string

    log.Printf("delshare") 
}

func viewshare(w http.ResponseWriter, r *http.Request) {
    config = readConfig()
    
    fileShare := r.URL.Path[len(config.WebRoot+"/share"):]
    
    dat, err := ioutil.ReadFile("share"+fileShare)
    if err != nil {
        fmt.Println(err)
        tmpl, err := template.ParseFiles("templates/notfound.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            log.Printf(err.Error())
        }
        tmpl.Execute(w, fileShare)
        return
    }
    filePath := string(dat)
    
    tmpl, err := template.ParseFiles("templates/share.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        log.Printf(err.Error())
    }
    
    UrlPath := config.WebRoot+"/getshare/"+fileShare
    
    share := Share{filePath, UrlPath}
    tmpl.Execute(w, share)
    log.Printf("viewshare")
}

func main() {
    config = readConfig()

    http.HandleFunc(config.WebRoot + "/", home)
    http.HandleFunc(config.WebRoot + "/share/", viewshare)
    http.HandleFunc(config.WebRoot + "/shareslist/", showshare)
    http.HandleFunc(config.WebRoot + "/createshare/", createshare)
    http.HandleFunc(config.WebRoot + "/getshare/", getshare)
    http.HandleFunc(config.WebRoot + "/delshare/", delshare)
    
    if _, err := os.Stat("share"); os.IsNotExist(err) {
        os.Mkdir("share",0755)
    }

    log.Printf("Starting goBrowser on %s\n", config.Listen)
    log.Println(http.ListenAndServe(config.Listen, nil))
}
