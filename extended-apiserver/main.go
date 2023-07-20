package main

import (
	"crypto/x509"
	"flag"
	"fmt"
	k8s_extended_apiserver "github.com/Saleh7127/k8s-extended-apiserver"
	"github.com/Saleh7127/k8s-extended-apiserver/lib/certstore"
	"github.com/Saleh7127/k8s-extended-apiserver/lib/server"
	"github.com/gorilla/mux"
	"github.com/spf13/afero"
	"k8s.io/client-go/util/cert"
	"log"
	"net"
	"net/http"
)

func ErrorNotNill(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

var crtPath = "/home/user/go/src/github.com/Saleh7127/k8s-extended-apiserver/certs"

func main() {

	var proxy = false
	flag.BoolVar(&proxy, "receive-proxy-request", proxy, "receive forwarded requests from api-server")
	flag.Parse()
	fmt.Println("forward", proxy)

	fs := afero.NewOsFs()
	store, err := certstore.NewCertStore(fs, crtPath)
	ErrorNotNill(err)

	err = store.NewCA(k8s_extended_apiserver.ExApiServer)
	ErrorNotNill(err)

	serverCert, serverKey, err := store.NewServerCertPair(cert.AltNames{
		IPs: []net.IP{
			net.ParseIP(k8s_extended_apiserver.ExApiIp),
		},
	})
	ErrorNotNill(err)

	err = store.Write("tls", serverCert, serverKey)
	ErrorNotNill(err)

	clientCert, clientKey, err := store.NewClientCertPair(cert.AltNames{
		DNSNames: []string{"abu"},
	})
	ErrorNotNill(err)

	err = store.Write("abu", clientCert, clientKey)
	ErrorNotNill(err)

	//-------------------------------------------------------------
	apiserverStore, err := certstore.NewCertStore(fs, crtPath)
	ErrorNotNill(err)

	if proxy {
		err = apiserverStore.LoadCA(k8s_extended_apiserver.ApiServer)
		ErrorNotNill(err)
	}

	//--------------------------------------------------------------

	rhCACertPool := x509.NewCertPool()
	rhStore, err := certstore.NewCertStore(fs, crtPath)

	ErrorNotNill(err)
	if proxy {
		err = rhStore.LoadCA(k8s_extended_apiserver.ReqHeader)
		ErrorNotNill(err)
		rhCACertPool.AppendCertsFromPEM(rhStore.CACertBytes())
	}

	// ------------------------------------------------------------

	cfg := server.Config{
		Address:     k8s_extended_apiserver.ExApiAdd,
		CACertFiles: []string{
			//store.CertFile("ca"),
		},
		CertFile: store.CertFile("tls"),
		KeyFile:  store.KeyFile("tls"),
	}

	if proxy {
		cfg.CACertFiles = append(cfg.CACertFiles, apiserverStore.CertFile("ca"))
		cfg.CACertFiles = append(cfg.CACertFiles, rhStore.CertFile("ca"))
	}

	srv := server.NewGenericServer(cfg)

	r := mux.NewRouter()
	r.HandleFunc(k8s_extended_apiserver.ExApiPath, func(writer http.ResponseWriter, request *http.Request) {

		user := k8s_extended_apiserver.SysAnonymous
		src := "-"

		if len(request.TLS.PeerCertificates) > 0 {
			opts := x509.VerifyOptions{
				Roots:     rhCACertPool,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			if _, err := request.TLS.PeerCertificates[0].Verify(opts); err != nil {
				user = request.TLS.PeerCertificates[0].Subject.CommonName // username from client cert
				src = k8s_extended_apiserver.ClientCertCn
			} else {
				user = request.Header.Get(k8s_extended_apiserver.RemoteUser) // username from header value passed by api-server
				src = k8s_extended_apiserver.RemoteUser
			}
		}

		vars := mux.Vars(request)
		writer.WriteHeader(http.StatusOK)
		fmt.Fprintf(writer, "Resource: %v requested by user[%s]=%s\n", vars["resource"], src, user)
	})
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})
	srv.ListenAndServe(r)

}
