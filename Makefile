build: ego.go files/files.go

ego.go: template.ego
	ego -package main -o ego.go

files/files.go: ego.go static/*
	@go get
	staticfiles --build-tags="!dev" -o files/files.go static/
