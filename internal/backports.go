package internal

import (
	"regexp"
	"strconv"
	"strings"
)

var versionRe = regexp.MustCompile(`v?(\d+)\.(\d+)`)

type versionLine struct {
	product string
	major   int
	minor   int
}

func (a versionLine) olderThan(b versionLine) bool {
	return a.major < b.major || (a.major == b.major && a.minor < b.minor)
}

func parseLine(title string) (versionLine, bool) {
	m := versionRe.FindStringSubmatchIndex(title)
	if m == nil {
		return versionLine{}, false
	}
	major, err := strconv.Atoi(title[m[2]:m[3]])
	if err != nil {
		return versionLine{}, false
	}
	minor, err := strconv.Atoi(title[m[4]:m[5]])
	if err != nil {
		return versionLine{}, false
	}
	product := strings.ToLower(strings.TrimSpace(title[:m[0]]))
	return versionLine{product, major, minor}, true
}

func backports(releases []Release) map[string]bool {
	lines := make([]versionLine, len(releases))
	parsed := make([]bool, len(releases))
	for i, r := range releases {
		lines[i], parsed[i] = parseLine(r.Title)
	}
	found := map[string]bool{}
	for i, r := range releases {
		if !parsed[i] {
			continue
		}
		for j, other := range releases {
			if !parsed[j] || lines[j].product != lines[i].product {
				continue
			}
			if lines[i].olderThan(lines[j]) && other.PublishedAt.Before(r.PublishedAt) {
				found[r.ID] = true
				break
			}
		}
	}
	return found
}

func (f Feed) Filter(releases []Release) []Release {
	kept := make([]Release, 0, len(releases))
	for _, r := range releases {
		if !f.IgnoreTitle(r.Title) {
			kept = append(kept, r)
		}
	}
	if !f.IgnoreBackports {
		return kept
	}
	old := backports(kept)
	current := kept[:0]
	for _, r := range kept {
		if !old[r.ID] {
			current = append(current, r)
		}
	}
	return current
}
