package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const endpoint = "https://hardgforgifs.github.io"

type Proxier struct {
	ips []string
	ch  chan string
}

func (p *Proxier) Dink(ipAddress string) {
	p.ch <- ipAddress
}

func NewProxier() *Proxier {
	return &Proxier{ch: make(chan string)}
}

func (p *Proxier) whatsHappen() {
	log.Println("cleaning the chan'el")
	for ip := range p.ch {
		p.ips = append(p.ips, ip)
	}
}

// HandleRequest will authenticate the request the request
// and pass it off the proxy handler
func (p *Proxier) HandleRequest(w http.ResponseWriter, r *http.Request) {
	url := p.getProxyURL(r.URL)
	p.serveReverseProxy(w, r, url)
}

func (p *Proxier) serveReverseProxy(w http.ResponseWriter, r *http.Request, url *url.URL) {
	proxy := httputil.NewSingleHostReverseProxy(url)

	p.Dink(r.RemoteAddr)

	// update headers to indicate proxy
	r.URL.Host = url.Host
	r.URL.Scheme = url.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = url.Host
	proxy.ServeHTTP(w, r)
}

func (p *Proxier) getProxyURL(url *url.URL) *url.URL {
	proxiedURL, _ := url.Parse(endpoint + url.Path)
	log.Printf("%s -> %s", url, proxiedURL)
	return proxiedURL
}

func main() {
	p := NewProxier()

	proxyServ := http.NewServeMux()
	proxyServ.HandleFunc("/", p.HandleRequest)

	go func() {
		log.Println("Proxy server started on 0.0.0.0:8081")
		http.ListenAndServe("0.0.0.0:8081", proxyServ)
	}()
	adminServ := http.NewServeMux()
	adminServ.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("admin site"))
	})
	adminServ.HandleFunc("/stat", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{\n    views: 4\n}"))
	})

	dur, _ := time.ParseDuration("5s")
	go func() {
		for {
			p.whatsHappen()
			log.Printf("the scoop: %+v", p.ips)
			time.Sleep(dur)
		}
	}()

	log.Println("Admin server started on 0.0.0.0:8082")
	http.ListenAndServe("0.0.0.0:8082", adminServ)
}
