package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var logPath *string
var updateDir *string
var templates = template.Must(template.ParseFiles("index.html"))

type home struct {
	Title string
}

type ClientLog struct {
	RecordTime int64
	UserID     uint64
	UserName   string
	Desc       string
	Content    string
}

func (c *ClientLog) save() error {
	path := *logPath + "/" + fmt.Sprintf("%d-%s", c.UserID, c.UserName)
	if !pathExists(path) {
		createDir(path)
	}
	filename := time.Unix(c.RecordTime, 0).Format("2006_01_02_03_04_05") + ".log"
	if len(c.Desc) > 0 {
		descFile, err := os.OpenFile(path+"/"+"desc.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		descFile.WriteString(filename + " : " + c.Desc + "\n")
		descFile.Close()
	}
	return ioutil.WriteFile(path+"/"+filename, []byte(c.Content), 0666)
}

func main() {
	port := flag.String("port", "8888", "port to serve on")
	logPath = flag.String("logPath", "./clientLog", "the upload log dir")
	updateDir = flag.String("updateDir", "./hotupdate", "the updateDir of static file to host")
	flag.Parse()

	if !pathExists(*logPath) {
		createDir(*logPath)
	}

	if !pathExists(*updateDir) {
		createDir(*updateDir)
	}

	http.HandleFunc("/", index)
	http.HandleFunc("/uploadLog/", uploadLogHandler)
	http.HandleFunc("/uploadZIP/", uploadZIPHandler)
	http.Handle("/clientLog/", http.StripPrefix("/clientLog/", http.FileServer(http.Dir(*logPath))))
	http.Handle("/hotupdate/", http.StripPrefix("/hotupdate/", http.FileServer(http.Dir(*updateDir))))
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", "index")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func uploadLogHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		var clientLog ClientLog
		if err := decoder.Decode(&clientLog); err != nil {
			panic(err)
		}
		clientLog.save()
	}
}

func uploadZIPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Println("uploadZIPHandler GET")
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Fprintf(w, "upload error %v", err)
			return
		}
		defer file.Close()
		filename := handler.Filename
		f, _ := os.OpenFile(*updateDir+"/"+filename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		_, err = io.Copy(f, file)
		if err != nil {
			fmt.Fprintf(w, "copy error %v", err)
			return
		}

		os.RemoveAll(*updateDir + "/target")
		unzip(*updateDir+"/"+filename, *updateDir+"/")
		os.Rename(*updateDir+"/"+filename[:len(filename)-19], *updateDir+"/target")
		fmt.Fprintf(w, "upload success")
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func createDir(path string) {
	os.MkdirAll(path, os.ModePerm)
}

func unzip(zipfile, dest string) error {
	reader, err := zip.OpenReader(zipfile)
	if err != nil {
		return err
	}
	defer reader.Close()
	unzip2dest := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		filename := dest + f.Name
		if err = os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
			return err
		}
		w, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer w.Close()
		_, err = io.Copy(w, rc)
		return err
	}
	for _, file := range reader.File {
		err := unzip2dest(file)
		if err != nil {
			return err
		}
	}
	return nil
}
