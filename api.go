package main

import (
	"fmt"
	"github.com/atotto/clipboard"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func fileUpload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20) //maxMemory

	file, handler, err := r.FormFile("upfile")

	if err != nil {

		log.Println(err)
		w.WriteHeader(500)
		fmt.Fprintf(w, "SERVER ERROR!"+err.Error())

		return

	}

	defer file.Close()

	//os.Mkdir("./files/", 0777)
	f, err := os.OpenFile(filepath.Join(TEMP_FILES_DIR, handler.Filename), os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {

		log.Println(err)
		w.WriteHeader(500)
		fmt.Fprintf(w, "SERVER ERROR!"+err.Error())

		return

	}

	io.Copy(f, file)

	fmt.Fprintf(w, "%s", "/files/"+handler.Filename)
}

func getClipboardData(w http.ResponseWriter, r *http.Request) {
	// Init returns an error if the package is not ready for use.
	//err := clipboard.Init()
	//if err != nil {
	//	panic(err)
	//}

	data, _ := clipboard.ReadAll()
	fmt.Fprintf(w, "%s", data)

	/*
		// write/read image format data of the clipboard, and
		// the byte buffer regarding the image are PNG encoded.
		clipboard.Write(clipboard.FmtImage, []byte("image data"))
		clipboard.Read(clipboard.FmtImage)
	*/
}

func getLocalIP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, getLocalIPInternal())
}

func getLocalIPInternal() string {
	//获取ip
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		log.Println(err)
		return "127.0.0.1"
	}
	for _, address := range addrs {

		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				log.Println(ipnet.IP.String())
				return ipnet.IP.String()
			}

		}
	}
	return "127.0.0.1"
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	switch r.URL.Path {

	case "/api/killself":
		fmt.Fprintf(w, "ok")
		if ui != nil {
			ui.Close()
		}

	case "/api/upload-file":
		fileUpload(w, r)
	case "/api/get-clipboard-data":
		getClipboardData(w, r)
	case "/api/get-local-ip":
		getLocalIP(w, r)
	case "/api/window-close":
		//bind browser window close event;
		go http.Get("http://127.0.0.1:8086/py/api/killself")

		fmt.Fprintf(w, "ok.")
		ui.Close()
	case "/api/get-available-pages":
		if os.Getenv("IN_DOCKER") == "" {
			//local
			fmt.Println(w, "")
		} else {
			//remote server

		}

	case "/api/search-ffmpeg":
		var files []string

		root := "/Users/zhouzhipeng"

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

			if err != nil {

				fmt.Println(err)
				return nil
			}

			if !info.IsDir() && filepath.Ext(path) == ".txt" {
				files = append(files, path)
			}

			return nil
		})

		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			fmt.Println(file)
		}
	case "/api/run-sql":
		sql := r.FormValue("sql")
		log.Println(sql)

		lines := strings.Split(strings.ReplaceAll(sql, "\r", ""), "\n")
		log.Println(lines)
		ds := ""
		rawSql := ""
		result := ""
		for _, l := range lines {
			if strings.Contains(l, "@ds") {
				if ds == "" {
					//first ds
					ds = strings.TrimSpace(strings.Split(l, "=")[1])
					rawSql = ""
					log.Println("ds >>>  " + ds)
				} else { //another ds

					//execute last sql
					result += querySql(rawSql, ds) + "<br/>"

					ds = strings.TrimSpace(strings.Split(l, "=")[1])
					rawSql = ""
					log.Println("ds >>>  " + ds)
					//then begin a new ds recording
				}
			} else {
				//sql
				//connect with whitespace
				trimedSql := strings.TrimSpace(l)
				if !strings.HasPrefix(trimedSql, "--") {
					rawSql += trimedSql + " "
				}

			}
		}

		rawSql = strings.TrimSpace(rawSql)

		//last
		if rawSql != "" && ds != "" {
			//execute last sql
			result += querySql(rawSql, ds) + "<br/>"
		} else {
			if ds == "" {
				result += " Error: No @ds selected!"
			} else if rawSql == "" {
				result += " Error: No sql selected!"
			}
		}

		fmt.Fprintf(w, result)

	default:
		w.WriteHeader(404)
	}
}
