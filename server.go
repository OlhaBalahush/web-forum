package main

import (
	"crypto/tls"
	"log"
	"main/helpers"

	"net/http"
	"time"
)

var dockerize *bool

func main() {
	helpers.ParseFlags()
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}
	srv := &http.Server{
		Addr:         ":8080",
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		Handler:      routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	err := srv.ListenAndServe()
	log.Println(err)
}
