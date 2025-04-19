build:
	go build -o bin/aistream-proxy main.go

# make run args="--target https://api.openai.com --sk-stdin"
run:
	go run main.go ${args}
