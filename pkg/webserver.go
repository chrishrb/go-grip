package pkg

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"text/template"

	"github.com/aarol/reload"
	"github.com/chrishrb/go-grip/defaults"
)

type htmlStruct struct {
	Content     string
	Theme       string
	BoundingBox bool
}

func (client *Client) Serve(file string) error {
	directory := path.Dir(file)
	filename := path.Base(file)

	reload := reload.New(directory)
	reload.Log = log.New(io.Discard, "", 0)

	validThemes := map[string]bool{"light": true, "dark": true, "auto": true}

	if !validThemes[client.Theme] {
		log.Println("Warning: Unknown theme ", client.Theme, ", defaulting to 'auto'")
		client.Theme = "auto"
	}

	dir := http.Dir(directory)
	chttp := http.NewServeMux()
	chttp.Handle("/static/", http.FileServer(http.FS(defaults.StaticFiles)))
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
			if err != nil {
				log.Fatal(err)
				return
			}
			htmlContent := client.MdToHTML(bytes)

			// Serve
			err = serveTemplate(w, htmlStruct{Content: string(htmlContent), Theme: client.Theme, BoundingBox: client.BoundingBox})
			if err != nil {
				log.Fatal(err)
				return
			}
		} else {
			chttp.ServeHTTP(w, r)
		}
	})

	addr := fmt.Sprintf("http://%s:%d/", client.Host, client.Port)
	if file == "" {
		// If README.md exists then open README.md at beginning
		readme := "README.md"
		f, err := dir.Open(readme)
		if err == nil {
			defer f.Close()
		}
		if err == nil && client.OpenReadme {
			addr, _ = url.JoinPath(addr, readme)
		}
	} else {
		addr, _ = url.JoinPath(addr, filename)
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
	tmpl, err := template.ParseFS(defaults.Templates, "templates/layout.html")
	if err != nil {
		return err
	}
	err = tmpl.Execute(w, html)
	return err
}
