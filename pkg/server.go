package pkg

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
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

	fmt.Printf("Starting server: %s\n", addr)

	if s.browser {
		err := Open(addr)
		if err != nil {
			fmt.Println("Error opening browser:", err)
		}
	}

	handler := reload.Handle(http.DefaultServeMux)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), handler)
}

func (s *Server) GenerateStaticSite(file string, outputDir string) error {
	fmt.Println("Warning: GenerateStaticSite is deprecated. Use GenerateSingleFile or GenerateDirectoryFiles instead.")

	absFilePath, err := filepath.Abs(file)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	if err := os.MkdirAll(absOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	staticDir := path.Join(absOutputDir, "static")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return fmt.Errorf("failed to create static directory: %v", err)
	}

	if err := copyStaticFiles(staticDir); err != nil {
		return fmt.Errorf("failed to copy static files: %v", err)
	}

	directory := path.Dir(absFilePath)
	if file == "" {
		directory = "."
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	var indexFile string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			content, err := os.ReadFile(path.Join(directory, entry.Name()))
			if err != nil {
				return fmt.Errorf("failed to read file %s: %v", entry.Name(), err)
			}

			htmlContent := s.parser.MdToHTML(content)

			htmlFile := strings.TrimSuffix(entry.Name(), ".md") + ".html"
			if entry.Name() == "README.md" {
				htmlFile = "index.html"
				indexFile = htmlFile
			}

			html := htmlStruct{
				Content:      string(htmlContent),
				Theme:        s.theme,
				BoundingBox:  s.boundingBox,
				CssCodeLight: getCssCode("github"),
				CssCodeDark:  getCssCode("github-dark"),
			}

			outputFilePath := path.Join(absOutputDir, htmlFile)
			if err := writeHTMLFile(outputFilePath, html); err != nil {
				return fmt.Errorf("failed to write HTML file %s: %v", htmlFile, err)
			}

			fmt.Printf("Generated HTML file: %s\n", outputFilePath)
		}
	}

	fmt.Printf("Output directory: %s\n", absOutputDir)

	if s.browser {
		indexPath := path.Join(absOutputDir, indexFile)
		if indexFile == "" {
			indexPath = path.Join(absOutputDir, "index.html")
		}
		fileURL := "file://" + indexPath
		err := Open(fileURL)
		if err != nil {
			fmt.Println("Error opening browser:", err)
		}
	}

	return nil
}

func copyStaticFiles(staticDir string) error {
	dirs := []string{"css", "js", "images", "emojis"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(staticDir, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	err := fs.WalkDir(defaults.StaticFiles, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		content, err := defaults.StaticFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %v", path, err)
		}

		outputPath := filepath.Join(staticDir, strings.TrimPrefix(path, "static/"))
		if err := os.WriteFile(outputPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %v", outputPath, err)
		}

		return nil
	})

	return err
}

func writeHTMLFile(path string, html htmlStruct) error {
	tmpl, err := template.ParseFS(defaults.Templates, "templates/layout.html")
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, html); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	return nil
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

func (s *Server) GenerateSingleFile(filePath string, outputDir string) error {
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	if err := os.MkdirAll(absOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	staticDir := path.Join(absOutputDir, "static")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return fmt.Errorf("failed to create static directory: %v", err)
	}

	if err := copyStaticFiles(staticDir); err != nil {
		return fmt.Errorf("failed to copy static files: %v", err)
	}

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	content, err := os.ReadFile(absFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", absFilePath, err)
	}

	htmlContent := s.parser.MdToHTML(content)

	baseFileName := filepath.Base(absFilePath)
	htmlFile := strings.TrimSuffix(baseFileName, ".md") + ".html"

	if baseFileName == "README.md" {
		htmlFile = "index.html"
	}

	outputFilePath := path.Join(absOutputDir, htmlFile)

	html := htmlStruct{
		Content:      string(htmlContent),
		Theme:        s.theme,
		BoundingBox:  s.boundingBox,
		CssCodeLight: getCssCode("github"),
		CssCodeDark:  getCssCode("github-dark"),
	}

	if err := writeHTMLFile(outputFilePath, html); err != nil {
		return fmt.Errorf("failed to write HTML file %s: %v", htmlFile, err)
	}

	fmt.Printf("Generated HTML file: %s\n", outputFilePath)

	if s.browser {
		fileURL := "file://" + outputFilePath
		err := Open(fileURL)
		if err != nil {
			fmt.Println("Error opening browser:", err)
		}
	}

	return nil
}

func (s *Server) GenerateDirectoryFiles(dirPath string, outputDir string) error {
	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	if err := os.MkdirAll(absOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	staticDir := path.Join(absOutputDir, "static")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return fmt.Errorf("failed to create static directory: %v", err)
	}

	if err := copyStaticFiles(staticDir); err != nil {
		return fmt.Errorf("failed to copy static files: %v", err)
	}

	entries, err := os.ReadDir(absDirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	foundMarkdown := false

	var indexFile string
	generatedFiles := make(map[string]string) // filename -> title

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			foundMarkdown = true

			mdFilePath := path.Join(absDirPath, entry.Name())
			content, err := os.ReadFile(mdFilePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %v", mdFilePath, err)
			}

			title := extractTitle(content, entry.Name())

			htmlContent := s.parser.MdToHTML(content)

			htmlFile := strings.TrimSuffix(entry.Name(), ".md") + ".html"

			if entry.Name() == "README.md" {
				htmlFile = "index.html"
				indexFile = htmlFile
			}

			outputFilePath := path.Join(absOutputDir, htmlFile)

			generatedFiles[htmlFile] = title

			html := htmlStruct{
				Content:      string(htmlContent),
				Theme:        s.theme,
				BoundingBox:  s.boundingBox,
				CssCodeLight: getCssCode("github"),
				CssCodeDark:  getCssCode("github-dark"),
			}

			if err := writeHTMLFile(outputFilePath, html); err != nil {
				return fmt.Errorf("failed to write HTML file %s: %v", outputFilePath, err)
			}

			fmt.Printf("Generated HTML file: %s\n", outputFilePath)
		}
	}

	if !foundMarkdown {
		return fmt.Errorf("no markdown files found in directory %s", absDirPath)
	}

	if indexFile == "" {
		dirName := filepath.Base(absDirPath)
		indexContent := generateDirectoryIndex(dirName, generatedFiles)

		html := htmlStruct{
			Content:      string(indexContent),
			Theme:        s.theme,
			BoundingBox:  s.boundingBox,
			CssCodeLight: getCssCode("github"),
			CssCodeDark:  getCssCode("github-dark"),
		}

		indexFile = "index.html"
		indexPath := path.Join(absOutputDir, indexFile)

		if err := writeHTMLFile(indexPath, html); err != nil {
			return fmt.Errorf("failed to write index file: %v", err)
		}

		fmt.Printf("Generated index file: %s\n", indexPath)
	}

	fmt.Printf("Output directory: %s\n", absOutputDir)

	if s.browser {
		fileURL := "file://" + path.Join(absOutputDir, indexFile)
		err := Open(fileURL)
		if err != nil {
			fmt.Println("Error opening browser:", err)
		}
	}

	return nil
}

func extractTitle(content []byte, filename string) string {
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "# ") {
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "# "))
		}
	}

	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

func generateDirectoryIndex(dirName string, files map[string]string) string {
	var sb strings.Builder

	sb.WriteString("<h1>Directory: " + dirName + "</h1>\n")
	sb.WriteString("<p>The following files were generated:</p>\n")
	sb.WriteString("<ul>\n")

	var filenames []string
	for filename := range files {
		if filename != "index.html" {
			filenames = append(filenames, filename)
		}
	}
	sort.Strings(filenames)

	for _, filename := range filenames {
		title := files[filename]
		sb.WriteString(fmt.Sprintf("  <li><a href=\"%s\">%s</a></li>\n", filename, title))
	}

	sb.WriteString("</ul>\n")
	return sb.String()
}
