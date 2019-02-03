package main

import (
	"log"
	"crypto/tls"
	"net/http"
	"flag"
)

//This file will 
//


func main() {
	var svrParam admServerParam
	flag.IntVar(&svrParam.port, "port", 4480, "port on which the webhook svr will listen")
	flag.StringVar(&svrParam.tlsCert, "cert", "./tls/admCert.pem", "Cert for the webhook")
	flag.StringVar(&svrParam.tlsKey, "key", "./tls/admKey.pem", "tls key for the webhook")

	flag.Parse()

	cert , err := tls.LoadX509KeyPair(svrParam.tlsCert, svrParam.tlsKey)
	if err != nil {
		log.Panicln("Error while loading the cert")
	}
	hookSvr := &WebhookSvr{
		Server: &http.Server{
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}}
		}
	}
}