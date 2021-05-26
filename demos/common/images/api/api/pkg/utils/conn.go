package utils

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"

	kubernetes "api/pkg/kubernetes"
)

func createTLSConfig(ca, crt, key string) (*tls.Config, error) {
	// Read ca certificate from given path
	caCertPEM, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caCertPEM)
	if !ok {
		panic("Failed to parse root certificate")
	}

	// Read server certificate and key
	cert, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    roots,
	}, nil
}

func CreateTLS(ca, crt, key, listen *string) net.Listener {
	// Load the necessary certificates and keys for tls connection
	config, err := createTLSConfig(*ca, *crt, *key)
	if err != nil {
		log.Fatal("Config failed:", err.Error())
	}

	// Start the server
	listener, err := tls.Listen("tcp", *listen, config)
	if err != nil {
		log.Fatal("Listen failed:", err.Error())
	}

	return listener
}

func CreateTCP(url *string) net.Listener {
	listener, err := net.Listen("tcp4", *url)
	if err != nil {
		log.Fatal(err)
	}

	return listener
}

func PrintTLSState(conn *tls.Conn) {
	log.Print(">>>>>>>>>>>>>>>> State <<<<<<<<<<<<<<<<")
	state := conn.ConnectionState()
	log.Printf("Version: %x", state.Version)
	log.Printf("HandshakeComplete: %t", state.HandshakeComplete)
	log.Printf("DidResume: %t", state.DidResume)
	log.Printf("CipherSuite: %x", state.CipherSuite)
	log.Printf("NegotiatedProtocol: %s", state.NegotiatedProtocol)
	log.Printf("NegotiatedProtocolIsMutual: %t", state.NegotiatedProtocolIsMutual)

	log.Print("Certificate chain:")
	for i, cert := range state.PeerCertificates {
		subject := cert.Subject
		issuer := cert.Issuer
		log.Printf(" %d s:/C=%v/ST=%v/L=%v/O=%v/OU=%v/CN=%s", i, subject.Country, subject.Province, subject.Locality, subject.Organization, subject.OrganizationalUnit, subject.CommonName)
		log.Printf("   i:/C=%v/ST=%v/L=%v/O=%v/OU=%v/CN=%s", issuer.Country, issuer.Province, issuer.Locality, issuer.Organization, issuer.OrganizationalUnit, issuer.CommonName)
	}
	log.Print(">>>>>>>>>>>>>>>> State End <<<<<<<<<<<<<<<<")
}

func handleTLSConnection(c net.Conn) {
	var result string
	// Recieve data from client
	reader := bufio.NewReader(c)
	netData := make([]byte, 4096)
	size, err := reader.Read(netData)
	if err != nil {
		log.Println(err)
	}

	// Execute the given command
	dataString := string(netData[:size])
	log.Println("Command received:", dataString)
	command := kubernetes.ProcessInput(dataString)
	if command.Action != "Error" {
		result = kubernetes.Execute(command)
	} else {
		result = "fail"
	}

	// Send back the responce to the client
	_, err = c.Write([]byte(result))
	if err != nil {
		log.Print(err)
	}

	c.Close()
}

func ListenAndServer(ln net.Listener) {
	defer ln.Close()
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Fatal("Accept failed:", err.Error())
			break
		}
		log.Printf("Connection open: %s", c.RemoteAddr())
		go handleTLSConnection(c)
	}
}
