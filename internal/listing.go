package internal

import (
	"fmt"
	"html/template"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/chrishrb/go-grip/defaults"
)

// listingRow is one table row of a folder page.
type listingRow struct {
	Name     string
	Href     string
	IsDir    bool
	Size     string
	DateTime string
	Exact    string
	Relative string
}

type listingData struct {
	Path   string
	Up     string
	Rows   []listingRow
	Readme template.HTML
}

// listingTemplate is parsed with html/template, unlike layout.html: file
// names are untrusted, so every name and href is escaped by the template.
var listingTemplate = template.Must(template.ParseFS(defaults.Templates, "templates/listing.html"))

// renderListing renders the folder page fragment for dirPath (a URL path
// ending in "/") from entries. A non-empty readme is placed below the table.
func renderListing(dirPath string, entries []Entry, readme template.HTML, now time.Time) (string, error) {
	data := listingData{
		Path:   dirPath,
		Up:     upHref(dirPath),
		Rows:   make([]listingRow, 0, len(entries)),
		Readme: readme,
	}
	for _, e := range entries {
		row := listingRow{
			Name:     e.Name,
			Href:     entryHref(e.Name, e.IsDir),
			IsDir:    e.IsDir,
			DateTime: e.ModTime.UTC().Format(time.RFC3339),
			Exact:    e.ModTime.UTC().Format("2006-01-02 15:04:05 UTC"),
			Relative: formatRelativeTime(e.ModTime, now),
		}
		if !e.IsDir {
			row.Size = formatSize(e.Size)
		}
		data.Rows = append(data.Rows, row)
	}

	var b strings.Builder
	if err := listingTemplate.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}

// entryHref is the relative link to a folder entry. The name is escaped as
// one path segment; ":" is escaped too, so a name such as "a:b" is not read
// as a URL scheme.
func entryHref(name string, isDir bool) string {
	href := strings.ReplaceAll(url.PathEscape(name), ":", "%3A")
	if isDir {
		href += "/"
	}
	return href
}

// upHref is the link to the parent folder, or "" at the root. It carries the
// current folder's name in ?from= so the parent page can focus its row.
func upHref(dirPath string) string {
	trimmed := strings.Trim(dirPath, "/")
	if trimmed == "" {
		return ""
	}
	return "../?from=" + url.QueryEscape(path.Base(trimmed))
}

// backHref is the absolute link from the file at urlPath to its containing
// folder. It carries the file name in ?from= so the folder page focuses the
// row of the file just left. Each folder segment is escaped like entryHref.
func backHref(urlPath string) string {
	dir, name := path.Split(path.Clean("/" + urlPath))
	var b strings.Builder
	for _, seg := range strings.Split(strings.Trim(dir, "/"), "/") {
		if seg == "" {
			continue
		}
		b.WriteString("/")
		b.WriteString(entryHref(seg, false))
	}
	b.WriteString("/?from=")
	b.WriteString(url.QueryEscape(name))
	return b.String()
}

// backLabel is the folder containing the file at urlPath, written as the
// folder view's path heading shows it, e.g. "/docs/".
func backLabel(urlPath string) string {
	dir := path.Dir(path.Clean("/" + urlPath))
	if dir == "/" {
		return dir
	}
	return dir + "/"
}

// formatSize renders a byte count in base 1024 units.
func formatSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	value := float64(n) / unit
	i := 0
	for value >= unit && i < len(units)-1 {
		value /= unit
		i++
	}
	return fmt.Sprintf("%.1f %s", value, units[i])
}

// formatRelativeTime describes t relative to now, falling back to the date
// for anything older than 30 days.
func formatRelativeTime(t, now time.Time) string {
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return plural(int(d/time.Minute), "minute")
	case d < 24*time.Hour:
		return plural(int(d/time.Hour), "hour")
	case d <= 30*24*time.Hour:
		return plural(int(d/(24*time.Hour)), "day")
	default:
		return t.UTC().Format("2006-01-02")
	}
}

func plural(n int, unit string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s ago", unit)
	}
	return fmt.Sprintf("%d %ss ago", n, unit)
}
