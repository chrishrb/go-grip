package pkg

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
)

type htmlStruct struct {
	Content  string
	Darkmode bool
}

func (client *Client) Serve(htmlContent []byte) error {
	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Serve website with rendered markdown
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := serveTemplate(w, htmlStruct{Content: string(htmlContent), Darkmode: client.Dark})
		if err != nil {
			log.Fatal(err)
		}
	})

	addr := fmt.Sprintf("http://localhost:%d", client.Port)
	fmt.Printf("Starting server: %s\n", addr)

	if client.OpenBrowser {
		Open(addr)
	}

	http.ListenAndServe(fmt.Sprintf(":%d", client.Port), nil)

	return nil
}

func serveTemplate(w http.ResponseWriter, html htmlStruct) error {
	w.Header().Set("Content-Type", "text/html")
	lp := filepath.Join("templates", "layout.html")
	tmpl, err := template.ParseFiles(lp)
	if err != nil {
		return err
	}
	err = tmpl.Execute(w, html)
	return err
}
