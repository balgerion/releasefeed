package internal

import (
	"regexp"
	"strings"
)

var (
	multiNLRe = regexp.MustCompile(`\n{3,}`)
	headingRe = regexp.MustCompile(`(?is)<h([1-6])[^>]*>(.*?)</h[1-6]>|<p[^>]*>\s*<(?:strong|b)[^>]*>(.*?)</(?:strong|b)>\s*:?\s*</p>`)
)

func Normalize(r Release, cfg Feed) Release {
	content := r.Content

	if len(cfg.ExcludeSections) > 0 {
		content = excludeSectionsHTML(content, cfg.ExcludeSections)
	}

	for _, re := range cfg.regexRemove {
		content = re.ReplaceAllString(content, "")
	}

	r.Content = strings.TrimSpace(multiNLRe.ReplaceAllString(content, "\n\n"))
	return r
}

func excludeSectionsHTML(content string, names []string) string {
	matches := headingRe.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return content
	}

	var result strings.Builder
	result.WriteString(content[:matches[0][0]])

	skipLevel := 0
	for i, m := range matches {
		level, heading := 6, ""
		if m[2] >= 0 {
			level = int(content[m[2]] - '0')
			heading = content[m[4]:m[5]]
		} else {
			heading = content[m[6]:m[7]]
		}
		if skipLevel > 0 && level > skipLevel {
			continue
		}
		skipLevel = 0
		if matchesAny(heading, names) {
			skipLevel = level
			continue
		}
		end := len(content)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		result.WriteString(content[m[0]:end])
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