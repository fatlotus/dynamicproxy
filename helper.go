// loadbalancer/main.go
//
// Here we define a simple HTTP Listener that binds to a URL instead of a port.
//
// (See bit.ly/dynamic-reverse-proxies for a specification.) 

package dynamicproxy

import (
    "net"
    "fmt"
    "net/http"
    "bufio"
    "net/url"
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
    "flag"
)

var cert = flag.String("dpcert", "", "What certificate for dynamic proxying.")
var key = flag.String("dpkey", "", "What key to use for dynamic proxying.")
var ca = flag.String("serverca", "", "Which server CA to use (default: system)")

// A fake acceptor that returns just one socket
type passthruListener struct {
    conn net.Conn
    accepted bool
    closed chan bool
}

func (l *passthruListener) Accept() (c net.Conn, e error) {
    if !l.accepted {
        l.accepted = true
        return l.conn, nil
    } else {
        <-l.closed
        err := fmt.Errorf("Socket closed.")
        return nil, err
    }
}

func (l *passthruListener) Close() error {
    l.closed <- true
    return nil
}

func (l *passthruListener) Addr() net.Addr {
    return nil
}

// Creates a Listener that binds to the given URL.
func BindURL(path string) (l net.Listener, e error) {
    // 1. Validate the path to the server.
    parsed, err := url.Parse(path) 
    if err != nil {
        return nil, err
    }
    
    if parsed.Scheme != "https" {
        return nil, fmt.Errorf("Scheme %s is unsupported", parsed.Scheme)
    }
    
    // 2. Prepare a new request payload.
    req, err := http.NewRequest("BIND", path, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Upgrade", "DynamicProxy")
    
    // 3. Add authentication details.
    var conn net.Conn
    var pool *x509.CertPool
    
    if *ca != "" {
        pool = x509.NewCertPool()
        data, err := ioutil.ReadFile(*ca)
        if err != nil {
            return nil, err
        }
        
        if !pool.AppendCertsFromPEM(data) {
            return nil, fmt.Errorf("Unable to add certificates.")
        }
    }
    
    var certs []tls.Certificate
    
    if *cert != "" && *key != "" {
        cert, err := tls.LoadX509KeyPair(*cert, *key)
        if err != nil {
            return nil, err
        }
        
        certs = []tls.Certificate{cert}
    }
    
    // 4. Send the resulting request to the server.
    conn, err = tls.Dial("tcp", parsed.Host, &tls.Config{
        RootCAs: pool,
        Certificates: certs,
    })
    
    if err != nil {
        return nil, err
    }
    
    if conn == nil {
        return nil, fmt.Errorf("Conn is nil!")
    }
    
    req.Write(conn)
    
    // 5. Wait for a response.
    resp, err := http.ReadResponse(bufio.NewReader(conn), req)
    if err != nil {
        return nil, err
    }
    
    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("Bad status code: %d", resp.StatusCode)
    }
    
    listener := &passthruListener{
        conn: conn,
        closed: make(chan bool, 1),
    }
    
    return listener, nil
}