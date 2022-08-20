run:
	go run main.go

watch:
	find . -name '*.go' | entr -rc go run main.go
