package main

import (
    "net/http"
    "github.com/fatlotus/dynamicproxy"
    "log"
    "flag"
)

func main() {
    flag.Parse()
    
    listener, err := dynamicproxy.BindURL("https://localhost:8080/app")
    
    if err != nil {
        log.Fatal(err)
    }
    
    http.Serve(listener, http.FileServer(http.Dir(".")))
}