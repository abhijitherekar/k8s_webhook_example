package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

//This file will
//

func main() {
	var svrParam admServerParam
	flag.IntVar(&svrParam.port, "port", 4480, "port on which the webhook svr will listen")
	flag.StringVar(&svrParam.tlsCert, "cert", "./tls/cert.pem", "Cert for the webhook")
	flag.StringVar(&svrParam.tlsKey, "key", "./tls/key.pem", "tls key for the webhook")

	flag.Parse()

	cert, err := tls.LoadX509KeyPair(svrParam.tlsCert, svrParam.tlsKey)
	if err != nil {
		log.Panicln("Error while loading the cert", err)
	}
	hookSvr := &WebhookSvr{
		Server: &http.Server{
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", hookSvr.Serve)
	mux.HandleFunc("/validate", hookSvr.Serve)
	hookSvr.Server.Handler = mux

	go func() {
		if err := hookSvr.Server.ListenAndServeTLS("", ""); err != nil {
			log.Panicln("could not the server", err)
		}
	}()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	log.Println("Started the server listening 4480")
}
