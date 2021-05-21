package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	kubernetes "api/pkg/kubernetes"
)

var (
	listen  *string
	ca      *string
	crt     *string
	key     *string
	logPath *string
)

func init() {
	listen = flag.String("l", "localhost:8000", "The server url to listen e.g. localhost:8000")
	ca = flag.String("ca", "./internal/ca.crt", "The file path to ca certificate e.g. ./ca.crt")
	crt = flag.String("crt", "./internal/server.crt", "The file path to server crt certificate e.g. ./server.crt")
	key = flag.String("key", "./internal/server.key", "The file path to server key e.g. ./server.key")
	logPath = flag.String("lp", "./server.log", "The path to the log file e.g. ./server.log")
	flag.Parse()

	// Open log file
	logFile, err := os.OpenFile(*logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	log.SetFlags(log.Ldate | log.Ltime)
	log.SetOutput(logFile)

	// Setup signal catching
	sigs := make(chan os.Signal, 1)
	// Catch all signals since not explicitly listing
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	// Method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL: %s", s)
		os.Exit(1)
	}()
}

func createServerConfig(ca, crt, key string) (*tls.Config, error) {
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

func printConnState(conn *tls.Conn) {
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

func handleConnection(c net.Conn) {
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

func main() {
	// Load the necessary certificates and keys
	config, err := createServerConfig(*ca, *crt, *key)
	if err != nil {
		log.Fatal("Config failed:", err.Error())
	}

	// Start the server
	ln, err := tls.Listen("tcp", *listen, config)
	if err != nil {
		log.Fatal("Listen failed:", err.Error())
	}

	log.Printf("Listen on %s", *listen)

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Fatal("Accept failed:", err.Error())
			break
		}
		log.Printf("Connection open: %s", c.RemoteAddr())
		printConnState(c.(*tls.Conn))
		go handleConnection(c)
	}
}
