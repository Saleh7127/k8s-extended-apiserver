package main

import (
	"fmt"
	"github.com/Saleh7127/k8s-extended-apiserver/lib/certstore"
	"github.com/Saleh7127/k8s-extended-apiserver/lib/server"
	"github.com/gorilla/mux"
	"github.com/spf13/afero"
	"k8s.io/client-go/util/cert"
	"log"
	"net"
	"net/http"
)

var crtPath = "/home/user/go/src/github.com/Saleh7127/k8s-extended-apiserver/certs"

func main() {

	fs := afero.NewOsFs()
	store, err := certstore.NewCertStore(fs, "/tmp/extended-server-connection")
	if err != nil {
		log.Fatalln(err)
	}
	err = store.NewCA("extended-apiserver")
	if err != nil {
		log.Fatalln(err)
	}

	serverCert, serverKey, err := store.NewServerCertPair(cert.AltNames{
		IPs: []net.IP{
			net.ParseIP("127.0.0.2"),
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	err = store.Write("tls", serverCert, serverKey)
	if err != nil {
		log.Fatalln(err)
	}

	clientCert, clientKey, err := store.NewClientCertPair(cert.AltNames{
		DNSNames: []string{"uba"},
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = store.Write("uba", clientCert, clientKey)
	if err != nil {
		log.Fatalln(err)
	}

	cfg := server.Config{
		Address: "127.0.0.2:8443",
		CACertFiles: []string{
			store.CertFile("ca"),
		},
		CertFile: store.CertFile("tls"),
		KeyFile:  store.KeyFile("tls"),
	}

	srv := server.NewGenericServer(cfg)

	r := mux.NewRouter()
	r.HandleFunc("/extended-apiserver/{resource}", func(writer http.ResponseWriter, request *http.Request) {
		//vars := mux.Vars(request)
		writer.WriteHeader(http.StatusOK)
		fmt.Fprintf(writer, "\"Resource: %v\\n\", vars[\"resource\"]")
	})
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})
	srv.ListenAndServe(r)

}
