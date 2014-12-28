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
    "io"
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
    "net/url"
    "flag"
    "sync"
)

var cert = flag.String("cert", "", "What certificate for SSL.")
var key = flag.String("key", "", "What key to use for SSL.")
var clientca = flag.String("clientca", "", "A CA to use for access control.")
var bind = flag.String("bind", ":80", "Which address:port to bind on.")

type backend struct {
    conn net.Conn
    bufrw *bufio.ReadWriter
    path string
}

type dynamicProxy struct {
    backend *backend
    mutex sync.Mutex
}

func (p *dynamicProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    hj, ok := w.(http.Hijacker)
    if !ok {
        panic("Webserver does not support hijacking!")
    }
    
    // Avoid concurrent proxying, at least for now.
    p.mutex.Lock()
    defer p.mutex.Unlock()
    
    if r.Header.Get("Upgrade") == "DynamicProxy" {
        // 1. Validate authentication details in the request.
        
        allowed := false
        
        for _, cert := range(r.TLS.PeerCertificates) {
            url, err := url.Parse(cert.Subject.CommonName)
            if err == nil {
                if url.Host == r.Host &&
                   strings.HasPrefix(r.URL.Path, url.Path) {
                    allowed = true
                }
            }
        }
        
        if !allowed {
            w.WriteHeader(401)
            fmt.Fprintf(w, "Dynamic proxying not allowed.")
            log.Print("Blocked attempt to proxy ", r.URL)
            return
        }
        
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
            for key := range(response.Header) {
                for _, value := range(response.Header[key]) {
                    w.Header().Add(key, value)
                }
            }
            
            w.WriteHeader(response.StatusCode)
            
            io.Copy(w, response.Body)
        }
    }
}

func main() {
    flag.Parse()
    
    // Run the dynamic proxy on /.
    http.Handle("/", new(dynamicProxy))
    
    var config *tls.Config 
    
    // Configure how we verify client certificates.
    if *clientca != "" {
        pool := x509.NewCertPool()
        data, err := ioutil.ReadFile(*clientca)
        if err != nil {
            log.Fatal(err)
        }
    
        if !pool.AppendCertsFromPEM(data) {
            log.Fatal("Unable to add certificates.")
        }
        
        config = &tls.Config{
            ClientCAs: pool,
            ClientAuth: tls.VerifyClientCertIfGiven,
        }
    } else {
        config = &tls.Config{
            ClientAuth: tls.RequestClientCert,
        }
        
        log.Print("Warning: Client certificate verification disabled!")
    }
    
    server := &http.Server{
        Addr: *bind,
        TLSConfig: config,
    }
    
    // Configure TLS certificates.
    if *cert == "" || *key == "" {
        log.Fatal("The cert and key options are required.")
    }
    
    log.Print("Listening on ", *bind)
    log.Fatal(server.ListenAndServeTLS(*cert, *key))
}