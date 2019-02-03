package main

import (
	"net/http"
)

type WebhookSvr struct {
	Server *http.Server
}

type admServerParam struct {
	port    int
	tlsCert string
	tlsKey  string
}
