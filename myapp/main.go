package main

import (
    "net/http"
    "github.com/fatlotus/dynamicproxy"
    "log"
    "flag"
)

var bindurl = flag.String("bindurl", "", "Where to run this application.")
var docroot = flag.String("docroot", "/var/www", "Which files to serve.")

func main() {
    flag.Parse()
    
    if *bindurl == "" {
        log.Fatal("Parameter -bindurl is required.")
    }
    
    listener, err := dynamicproxy.BindURL(*bindurl)
    
    if err != nil {
        log.Fatal(err)
    }
    
    http.Serve(listener, http.FileServer(http.Dir(*docroot)))
}