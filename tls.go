package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
)

func getTLSConfig() (string, string, *tls.Config) {
	caCert, err := ioutil.ReadFile("ca.crt")
	if err != nil {
		panic("Error read cert: " + err.Error())
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	tlsConfig.BuildNameToCertificate()
	return "ca.crt", "ca.key", tlsConfig
}
