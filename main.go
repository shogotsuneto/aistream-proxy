package main

import (
	"bufio"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Flushable ResponseWriter helper
type flushWriter struct {
	http.ResponseWriter
	flusher http.Flusher
}

func (fw flushWriter) Write(p []byte) (int, error) {
	n, err := fw.ResponseWriter.Write(p)
	if err == nil {
		fw.flusher.Flush()
	}
	return n, err
}

func main() {
	// CLI flags
	port := flag.String("port", "8080", "Port to listen on")
	bind := flag.String("bind", "127.0.0.1", "Host/IP to bind to")
	target := flag.String("target", "", "Target base URL (e.g. https://api.openai.com)")
	sk := flag.String("sk", "", "Secret key for Authorization header")
	skFile := flag.String("sk-file", "", "File containing the secret key")
	skStdin := flag.Bool("sk-stdin", false, "Read the secret key from stdin")

	flag.Parse()

	// Read SK from file or stdin if not passed directly
	secret := *sk
	switch {
	case *skFile != "":
		data, err := os.ReadFile(*skFile)
		if err != nil {
			log.Fatalf("Failed to read sk file: %v", err)
		}
		secret = strings.TrimSpace(string(data))
	case *skStdin:
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read from stdin: %v", err)
		}
		secret = strings.TrimSpace(input)
	}

	if *target == "" || secret == "" {
		log.Fatal("Must provide --target and one of --sk, --sk-file, or --sk-stdin.")
	}

	targetURL, err := url.Parse(*target)
	if err != nil {
		log.Fatalf("Invalid target URL: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxyURL := *targetURL
		proxyURL.Path = r.URL.Path
		proxyURL.RawQuery = r.URL.RawQuery

		req, err := http.NewRequest(r.Method, proxyURL.String(), r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		req.Header = r.Header.Clone()
		req.Host = proxyURL.Host
		req.Header.Set("Authorization", "Bearer "+secret)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy headers first
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)

		// Ensure ResponseWriter is flushable
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}
		fw := flushWriter{w, flusher}

		// Stream line-by-line
		scanner := bufio.NewScanner(resp.Body) // The split function defaults to [ScanLines]

		// each call to scanner.Scan() reads up to the next '\n'.
		// It strips the newline from the result.
		// This is perfect for handling text/event-streams like:
		//   data: { "id": "..." }
		//   data: { "delta": ... }
		for scanner.Scan() {
			line := scanner.Bytes()
			fw.Write(append(line, '\n')) // restore the newline which the scanner stripped
		}
	})

	address := *bind + ":" + *port
	log.Printf("Proxy listening on http://%s\n", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
