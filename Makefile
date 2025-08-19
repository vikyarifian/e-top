run: build
	@./bin/app

build:
	@go build -o bin/app .

css:
	bunx tailwindcss -i templates/css/app.css -o public/css/styles.css --watch

proxy:
	templ generate --watch --proxy=http://localhost:7878