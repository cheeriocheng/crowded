// crowded: phones send input, main screens show the combined result.
package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net"
	"net/http"
)

//go:embed web
var webFiles embed.FS

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	web, err := fs.Sub(webFiles, "web")
	if err != nil {
		log.Fatal(err)
	}

	h := newHub()
	go h.run()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", page(web, "phone.html"))
	mux.HandleFunc("GET /crowd", page(web, "crowd.html"))
	mux.HandleFunc("GET /ws", h.handleWS)

	_, port, _ := net.SplitHostPort(*addr)
	log.Printf("main screen: http://localhost:%s/crowd", port)
	log.Printf("phone page:  http://localhost:%s/", port)
	for _, ip := range lanAddrs() {
		log.Printf("from a phone on the same Wi-Fi: http://%s:%s/", ip, port)
	}
	log.Fatal(http.ListenAndServe(*addr, mux))
}

// page serves one embedded HTML file.
func page(fsys fs.FS, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFileFS(w, r, fsys, name)
	}
}

// lanAddrs lists this machine's local network addresses.
func lanAddrs() []string {
	var out []string
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		ipn, ok := a.(*net.IPNet)
		if ok && !ipn.IP.IsLoopback() && ipn.IP.To4() != nil {
			out = append(out, ipn.IP.String())
		}
	}
	return out
}
