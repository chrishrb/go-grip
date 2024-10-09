package pkg

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/aarol/reload"
)

type htmlStruct struct {
	Content  string
	Darkmode bool
}

func (client *Client) Serve(file string) error {
	reload := reload.New("./")
	reload.Log = log.New(io.Discard, "", 0)

	dir := http.Dir("./")
	chttp := http.NewServeMux()
	chttp.Handle("/", http.FileServer(dir))

	// Regex for markdown
	regex := regexp.MustCompile(`(?i)\.md$`)

	// Serve website with rendered markdown
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		f, err := dir.Open(r.URL.Path)
    if err == nil {
	    defer f.Close()
    }

		if err == nil && regex.MatchString(r.URL.Path) {
			// Open file and convert to html
			bytes, err := readToString(dir, r.URL.Path)
			htmlContent := client.MdToHTML(bytes)

			// Serve
			err = serveTemplate(w, htmlStruct{Content: string(htmlContent), Darkmode: client.Dark})
			if err != nil {
				log.Fatal(err)
			}
		} else {
			chttp.ServeHTTP(w, r)
		}
	})

	addr := fmt.Sprintf("http://localhost:%d", client.Port)
	if file == "" {
		// If README.md exists then open README.md at beginning
		readme := "README.md"
		f, err := dir.Open(readme)
    if err == nil {
	    defer f.Close()
    }
		if err == nil && client.OpenReadme {
			addr = path.Join(addr, readme)
		}
	} else {
		addr = path.Join(addr, file)
	}

	fmt.Printf("🚀 Starting server: %s\n", addr)

	if client.OpenBrowser {
		err := Open(addr)
		if err != nil {
			log.Println("Error:", err)
		}
	}

	handler := reload.Handle(http.DefaultServeMux)
	err := http.ListenAndServe(fmt.Sprintf(":%d", client.Port), handler)
	return err
}

func readToString(dir http.Dir, filename string) ([]byte, error) {
	f, err := dir.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(f)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
