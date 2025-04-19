# aistream-proxy

A lightweight proxy for the OpenAI API, providing a single local entry point for multiple projects.  
🧪 Great for testing, development tools, and custom AI clients.

## Usage

```bash
aistream-proxy \
  --target https://api.openai.com \
  --port 8080 \
  --sk-file ./secret_key
```

## Usage (Docker)

```bash
cat secret-key | sudo docker run -i -p 18080:8080 aistream-proxy:latest --bind 0.0.0.0 --port 8080 --target https://api.openai.com --sk-stdin
```

## License

MIT
