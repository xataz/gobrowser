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
    "flag"
)

type Config struct {
    Listen      string `json:"listen"`
    WebRoot     string `json:"webroot"`
    Path        string `json:"path"`
    HiddenFile  bool `json:"hiddenfile"`
    ForceSSL    bool `json:"forcessl"`
    ForceUrl    string `json:"forceurl"`
    SharePath   string `json:"sharepath"`
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
    Name        string `json:"name"`
    ShareUrl        string `json:"shareurl"`
    GetUrl        string `json:"geturl"`
    File    string `json:"file"`
    WebRoot string `json:"webroot"`
}

type ShareList struct {
    List  []Share
    WebRoot string
}

type ErrStruct struct {
    File string
    WebRoot string
}

var config Config

func readConfig(configfile string) Config {
    file, _ := os.Open(configfile)
    decoder := json.NewDecoder(file)
    config := Config{}
    err := decoder.Decode(&config)
    if err != nil {
        fmt.Println("Error reading config: ", err)
    }
    return config
}

func readShareFile(sharefile string) (Share, error) {
    file, _ := os.Open(sharefile)
    decoder := json.NewDecoder(file)
    share := Share{}
    err := decoder.Decode(&share)
    if err != nil {
        fmt.Println("Error reading sharefile: ", err)
    }
    return share, err
}

func writeShareFile(sharefile string, share Share) {
   sharefileJSON, err := json.MarshalIndent(&share, "", "\t\t")
   if err != nil {
      fmt.Println("Error marshalling to JSON:", err)
      return
   }
   err = ioutil.WriteFile(sharefile, sharefileJSON, 0644)
   if err != nil {
      fmt.Println("Error encoding JSON to sharefile:", err)
      return
   }
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
    //path := r.URL.Path
    fileShare := r.URL.Path[len(config.WebRoot+"/getshare"):]

    share, err := readShareFile(config.SharePath+fileShare)
    if err != nil {
        fmt.Println(err)
        tmpl, err := template.ParseFiles("templates/notfound.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            log.Printf(err.Error())
        }
        Err := ErrStruct{fileShare, config.WebRoot}
        tmpl.Execute(w, Err)
        return
    }
    filePath := share.File
    
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        tmpl, err := template.ParseFiles("templates/notfound.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            log.Printf(err.Error())
        }
        Err := ErrStruct{fileShare, config.WebRoot}
        tmpl.Execute(w, Err)
        return
    }
    
    w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(filePath)+"\"")
    http.ServeFile(w, r, filePath)
    log.Printf(filePath)
}

func createshare(w http.ResponseWriter, r *http.Request) {
    var Chars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
    var HTTP string
    if config.ForceUrl != "" { r.Host = config.ForceUrl }
    if config.ForceSSL { HTTP = "https://" } else { HTTP = "http://" }
    newLink := make([]byte, 32)
    
    rand.Read(newLink)
    for i, b := range newLink {
        newLink[i] = Chars[b % byte(len(Chars))]
    }
    
    fileShare := r.URL.Path[len(config.WebRoot+"/createshare"+config.WebRoot):]
    fileShare = config.Path + fileShare
    if _, err := os.Stat(fileShare); os.IsNotExist(err) {
        w.Write([]byte("Error : File "+fileShare+" not found"))
        return
        
    }

    s := Share{}
    s.Name = string(newLink[:])
    s.ShareUrl = HTTP+r.Host+config.WebRoot+"/share/"+s.Name
    s.GetUrl = HTTP+r.Host+config.WebRoot+"/getshare/"+s.Name
    s.File = fileShare
    s.WebRoot = config.WebRoot

    writeShareFile(config.SharePath+"/"+s.Name, s)
    
    w.Write([]byte("Share Url : "+HTTP+r.Host+config.WebRoot+"/share/"+s.Name)) 
}

func listshares(w http.ResponseWriter, r *http.Request) {
    shares, _ := ioutil.ReadDir(config.SharePath)
    var share []Share
    
    for _, f := range shares {
        shareTmp, _ := readShareFile(config.SharePath+"/"+f.Name())
        share=append(share, shareTmp)
        log.Printf("%s", share)
    }
    
    sharelist := ShareList{share, config.WebRoot}
    tmpl, _ := template.ParseFiles("templates/listshares.html")
    tmpl.Execute(w, sharelist)
}

func delshare(w http.ResponseWriter, r *http.Request) {
    fileDel := r.URL.Path[len(config.WebRoot+"/delshare/"):]
    
    if _, err := os.Stat(config.SharePath+"/"+fileDel); os.IsNotExist(err) {
        w.Write([]byte("Error : File "+fileDel+" not found"))
        return
    }
    
    os.Remove(config.SharePath+"/"+fileDel)
    
    w.Write([]byte("Share deleted"))
}

func viewshare(w http.ResponseWriter, r *http.Request) {
    fileShare := r.URL.Path[len(config.WebRoot+"/share"):]
    
    share, err := readShareFile(config.SharePath+fileShare)
    if err != nil {
        fmt.Println(err)
        tmpl, err := template.ParseFiles("templates/notfound.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            log.Printf(err.Error())
        }
        Err := ErrStruct{fileShare, config.WebRoot}
        tmpl.Execute(w, Err)
        return
    }
    
    tmpl, err := template.ParseFiles("templates/share.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        log.Printf(err.Error())
    }

    tmpl.Execute(w, share)
}


func initFlag() {
    Listen := flag.String("listen", "127.0.0.1:5000", "a string")
    WebRoot := flag.String("webroot", "", "a string")
    Path := flag.String("path", "/home", "a string")
    HiddenFile := flag.Bool("hiddenfile", false, "a bool")
    ForceUrl := flag.String("forceurl", "", "a string")
    ForceSSL := flag.Bool("forcessl", false, "a bool")
    SharePath := flag.String("sharepath", "share", "a string")
    ConfigFile := flag.String("config", "", "a string") 

    flag.Parse()
 
    if *ConfigFile != "" {
        if _, err := os.Stat(*ConfigFile); os.IsNotExist(err) {
            log.Printf("Configfile not found !!!")
        } else {
            log.Printf("Configfile found !!!")
            config = readConfig(*ConfigFile)
        }
    } else {
        config.Listen = *Listen
        config.WebRoot = *WebRoot
        config.Path = *Path
        config.HiddenFile = *HiddenFile
        config.ForceUrl = *ForceUrl
        config.ForceSSL = *ForceSSL
        config.SharePath = *SharePath
    }
    
}

func main() {
    initFlag()
    

    http.HandleFunc(config.WebRoot + "/", home)
    http.HandleFunc(config.WebRoot + "/share/", viewshare)
    http.HandleFunc(config.WebRoot + "/shareslist/", listshares)
    http.HandleFunc(config.WebRoot + "/createshare/", createshare)
    http.HandleFunc(config.WebRoot + "/getshare/", getshare)
    http.HandleFunc(config.WebRoot + "/delshare/", delshare)
    
    if _, err := os.Stat(config.SharePath); os.IsNotExist(err) {
        os.Mkdir(config.SharePath,0755)
    }

    log.Printf("Starting goBrowser on %s\n", config.Listen)
    log.Println(http.ListenAndServe(config.Listen, nil))
}

