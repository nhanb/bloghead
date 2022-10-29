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

	models.Init("Site1.bloghead")

	for _, seed := range seeds {
		data, _ := os.ReadFile(seed)
		parts := strings.SplitN(string(data), "\n", 3)
		(&models.Post{
			Slug:    parts[0],
			Title:   parts[1],
			Content: parts[2],
		}).Create()
	}
}
