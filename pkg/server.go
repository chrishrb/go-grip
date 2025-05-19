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
	"strings"
	"text/template"

	"github.com/aarol/reload"
	chroma_html "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/chrishrb/go-grip/defaults"
)

type Server struct {
	parser      *Parser
	theme       string
	boundingBox bool
	host        string
	port        int
	browser     bool
}

func NewServer(host string, port int, theme string, boundingBox bool, browser bool, parser *Parser) *Server {
	return &Server{
		host:        host,
		port:        port,
		theme:       theme,
		boundingBox: boundingBox,
		browser:     browser,
		parser:      parser,
	}
}

func (s *Server) Serve(file string) error {
	directory := path.Dir(file)
	filename := path.Base(file)

	reload := reload.New(directory)
	reload.DebugLog = log.New(io.Discard, "", 0)

	validThemes := map[string]bool{"light": true, "dark": true, "auto": true}

	if !validThemes[s.theme] {
		log.Println("Warning: Unknown theme ", s.theme, ", defaulting to 'auto'")
		s.theme = "auto"
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
			htmlContent := s.parser.MdToHTML(bytes)

			// Serve
			err = serveTemplate(w, htmlStruct{
				Content:      string(htmlContent),
				Theme:        s.theme,
				BoundingBox:  s.boundingBox,
				CssCodeLight: getCssCode("github"),
				CssCodeDark:  getCssCode("github-dark"),
			})
			if err != nil {
				log.Fatal(err)
				return
			}
		} else {
			chttp.ServeHTTP(w, r)
		}
	})

	addr := fmt.Sprintf("http://%s:%d/", s.host, s.port)
	if file == "" {
		// If README.md exists then open README.md at beginning
		readme := "README.md"
		f, err := dir.Open(readme)
		if err == nil {
			defer f.Close()
		}
		if err == nil {
			addr, _ = url.JoinPath(addr, readme)
		}
	} else {
		addr, _ = url.JoinPath(addr, filename)
	}

	fmt.Printf("🚀 Starting server: %s\n", addr)

	if s.browser {
		err := Open(addr)
		if err != nil {
			fmt.Println("❌ Error opening browser:", err)
		}
	}

	handler := reload.Handle(http.DefaultServeMux)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), handler)
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

type htmlStruct struct {
	Content      string
	Theme        string
	BoundingBox  bool
	CssCodeLight string
	CssCodeDark  string
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

func getCssCode(style string) string {
	buf := new(strings.Builder)
	formatter := chroma_html.New(chroma_html.WithClasses(true))
	s := styles.Get(style)
	_ = formatter.WriteCSS(buf, s)
	return buf.String()
}
