package main

import (
	"embed"
	"fmt"
	"gogo/lorca"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
)

//go:embed www
var www embed.FS

//go:embed www/shell.html
var shellHtml string

func main() {

	// open ui window
	args := []string{}
	if runtime.GOOS == "linux" {
		args = append(args, "--class=gogo")
	}
	ui, err := lorca.New("data:text/html,"+shellHtml, "", 480, 320, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	//http & file server
	http.Handle("/", http.FileServer(http.FS(www)))
	ln, err := net.Listen("tcp", "0.0.0.0:8000")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	go http.Serve(ln, nil)

	//bind api
	http.HandleFunc("/api/window-close", func(w http.ResponseWriter, r *http.Request) {
		//bind browser window close event;
		fmt.Fprintf(w, "ok.")
		ui.Close()
	})

	// Wait until the interrupt signal arrives or browser window is closed
	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt)
	select {
	case <-sigc:
	case <-ui.Done():
	}

	log.Println("exiting...")
}
