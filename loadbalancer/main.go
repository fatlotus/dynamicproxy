// loadbalancer/main.go
//
// Here we define the simplest possible "dynamic load balancer" that complies
// with the specification.
//
// (See bit.ly/dynamic-reverse-proxies for aformentioned spec.) 

package main

import (
    "fmt"
    "net"
    "net/http"
    "log"
    "bufio"
    "strings"
)

type backend struct {
    conn net.Conn
    bufrw *bufio.ReadWriter
    path string
}

type dynamicProxy struct {
    backend *backend
}

func (p *dynamicProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    hj, ok := w.(http.Hijacker)
    if !ok {
        panic("Webserver does not support hijacking!")
    }
    
    if r.Header.Get("Upgrade") == "DynamicProxy" {
        // 1. Validate authentication details in the request.
        
        //    (no authentication, since this is a demo)
        
        // 2. Return 200 Success.
        
        w.WriteHeader(200)
        
        // 3. Keep the connection open-
        
        conn, bufrw, err := hj.Hijack()
        if err != nil {
            panic(err)
        }
        
        if p.backend != nil {
            log.Print("Switching backends")
            p.backend.conn.Close()
        } else {
            log.Print("We now have a backend!")
        }
        
        //    - making sure to save the request path.
        
        p.backend = &backend{
            conn: conn,
            bufrw: bufrw,
            path: r.URL.Path,
        }
    } else {
        // 1. Find the backends whose paths are a prefix of this request.
        
        if p.backend == nil || !strings.HasPrefix(r.URL.Path, p.backend.path) {
            
            // If no such backend exists, return 502 Bad Gateway
            w.WriteHeader(502)
            fmt.Fprintf(w, "502: No proxies available.")
        } else {
            
            // 3. Proxy the request onto the socket-
            if err := r.Write(p.backend.conn); err != nil {
                panic(err)
            }
            
            //    - and wait for a response from the backend.
            response, err := http.ReadResponse(p.backend.bufrw.Reader, r)
            if err != nil {
                w.WriteHeader(502)
                fmt.Fprintf(w, "%s", err.Error())
                log.Print("Backend crashed: ", err.Error())
                p.backend = nil
                return
            }
            
            // 4. Forward the response back to the client.
            conn, bufrw, err := hj.Hijack()
            if err != nil {
                panic(err)
            }
            
            if err := response.Write(bufrw.Writer); err != nil {
                log.Print(err)
            }
            
            bufrw.Flush()
            conn.Close() // TODO(Add support for Keep-Alive)
        }
    }
}

func main() {
    http.Handle("/", new(dynamicProxy))
    log.Fatal(http.ListenAndServe(":8080", nil))
}