build: template.ego.go files/files.go

template.ego.go: template.ego
	ego

files/files.go: template.ego.go static/*
	@go get
	staticfiles --build-tags="!dev" --exclude="*.scss" -o files/files.go static/
