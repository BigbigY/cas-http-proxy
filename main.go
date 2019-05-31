package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-cas/cas"
	"github.com/go-chi/chi"
	"github.com/golang/glog"
)

var (
	casURL    string
	appAddr   string
	authFile  string
	proxyAddr string
)

func init() {
	flag.StringVar(&casURL, "cas-url", "https://cas.example.com/cas", "cas url")
	flag.StringVar(&authFile, "auth-file", "./AuthUser.db", "auth file")
	flag.StringVar(&appAddr, "app-addr", "127.0.0.1:5601", "app addr")
	flag.StringVar(&proxyAddr, "proxy-addr", ":8080", "proxy addr")
	flag.Parse()
}

func main() {
	url, _ := url.Parse(casURL)
	client := cas.NewClient(&cas.Options{URL: url})
	root := chi.NewRouter()
	root.Use(client.Handler)
	root.HandleFunc("/*", handle)

	server := &http.Server{
		Addr:    proxyAddr,
		Handler: client.Handle(root),
	}

	if err := server.ListenAndServe(); err != nil {
		glog.Fatal(err)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	user := strings.Split(cas.Username(r), "@")[0]
	r.Header.Add("X-WEBAUTH-USER", user)
	r.Host, r.URL.Host = appAddr, appAddr
	r.RequestURI, r.URL.Scheme = "", "http"

	// Access permissions
	if !isAccess(user) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	resp, err := (&http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}).Do(r)

	// This prints to STDOUT to show that processing has started
	ctx := r.Context()

	select {
	case <-time.After(1 * time.Second): // If we receive a message after 2 seconds
		// that means the request has been processed
		// We then write this as the response
		if err != nil {
			glog.Fatal(err)
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			glog.Fatal(err)
		}

		header := w.Header()
		for key, value := range resp.Header {
			for _, item := range value {
				header.Add(key, item)
			}
		}
		w.WriteHeader(resp.StatusCode)
		if _, err = io.Copy(w, bytes.NewReader(b)); err != nil {
			glog.Fatal(err)
		}
		resp.Body.Close()
	case <-ctx.Done(): // If the request gets cancelled, log it
		// to STDERR
		fmt.Fprint(os.Stderr, "request cancelled\n")
	}
}

func isAccess(username string) bool {
	file, err := os.Open(authFile)
	if err != nil {
		fmt.Println("open file failed, err:", err)
		os.OpenFile(authFile, os.O_CREATE, 0644)
	}

	defer file.Close()
	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		line = strings.Replace(line, " ", "", -1)
		line = strings.Replace(line, "\n", "", -1)

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println("read file failed,err:", err)
		}

		if username == line {
			return true
		}
	}
	return false
}
