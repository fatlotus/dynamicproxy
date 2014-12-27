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
)

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
    // TODO(Add support for HTTPS)
    
    // 1. Validate the path to the server.
    parsed, err := url.Parse(path) 
    if err != nil {
        return nil, err
    }
    
    // 2. Prepare a new request payload.
    req, err := http.NewRequest("BIND", path, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Upgrade", "DynamicProxy")
    
    // 3. Add authentication details.
    
    //    (none, since this is just a demo)
    
    // 4. Send the request to the server.
    conn, err := net.Dial("tcp", parsed.Host)
    if err != nil {
        return nil, err
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