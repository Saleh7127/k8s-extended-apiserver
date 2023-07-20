package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	k8s_extended_apiserver "github.com/Saleh7127/k8s-extended-apiserver"
	"github.com/Saleh7127/k8s-extended-apiserver/lib/certstore"
	"github.com/Saleh7127/k8s-extended-apiserver/lib/server"
	"github.com/gorilla/mux"
	"github.com/spf13/afero"
	"io"
	"k8s.io/client-go/util/cert"
	"log"
	"net"
	"net/http"
	"time"
)

func ErrorNotNill(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

var crtPath = "/home/user/go/src/github.com/Saleh7127/k8s-extended-apiserver/certs"

func main() {

	var proxy = false
	flag.BoolVar(&proxy, "send-proxy-request", proxy, "forward requests to database extended apiserver")
	flag.Parse()
	fmt.Println("forward", proxy)

	fs := afero.NewOsFs()
	store, err := certstore.NewCertStore(fs, crtPath)

	ErrorNotNill(err)

	err = store.NewCA(k8s_extended_apiserver.ApiServer)
	ErrorNotNill(err)

	serverCert, serverKey, err := store.NewServerCertPair(cert.AltNames{
		IPs: []net.IP{
			net.ParseIP(k8s_extended_apiserver.ApiIp),
		},
	})
	ErrorNotNill(err)

	err = store.Write("tls", serverCert, serverKey)
	ErrorNotNill(err)

	clientCert, clientKey, err := store.NewClientCertPair(cert.AltNames{
		DNSNames: []string{"uba"},
	})
	ErrorNotNill(err)

	err = store.Write("uba", clientCert, clientKey)
	ErrorNotNill(err)

	// -> generate client certificate for extended server <-
	// -----------------------------------------------------------------------------
	rhStore, err := certstore.NewCertStore(fs, crtPath)
	ErrorNotNill(err)

	err = rhStore.InitCA(k8s_extended_apiserver.ReqHeader)
	ErrorNotNill(err)

	rhClientCert, rhClientKey, err := rhStore.NewClientCertPair(cert.AltNames{
		DNSNames: []string{k8s_extended_apiserver.ApiServer}, // because apiserver is making the calls to extended api server
	})
	ErrorNotNill(err)

	err = rhStore.Write(k8s_extended_apiserver.ApiServer, rhClientCert, rhClientKey)
	ErrorNotNill(err)

	rhCert, err := tls.LoadX509KeyPair(rhStore.CertFile(k8s_extended_apiserver.ApiServer), rhStore.KeyFile(k8s_extended_apiserver.ApiServer))
	ErrorNotNill(err)

	// -----------------------------------------------------------------------------

	// -----------------------------------------------------------------------------

	easCACertPoo := x509.NewCertPool()
	if proxy {
		easStore, err := certstore.NewCertStore(fs, crtPath)
		ErrorNotNill(err)

		err = easStore.LoadCA(k8s_extended_apiserver.ExApiServer)
		ErrorNotNill(err)

		easCACertPoo.AppendCertsFromPEM(easStore.CACertBytes())
	}

	cfg := server.Config{
		Address: k8s_extended_apiserver.ApiAdd,
		CACertFiles: []string{
			store.CertFile("ca"),
		},
		CertFile: store.CertFile("tls"),
		KeyFile:  store.KeyFile("tls"),
	}

	srv := server.NewGenericServer(cfg)

	r := mux.NewRouter()
	r.HandleFunc(k8s_extended_apiserver.ApiPath, func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		writer.WriteHeader(http.StatusOK)
		fmt.Fprintf(writer, "Resource: %v\n", vars["resource"])
	})

	if proxy {
		r.HandleFunc(k8s_extended_apiserver.ExApiPath, func(writer http.ResponseWriter, request *http.Request) {
			tr := &http.Transport{
				MaxIdleConnsPerHost: 10,
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{rhCert},
					RootCAs:      easCACertPoo,
				},
			}
			Client := http.Client{
				Transport: tr,
				Timeout:   time.Duration(5 * time.Second),
			}
			u := *request.URL
			u.Scheme = "https"
			u.Host = "127.0.0.2:8443"
			fmt.Printf("forwarding request to %v\n", u.String())

			req, _ := http.NewRequest(request.Method, u.String(), nil)
			if len(request.TLS.PeerCertificates) > 0 {
				req.Header.Set(k8s_extended_apiserver.RemoteUser, request.TLS.PeerCertificates[0].Subject.CommonName)
			}
			resp, err := Client.Do(req)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(writer, "error: %v\n", err.Error())
				return
			}
			defer resp.Body.Close()

			writer.WriteHeader(http.StatusOK)

			io.Copy(writer, resp.Body)
		})
	}

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})
	srv.ListenAndServe(r)

}
