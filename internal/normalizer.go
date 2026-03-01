package internal

import (
	"html"
	"regexp"
	"strings"
)

var (
	multiNLRe = regexp.MustCompile(`\n{3,}`)
	hTagRe    = regexp.MustCompile(`(?i)<(h[1-6])[^>]*>(.*?)</h[1-6]>`)
)

func Normalize(r Release, cfg Feed) Release {
	content := html.UnescapeString(r.Content)

	if len(cfg.Sections.Names) > 0 {
		switch cfg.Sections.Mode {
		case "include":
			content = filterSectionsHTML(content, cfg.Sections.Names, true)
		case "exclude":
			content = filterSectionsHTML(content, cfg.Sections.Names, false)
		}
	}

	for _, pattern := range cfg.RegexRemove {
		re, err := regexp.Compile(pattern)
		if err == nil {
			content = re.ReplaceAllString(content, "")
		}
	}

	content = strings.TrimSpace(multiNLRe.ReplaceAllString(content, "\n\n"))
	r.Content = content
	return r
}

func filterSectionsHTML(content string, names []string, include bool) string {
	parts := hTagRe.Split(content, -1)
	matches := hTagRe.FindAllStringSubmatchIndex(content, -1)

	if len(matches) == 0 {
		return content
	}

	type section struct {
		tag     string
		heading string
		body    string
	}

	sections := []section{}
	preamble := ""

	if len(matches) > 0 && matches[0][0] > 0 {
		preamble = content[:matches[0][0]]
	}

	for i, m := range matches {
		tag := content[m[2]:m[3]]
		heading := content[m[4]:m[5]]
		body := ""
		if i+1 < len(parts) {
			body = parts[i+1]
		}
		sections = append(sections, section{tag, heading, body})
	}

	var result strings.Builder
	result.WriteString(preamble)

	for _, s := range sections {
		matched := matchesAny(s.heading, names)
		if (include && matched) || (!include && !matched) {
			result.WriteString("<" + s.tag + ">" + s.heading + "</" + s.tag + ">")
			result.WriteString(s.body)
		}
	}

	return result.String()
}

func matchesAny(text string, names []string) bool {
	lower := strings.ToLower(text)
	for _, name := range names {
		if strings.Contains(lower, strings.ToLower(name)) {
			return true
		}
	}
	return false
}