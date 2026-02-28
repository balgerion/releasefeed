package internal

import (
	"html"
	"regexp"
	"strings"
)

var (
	htmlTagRe = regexp.MustCompile(`<[^>]+>`)
	multiNLRe = regexp.MustCompile(`\n{3,}`)
	mdLinkRe  = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	mdBoldRe  = regexp.MustCompile(`\*{1,3}([^*]+)\*{1,3}`)
	mdCodeRe  = regexp.MustCompile("`([^`]+)`")
	h1Re      = regexp.MustCompile(`(?i)<h1[^>]*>`)
	h2Re      = regexp.MustCompile(`(?i)<h2[^>]*>`)
	h3Re      = regexp.MustCompile(`(?i)<h3[^>]*>`)
	hCloseRe  = regexp.MustCompile(`(?i)</h[1-6]>`)
	brRe      = regexp.MustCompile(`(?i)<br\s*/?>`)
)

func Normalize(r Release, cfg Feed) Release {
	if cfg.Raw {
		return r
	}

	content := html.UnescapeString(r.Content)

	content = h1Re.ReplaceAllString(content, "\n# ")
	content = h2Re.ReplaceAllString(content, "\n## ")
	content = h3Re.ReplaceAllString(content, "\n### ")
	content = hCloseRe.ReplaceAllString(content, "\n")
	content = brRe.ReplaceAllString(content, "\n")
	content = htmlTagRe.ReplaceAllString(content, "")

	if len(cfg.Sections.Names) > 0 {
		switch cfg.Sections.Mode {
		case "exclude":
			content = excludeSections(content, cfg.Sections.Names)
		default:
			if extracted := includeSections(content, cfg.Sections.Names); extracted != "" {
				content = extracted
			}
		}
	}

	for _, pattern := range cfg.RegexRemove {
		re, err := regexp.Compile(pattern)
		if err == nil {
			content = re.ReplaceAllString(content, "")
		}
	}

	content = mdLinkRe.ReplaceAllString(content, "$1")
	content = mdBoldRe.ReplaceAllString(content, "$1")
	content = mdCodeRe.ReplaceAllString(content, "$1")
	content = multiNLRe.ReplaceAllString(content, "\n\n")
	content = strings.TrimSpace(content)

	r.Content = content
	return r
}

func includeSections(content string, names []string) string {
	lines := strings.Split(content, "\n")
	var result []string
	var current []string
	inSection := false

	flush := func() {
		if len(current) > 0 {
			result = append(result, strings.Join(current, "\n"))
			current = nil
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			if inSection {
				flush()
			}
			inSection = matchesAny(trimmed, names)
			if inSection {
				current = append(current, line)
			}
		} else if inSection {
			current = append(current, line)
		}
	}
	flush()

	return strings.Join(result, "\n\n")
}

func excludeSections(content string, names []string) string {
	lines := strings.Split(content, "\n")
	var result []string
	skip := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			skip = matchesAny(trimmed, names)
		}
		if !skip {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func matchesAny(line string, names []string) bool {
	lower := strings.ToLower(line)
	for _, name := range names {
		if strings.Contains(lower, strings.ToLower(name)) {
			return true
		}
	}
	return false
}