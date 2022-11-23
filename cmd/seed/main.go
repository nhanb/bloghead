package main

import (
	"os"
	"path/filepath"
	"strings"

	"go.imnhan.com/bloghead/models"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	seeds, err := filepath.Glob("seed-data/*.seed")
	check(err)

	models.RegisterRegexFunc()
	models.SetDbFile("Site1.bloghead")

	for _, seed := range seeds {
		data, _ := os.ReadFile(seed)
		parts := strings.SplitN(string(data), "\n", 4)
		(&models.Post{
			IsDraft: parts[0] == "draft",
			Slug:    parts[1],
			Title:   parts[2],
			Content: parts[3],
		}).Create()
	}
}
