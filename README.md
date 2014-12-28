# Dynamic Reverse Proxy Example

This program implements a basic dynamic reverse proxy. To install it, configure
a Go workspace, then run:

```
$ go get github.com/fatlotus/dynamicproxy
$ go install github.com/fatlotus/dynamicproxy/...
```

These commands will create `bin/loadbalancer` and `bin/myapp`; the first is the
load balancer program, and the second is a sample application to facilitate 
testing.

Please see [bit.ly/dynamic-reverse-proxy](http://bit.ly/dynamic-reverse-proxy)
for the complete protocol specification, as well as general considerations for
other implementors.

## SSL Configuration

This application requires the use of HTTPS with SSL client certificates, which
require performing the following steps:

First, create a new server SSL certificate. If you need a self-signed
certificate, run the following steps; otherwise, request one from a local
signing authority, or purchase one from an online SSL vendor.

If you're generating your own "self-signed" certificate, use the following
OpenSSL command. Be sure to replace "localhost" in the instructions below

```
$ openssl req -new -x509 -nodes -days 365 -keyout server.key -out server.crt
Generating a 1024 bit RSA private key
...................++++++
......++++++
writing new private key to 'server.key'
-----
You are about to be asked to enter information that will be incorporated
into your certificate request.
What you are about to enter is what is called a Distinguished Name or a DN.
There are quite a few fields but you can leave some blank
For some fields there will be a default value,
If you enter '.', the field will be left blank.
-----
Country Name (2 letter code) [AU]:US
State or Province Name (full name) [Some-State]:Illinois
Locality Name (eg, city) []:Chicago
Organization Name (eg, company) [Internet Widgits Pty Ltd]:University of Chicago
Organizational Unit Name (eg, section) []:Department of Computer Science
Common Name (e.g. server FQDN or YOUR name) []:localhost    
Email Address []:
$
```

Next, prepare a Certificate Authority; this authority will sign the certificates
of applications wishing to host code on your website. As before, create a new
self-signed certificate, but this time, use your own name for the Common Name.

```
$ openssl req -new -x509 -nodes -days 365 -keyout ca.key -out ca.crt
Generating a 1024 bit RSA private key
...................++++++
......++++++
writing new private key to 'ca.key'
-----
You are about to be asked to enter information that will be incorporated
into your certificate request.
What you are about to enter is what is called a Distinguished Name or a DN.
There are quite a few fields but you can leave some blank
For some fields there will be a default value,
If you enter '.', the field will be left blank.
-----
Country Name (2 letter code) [AU]:US
State or Province Name (full name) [Some-State]:Illinois
Locality Name (eg, city) []:Chicago
Organization Name (eg, company) [Internet Widgits Pty Ltd]:University of Chicago
Organizational Unit Name (eg, section) []:Department of Computer Science
Common Name (e.g. server FQDN or YOUR name) []:Jeremy Archer (CA)    
Email Address []:
$
```

Whenever you run a new application, have the application author generate a 
Certificate Signing Request, or CSR, and forward it to you. This can be
performed with the following command. Under Common Name, they should put the URL
they are requesting to proxy from.

```
$ openssl req -new -nodes -keyout client.key -out client.csr
Generating a 1024 bit RSA private key
........................................................++++++
.........++++++
writing new private key to 'client.key'
-----
You are about to be asked to enter information that will be incorporated
into your certificate request.
What you are about to enter is what is called a Distinguished Name or a DN.
There are quite a few fields but you can leave some blank
For some fields there will be a default value,
If you enter '.', the field will be left blank.
-----
Country Name (2 letter code) [AU]:US
State or Province Name (full name) [Some-State]:Illinois
Locality Name (eg, city) []:Chicago
Organization Name (eg, company) [Internet Widgits Pty Ltd]:University of Chicago
Organizational Unit Name (eg, section) []:Department of Computer Science
Common Name (e.g. server FQDN or YOUR name) []:https://localhost:8080/myapplication
Email Address []:

Please enter the following 'extra' attributes
to be sent with your certificate request
A challenge password []:
An optional company name []:
```

They should then send you `client.csr`, along with verification of their 
identity. With the identity verification and CSR, check that they are allowed
access to the given CSR.

```
$ openssl req -noout -text -verify -in client.csr
verify OK
Certificate Request:
    Data:
        Version: 0 (0x0)
        Subject: C=US, ST=Illinois, L=Chicago, O=University of Chicago, OU=Department of Computer Science, CN=https://localhost:8080/myapplication
        Subject Public Key Info:
            Public Key Algorithm: rsaEncryption
            RSA Public Key: (1024 bit)
                Modulus (1024 bit):
                    00:b6:89:b5:42:40:70:ee:11:1f:c8:4b:55:a6:16:
                    d1:a7:c7:0c:86:c4:54:71:bb:18:98:bc:7f:67:67:
                    63:69:87:a7:7b:6e:4b:d7:be:b5:28:8d:67:b7:3b:
                    ca:83:51:28:23:e7:cd:8e:ef:85:f7:2f:0d:2b:bd:
                    3c:2d:94:64:a8:3c:8c:b6:a3:f6:33:24:04:c5:96:
                    92:af:50:b2:a4:c4:61:91:18:75:f4:19:89:6a:c8:
                    20:7f:a9:9b:61:c7:67:8a:f7:5f:a0:db:23:61:a3:
                    a1:a2:aa:c3:28:50:85:a3:12:66:94:51:53:11:ab:
                    05:58:4d:18:8c:1e:48:e0:bd
                Exponent: 65537 (0x10001)
        Attributes:
            a0:00
    Signature Algorithm: sha1WithRSAEncryption
        b6:78:f3:6f:58:16:18:2b:c3:8b:88:92:5e:15:b6:03:11:8d:
        f4:fb:a3:12:37:d9:a8:5c:3b:17:c8:08:8e:98:e8:0b:a0:2f:
        e3:43:65:9b:b1:15:c2:c6:6b:ab:94:e0:0c:46:1d:26:e6:57:
        04:08:99:4b:a6:c3:b6:22:bf:8c:55:d6:48:7b:35:e2:9d:4c:
        ad:63:bb:b7:2d:8c:e4:94:c4:05:02:4b:72:b7:42:47:e5:ed:
        7e:06:cd:ea:3f:7a:4f:f6:e0:39:6c:71:e4:19:dd:2a:c5:b8:
        d4:dd:be:c2:ee:41:9d:98:67:ff:e7:83:95:67:37:da:9f:6a:
        fb:88
```

Once you've observed that the certificate is okay (`verify OK`) and that the
user is allowed to proxy for the given URL (see `CN=` under `Subject`), sign
the CSR, creating a certificate file to send back to the author.

```
$ cat > openssl.cnf
[ ssl_client ]
extendedKeyUsage = clientAuth
^D
$ openssl x509 -req -days 365 -in client.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out client.crt -extensions ssl_client -extfile openssl.cnf
Signature ok
subject=/C=US/ST=Illinois/L=Chicago/O=University of Chicago/OU=Department of Computer Science/CN=https://localhost:8080/myapplication
Getting CA Private Key
```

After that, send `client.crt` back to the author. They will use it, in tandem
with their private key, to run their application.

## Running the Reverse Proxy

To run the reverse proxy (assuming the certificates are located as suggested
above), run the following command:

```
$ bin/loadbalancer -cert=server.crt -key=server.key -clientca=ca.crt -bind=:8080
2014/12/28 13:50:48 Listening on :8080
...
^C
```

## Running Proxied Applications

Generally, the configuration of applications will vary depending on how they 
are developed. With the sample application provided, use the following command
to launch a listener on :8080.

```
$ bin/myapp -serverca=server.crt -dpcert=client.crt -dpkey=client.key -bindurl=https://localhost:8080/app
```

## License

All code in this repository is covered under the MIT License:

> Copyright (c) 2014 Jeremy Archer
> 
> Permission is hereby granted, free of charge, to any person obtaining a copy
> of this software and associated documentation files (the "Software"), to deal
> in the Software without restriction, including without limitation the rights
> to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
> copies of the Software, and to permit persons to whom the Software is
> furnished to do so, subject to the following conditions:
> 
> The above copyright notice and this permission notice shall be included in
> all copies or substantial portions of the Software.
> 
> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
> IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
> FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
> AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
> LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
> OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
> THE SOFTWARE.