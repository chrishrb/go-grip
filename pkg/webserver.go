package pkg

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"text/template"
)

type htmlStruct struct {
	Content  string
	Darkmode bool
}

func (client *Client) Serve(file string) error {
	dir := http.Dir("./")
	chttp := http.NewServeMux()
	chttp.Handle("/", http.FileServer(dir))

	// Regex for markdown
	regex := regexp.MustCompile(`(?i)\.md$`)

	// Serve website with rendered markdown
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if regex.MatchString(r.URL.Path) {
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

	addr := fmt.Sprintf("http://localhost:%d/", client.Port)
	fmt.Printf("Starting server: %s\n", addr)

	if file != "" {
		addr = path.Join(addr, file)
	}

	if client.OpenBrowser {
		err := Open(addr)
		if err != nil {
			log.Println("Error:", err)
		}
	}

	err := http.ListenAndServe(fmt.Sprintf(":%d", client.Port), nil)
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
