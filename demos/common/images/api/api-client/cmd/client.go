package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"io/ioutil"
	log "logging"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type Command struct {
	Action    string
	Target    string
	Arguments []string
}

var (
	connect   *string
	ca        *string
	crt       *string
	key       *string
	arguments *string
	logPath   *string
)

func init() {
	connect = flag.String("conn", "localhost:8000", "The server url e.g. localhost:8000")
	ca = flag.String("ca", "internal/ca.crt", "The file path to ca certificate e.g. ./ca.crt")
	crt = flag.String("crt", "internal/client.crt", "The file path to crt certificate e.g. ./client.crt")
	key = flag.String("key", "internal/client.key", "The file path to client key e.g. ./client.key")
	arguments = flag.String("arg", "", "The arguments to be passed along with the action in JSON format")
	logPath = flag.String("lp", "client.log", "The path to the log file e.g. ./client.log")
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

func createClientConfig(ca, crt, key string) (*tls.Config, error) {
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

	// Read client certificate and key
	cert, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      roots,
	}, nil
}

/* Parse the action and the arguments, create a json from them
*  and return it to the main program
 */
func parseCommand(arguments *string) Command {
	var command Command

	err := json.Unmarshal([]byte(*arguments), &command)
	if err != nil {
		log.Println(err)
	}

	return command
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

func handleConnection(c net.Conn, command Command) {
	// Send command to server
	msg, err := json.Marshal(command)
	if err != nil {
		log.Print(err)
		return
	}
	_, err = c.Write(msg)
	if err != nil {
		log.Print(err)
		return
	}

	reader := bufio.NewReader(c)
	netData := make([]byte, 4096)
	_, err = reader.Read(netData)
	if err != nil {
		log.Println(err)
	}
	dataString := string(netData)
	log.Println("Received response:", dataString)
}

func main() {
	addr := *connect

	// Load the necessary certificates and keys
	config, err := createClientConfig(*ca, *crt, *key)
	if err != nil {
		log.Fatal("Config failed:", err.Error())
	}

	// Connect to the tls server
	conn, err := tls.Dial("tcp", addr, config)
	if err != nil {
		log.Fatalf("Failed to connect: %s", err.Error())
	}
	defer conn.Close()
	log.Printf("Connect to %s succeed", addr)

	printConnState(conn)
	// Parse command given by the user
	command := parseCommand(arguments)

	handleConnection(conn, command)
}
