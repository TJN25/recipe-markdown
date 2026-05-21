package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/TJN25/clilog"
)

func main() {
	src := flag.String("src", "", "path to source directory")
	out := flag.String("out", "docs", "path to output directory (default: docs)")
	index := flag.String("index", "meal-ideas.md", "file to target as the 'index.md' (default: meal-ideas.md)")
	mkdocs := flag.String("mkdocs", "mkdocs.yml", "path to MkDocs config to write (default: mkdocs.yml)")

	if err := clilog.InitializeLogger(5); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}

	flag.Parse()
	if *src == "" {
		clilog.Fprintln(os.Stderr, "--src is required")
		os.Exit(1)
	}

	clilog.Infof("Using %s to write to %s with %s as the landing page\n", *src, *out, *index)
	err := walkTarget(src, out, index, mkdocs)
	if err != nil {
		clilog.Errorf("%s\n", err)
		os.Exit(1)
	}
}

type Recipes []Recipe

type Recipe struct {
	Path     string
	Name     string
	Tags     []string
	Timing   Time
	Contents []string
}

type Time struct {
	Total  string
	Prep   string
	Active string
}

func walkTarget(src, out, index, mkdocs *string) error {
	files := []string{}
	recipes := Recipes{}
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
		recipe := parseRecipe(path, contents)
		if recipe.isIndexable() {
			recipes = append(recipes, recipe)
		}

		modifiedContents, err := processData(contents)
		if err != nil {
			clilog.Errorf("%s\n", err)
			return nil
		}

		outPath, err := writePage(modifiedContents, out, path)
		if err != nil {
			clilog.Errorf("%s\n", err)
			return nil
		}
		files = append(files, outPath)
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
	if err := writeContentsPage(out, files); err != nil {
		return err
	}
	if err := writeIndexPage(out, recipes); err != nil {
		return err
	}
	return writeMkdocsConfig(mkdocs, recipes)
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

func writePage(contents []string, out *string, path string) (string, error) {
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

		target := strings.TrimPrefix(parts[1], "recipes/")
		target = filepath.Base(target)
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

func parseRecipe(path string, contents []string) Recipe {
	recipe := Recipe{
		Path: filepath.Base(path),
		Name: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
	}

	inFrontmatter := false
	inTags := false
	inAliases := false
	inTiming := false
	for i, line := range contents {
		trimmed := strings.TrimSpace(line)
		if i == 0 && trimmed == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter && trimmed == "---" {
			break
		}
		if !inFrontmatter {
			if strings.HasPrefix(trimmed, "# ") {
				recipe.Name = strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
				break
			}
			continue
		}

		switch {
		case trimmed == "tags:":
			inTags = true
			inAliases = false
			inTiming = false
			continue
		case trimmed == "aliases:":
			inTags = false
			inAliases = true
			inTiming = false
			continue
		case trimmed == "timing:":
			inTags = false
			inAliases = false
			inTiming = true
			continue
		case !strings.HasPrefix(trimmed, "- ") && !strings.HasPrefix(line, " ") && strings.Contains(trimmed, ":"):
			inTags = false
			inAliases = false
			inTiming = false
		}

		if inTags && strings.HasPrefix(trimmed, "- ") {
			recipe.Tags = append(recipe.Tags, strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
			continue
		}
		if inAliases && strings.HasPrefix(trimmed, "- ") && recipe.Name == strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) {
			recipe.Name = strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")), `"`)
			continue
		}
		if inTiming && strings.HasPrefix(trimmed, "total:") {
			recipe.Timing.Total = strings.TrimSpace(strings.TrimPrefix(trimmed, "total:"))
			continue
		}
	}

	return recipe
}

func (r Recipe) isIndexable() bool {
	return r.hasTag("recipe/full-meal") || r.hasTag("recipe/component")
}

func (r Recipe) hasTag(tag string) bool {
	for _, candidate := range r.Tags {
		if candidate == tag {
			return true
		}
	}
	return false
}

func (r Recipe) timeLabel() string {
	total := strings.TrimSpace(r.Timing.Total)
	if total == "" || total == "0 minutes" || total == "0 minute" {
		return ""
	}
	return " (" + total + ")"
}

func writeIndexPage(out *string, recipes Recipes) error {
	type recipeGroup struct {
		Title string
		Tag   string
	}
	type recipeSection struct {
		Title  string
		Tag    string
		Groups []recipeGroup
	}

	sections := []recipeSection{
		{
			Title: "Full meals",
			Tag:   "recipe/full-meal",
			Groups: []recipeGroup{
				{Title: "Go to", Tag: "recipe/status/go-to"},
				{Title: "Reliable meals", Tag: "recipe/status/reliable"},
				{Title: "Have made at least once", Tag: "recipe/status/made-once"},
				{Title: "Ideas", Tag: "recipe/status/idea"},
			},
		},
		{
			Title: "Components",
			Tag:   "recipe/component",
			Groups: []recipeGroup{
				{Title: "Go to", Tag: "recipe/status/go-to"},
				{Title: "Reliable", Tag: "recipe/status/reliable"},
				{Title: "Have made at least once", Tag: "recipe/status/made-once"},
				{Title: "Ideas", Tag: "recipe/status/idea"},
			},
		},
	}

	sort.Slice(recipes, func(i, j int) bool {
		return strings.ToLower(recipes[i].Name) < strings.ToLower(recipes[j].Name)
	})

	lines := []string{"# Recipes", ""}
	for _, section := range sections {
		lines = append(lines, `??? note "`+section.Title+`"`)
		for _, group := range section.Groups {
			matches := recipes.withTags(section.Tag, group.Tag)
			if len(matches) == 0 {
				continue
			}
			lines = append(lines, `    ??? info "`+group.Title+`"`)
			for _, recipe := range matches {
				lines = append(lines, "        - ["+recipe.Name+"]("+recipe.Path+")"+recipe.timeLabel())
			}
			lines = append(lines, "")
		}
	}

	outPath := filepath.Join(*out, "index.md")
	output := strings.TrimRight(strings.Join(lines, "\n"), "\n") + "\n"
	return os.WriteFile(outPath, []byte(output), 0o644)
}

func (recipes Recipes) withTags(tags ...string) Recipes {
	matches := Recipes{}
	for _, recipe := range recipes {
		matched := true
		for _, tag := range tags {
			if !recipe.hasTag(tag) {
				matched = false
				break
			}
		}
		if matched {
			matches = append(matches, recipe)
		}
	}
	return matches
}

func writeMkdocsConfig(path *string, recipes Recipes) error {
	fullMealGoTos := recipes.withTags("recipe/full-meal", "recipe/status/go-to")
	componentGoTos := recipes.withTags("recipe/component", "recipe/status/go-to")

	lines := []string{
		"site_name: Recipes",
		"docs_dir: docs",
		"nav:",
		"  - Home: index.md",
		"  - Contents: contents.md",
	}
	appendNavGroup := func(title string, items Recipes) {
		if len(items) == 0 {
			return
		}
		lines = append(lines, "  - "+title+":")
		lines = append(lines, "      - Go to:")
		for _, recipe := range items {
			lines = append(lines, "          - "+yamlKey(recipe.Name)+": "+recipe.Path)
		}
	}
	appendNavGroup("Full meals", fullMealGoTos)
	appendNavGroup("Components", componentGoTos)

	lines = append(lines,
		"theme:",
		"  name: material",
		"  features:",
		"    - navigation.expand",
		"",
		"plugins:",
		"  - search",
		"",
		"markdown_extensions:",
		"  - admonition",
		"  - pymdownx.details",
	)

	return os.WriteFile(*path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

func yamlKey(value string) string {
	value = strings.ReplaceAll(value, `"`, `\"`)
	return `"` + value + `"`
}
