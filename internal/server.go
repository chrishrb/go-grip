package internal

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
	parser       *Parser
	boundingBox  bool
	host         string
	port         int
	browser      bool
	enableReload bool
}

func NewServer(host string, port int, boundingBox bool, browser bool, enableReload bool, parser *Parser) *Server {
	return &Server{
		host:         host,
		port:         port,
		boundingBox:  boundingBox,
		browser:      browser,
		enableReload: enableReload,
		parser:       parser,
	}
}

func (s *Server) Serve(file string) error {
	directory := path.Dir(file)
	filename := path.Base(file)

	var reloadMiddleware *reload.Reloader
	if s.enableReload {
		reloadMiddleware = reload.New(directory)
		reloadMiddleware.DebugLog = log.New(io.Discard, "", 0)
		// Fix WebSocket CORS issues for development
		reloadMiddleware.Upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}

	dir := http.Dir(directory)
	handler := s.newHandler(dir)

	addr := fmt.Sprintf("http://%s:%d/", s.host, s.port)
	if file == "" {
		// If README.md exists then open README.md at beginning
		readme := "README.md"
		f, err := dir.Open(readme)
		if err == nil {
			//nolint:errcheck
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

	if s.enableReload {
		handler = reloadMiddleware.Handle(handler)
		fmt.Printf("📡 Auto-reload enabled. Files will trigger browser refresh.\n")
	} else {
		fmt.Printf("🔄 Auto-reload disabled. Use F5 to manually refresh.\n")
	}
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), handler)
}

func (s *Server) newHandler(dir http.Dir) http.Handler {
	fileServer := http.FileServer(dir)
	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(http.FS(defaults.StaticFiles)))

	regex := regexp.MustCompile(`(?i)\.md$`)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if regex.MatchString(r.URL.Path) {
			isFile, err := isRegularFile(dir, r.URL.Path)
			if err == nil && isFile {
				setNoCacheHeaders(w)

				bytes, err := readToString(dir, r.URL.Path)
				if err != nil {
					log.Fatal(err)
					return
				}
				htmlContent, err := s.parser.MdToHTML(bytes)
				if err != nil {
					log.Fatal(err)
					return
				}

				err = serveTemplate(w, htmlStruct{
					Content:      string(htmlContent),
					BoundingBox:  s.boundingBox,
					CssCodeLight: getCssCode("github"),
					CssCodeDark:  getCssCode("github-dark"),
				})
				if err != nil {
					log.Fatal(err)
					return
				}
				return
			}
		}

		isDirectory, err := isDirectory(dir, r.URL.Path)
		if err == nil && isDirectory {
			setNoCacheHeaders(w)
			stripCacheValidators(r)
		}

		fileServer.ServeHTTP(w, r)
	})

	return mux
}

func readToString(dir http.Dir, filename string) ([]byte, error) {
	f, err := dir.Open(filename)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck
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

func setNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

func stripCacheValidators(r *http.Request) {
	r.Header.Del("If-Modified-Since")
	r.Header.Del("If-None-Match")
}

func isDirectory(dir http.Dir, name string) (bool, error) {
	file, err := dir.Open(name)
	if err != nil {
		return false, err
	}
	//nolint:errcheck
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}

func isRegularFile(dir http.Dir, name string) (bool, error) {
	file, err := dir.Open(name)
	if err != nil {
		return false, err
	}
	//nolint:errcheck
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return false, err
	}

	return !info.IsDir(), nil
}
