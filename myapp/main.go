package main

import (
    "net/http"
    "dynamicproxy"
)

func main() {
    listener, err := dynamicproxy.BindURL("http://localhost:8080")
    
    if err != nil {
        panic(err)
    }
    
    http.Serve(listener, http.FileServer(http.Dir(".")))
}