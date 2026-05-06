package main

import (
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TJN25/clilog"
)

func main() {
	src := flag.String("src", "", "path to source directory")
	out := flag.String("out", "docs", "path to output directory (default: docs)")
	index := flag.String("index", "meal-ideas.md", "file to target as the 'index.md' (default: meal-ideas.md)")

	flag.Parse()
	if *src == "" {
		clilog.Fprintln(os.Stderr, "--src is required")
		os.Exit(1)
	}

	clilog.Infof("Using %s to write to %s with %s as the landing page\n", *src, *out, *index)
	err := walkTarget(src, out, index)
	if err != nil {
		os.Exit(1)
	}
}

func walkTarget(src, out, index *string) error {
	files := []string{}
	err := filepath.WalkDir(*src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) != ".md" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			clilog.Errorf("%s\n", err)
			return nil
		}

		contents := strings.Split(string(data), "\n")

		modifiedContents, err := processData(contents)
		if err != nil {
			clilog.Errorf("%s\n", err)
			return nil
		}

		outPath, err := writeContents(modifiedContents, out, path)
		if err != nil {
			clilog.Errorf("%s\n", err)
			return nil
		}
		files = append(files, outPath)
		if filepath.Base(path) == *index {
			_, err = writeContents(modifiedContents, out, "index.md")
			if err != nil {
				clilog.Errorf("%s\n", err)
				return nil
			}
		}
		err = writeContentsPage(out, files)
		if err != nil {
			clilog.Errorf("%s\n", err)
			return nil
		}

		return nil
	})
	if err != nil {
		clilog.Errorf("%s\n", err)
		return err
	}
	return nil
}

func processData(contents []string) ([]string, error) {
	frontmatter := len(contents) > 0 && strings.HasPrefix(contents[0], "---")
	seenOpenFrontmatter := false
	seenCloseFrontmatter := false
	modifiedContents := []string{}
	headingSet := ""
	for _, line := range contents {
		if frontmatter && !seenCloseFrontmatter {
			if strings.HasPrefix(line, "---") {
				if !seenOpenFrontmatter {
					seenOpenFrontmatter = true
					continue
				}
				seenCloseFrontmatter = true
				continue
			}
			continue
		}
		line = rewriteWikiLinks(line)
		switch {
		case strings.HasPrefix(line, "### "):
			title := strings.TrimPrefix(line, "### ")
			modifiedContents = append(modifiedContents, `    ??? info "`+title+`"`)
			headingSet = "h3"
			continue
		case strings.HasPrefix(line, "## "):
			title := strings.TrimPrefix(line, "## ")
			modifiedContents = append(modifiedContents, `??? note "`+title+`"`)
			headingSet = "h2"
			continue
		}

		if line != "" {
			if headingSet == "h3" {
				line = "        " + line
			} else if headingSet == "h2" {
				line = "    " + line
			}
		}
		modifiedContents = append(modifiedContents, line)

	}

	return modifiedContents, nil
}

func writeContents(contents []string, out *string, path string) (string, error) {
	base := filepath.Base(path)
	outPath := filepath.Join(*out, base)
	err := os.MkdirAll(filepath.Dir(outPath), 0o755)
	if err != nil {
		return "", err
	}

	output := strings.Join(contents, "\n")
	err = os.WriteFile(outPath, []byte(output), 0o644)
	if err != nil {
		return "", err
	}
	return outPath, nil
}

var wikiLinkRE = regexp.MustCompile(`\[\[([^|\]]+)(?:\|([^\]]+))?\]\]`)

func rewriteWikiLinks(line string) string {
	return wikiLinkRE.ReplaceAllStringFunc(line, func(match string) string {
		parts := wikiLinkRE.FindStringSubmatch(match)

		target := parts[1]
		text := target
		if parts[2] != "" {
			text = parts[2]
		}
		return "[" + text + "](" + target + ".md)"
	})
}

func writeContentsPage(out *string, files []string) error {
	lines := []string{"# Contents\n"}

	for _, file := range files {
		name := strings.TrimPrefix(strings.TrimSuffix(file, filepath.Ext(file)), *out+"/")
		lines = append(lines, "- ["+name+"]("+name+".md)")
	}
	outPath := filepath.Join(*out, "contents.md")
	output := strings.Join(lines, "\n")
	return os.WriteFile(outPath, []byte(output), 0o644)
}
