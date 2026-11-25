package internal

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

func releasesAt(entries ...string) []Release {
	out := make([]Release, 0, len(entries))
	for _, e := range entries {
		day, title, _ := strings.Cut(e, " ")
		at, err := time.Parse("2006-01-02T15:04", day)
		if err != nil {
			panic(err)
		}
		out = append(out, Release{ID: title, Title: title, PublishedAt: at})
	}
	return out
}

func titles(releases []Release) string {
	out := make([]string, 0, len(releases))
	for _, r := range releases {
		out = append(out, r.Title)
	}
	return strings.Join(out, ", ")
}

func TestFilterIgnoresBackports(t *testing.T) {
	prerelease := regexp.MustCompile(`(?i)[0-9][-._+]?rc`)
	cases := []struct {
		name string
		in   []Release
		want string
	}{
		{
			"argo cd: patches to 3.4 and 3.5 after 3.6 is out",
			releasesAt(
				"2026-11-24T18:20 v3.6.3",
				"2026-11-24T18:20 v3.5.9",
				"2026-11-24T17:20 v3.4.15",
				"2026-11-18T10:00 v3.6.2",
				"2026-11-18T10:00 v3.5.8",
				"2026-11-18T09:00 v3.4.14",
				"2026-11-10T10:00 v3.6.1",
			),
			"v3.6.3, v3.6.2, v3.6.1",
		},
		{
			"k3s: three older lines released with the newest one",
			releasesAt(
				"2026-11-04T10:03 v1.37.2+k3s1",
				"2026-11-04T10:02 v1.36.6+k3s1",
				"2026-11-04T10:01 v1.35.10+k3s1",
				"2026-11-04T10:00 v1.34.13+k3s1",
				"2026-10-14T10:00 v1.37.1+k3s1",
			),
			"v1.37.2+k3s1, v1.37.1+k3s1",
		},
		{
			"valkey: older major lines",
			releasesAt(
				"2026-11-19T09:00 9.2.0",
				"2026-11-18T09:00 9.0.8",
				"2026-11-18T09:00 8.1.12",
				"2026-11-18T09:00 7.2.16",
				"2026-10-30T09:00 9.1.0",
			),
			"9.2.0, 9.1.0",
		},
		{
			"history stays: each line was the newest when it came out",
			releasesAt(
				"2026-11-20T10:00 v2.4.0",
				"2026-11-12T10:00 v2.3.0",
				"2026-09-25T10:00 v2.2.0",
			),
			"v2.4.0, v2.3.0, v2.2.0",
		},
		{
			"a prerelease of the next line does not make the stable line a backport",
			releasesAt(
				"2026-11-24T18:20 v3.6.3",
				"2026-11-20T10:00 v3.7.0-rc1",
				"2026-11-18T10:00 v3.6.2",
			),
			"v3.6.3, v3.6.2",
		},
		{
			"products of one feed are compared apart",
			releasesAt(
				"2026-11-23T15:21 UniFi InnerSpace MCP v0.1.0",
				"2026-11-20T20:36 unifi-mcp-shared v0.6.45",
				"2026-11-15T10:00 UniFi Network MCP v0.9.0",
			),
			"UniFi InnerSpace MCP v0.1.0, unifi-mcp-shared v0.6.45, UniFi Network MCP v0.9.0",
		},
		{
			"titles without a version stay",
			releasesAt(
				"2026-11-20T10:00 Winter release",
				"2026-11-10T10:00 v2.0.0",
			),
			"Winter release, v2.0.0",
		},
	}
	for _, c := range cases {
		f := Feed{IgnoreBackports: true, ignoreTitles: []*regexp.Regexp{prerelease}}
		if got := titles(f.Filter(c.in)); got != c.want {
			t.Errorf("%s:\n got: %s\nwant: %s", c.name, got, c.want)
		}
	}
}

func TestFilterKeepsBackportsByDefault(t *testing.T) {
	in := releasesAt(
		"2026-11-24T18:20 v3.6.3",
		"2026-11-24T18:20 v3.5.9",
		"2026-11-18T10:00 v3.6.2",
	)
	if got := titles(Feed{}.Filter(in)); got != "v3.6.3, v3.5.9, v3.6.2" {
		t.Errorf("got: %s", got)
	}
}
