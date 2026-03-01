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

	if len(cfg.ExcludeSections) > 0 {
		content = excludeSectionsHTML(content, cfg.ExcludeSections)
	}

	for _, pattern := range cfg.RegexRemove {
		if re, err := regexp.Compile(pattern); err == nil {
			content = re.ReplaceAllString(content, "")
		}
	}

	r.Content = strings.TrimSpace(multiNLRe.ReplaceAllString(content, "\n\n"))
	return r
}

func excludeSectionsHTML(content string, names []string) string {
	parts := hTagRe.Split(content, -1)
	matches := hTagRe.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return content
	}

	var result strings.Builder
	if matches[0][0] > 0 {
		result.WriteString(content[:matches[0][0]])
	}

	for i, m := range matches {
		tag := content[m[2]:m[3]]
		heading := content[m[4]:m[5]]
		if !matchesAny(heading, names) {
			result.WriteString("<" + tag + ">" + heading + "</" + tag + ">")
			if i+1 < len(parts) {
				result.WriteString(parts[i+1])
			}
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