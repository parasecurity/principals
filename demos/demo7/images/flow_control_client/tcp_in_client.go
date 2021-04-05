package main

import (
	"flag"
	"fmt"
	"net"
)

var args struct {
	ip   *string
	port *string
	msg  *string
}

func init() {
	args.ip = flag.String("ip", "localhost", "The server ip")
	args.port = flag.String("port", "12345", "The server port")
	args.msg = flag.String("msg", "Hello World!", "The message to send")
	flag.Parse()
}

func main() {
	//test client that reads data from stdin and writes to connection
	url := *args.ip + ":" + *args.port
	c, err := net.Dial("tcp4", url)
	if err != nil {
		fmt.Println(err)
		return
	}

	// for {
	// 	reader := bufio.NewReader(os.Stdin)
	// 	fmt.Print(">> ")
	// 	text, err := reader.ReadString('\n')
	// 	if err != nil {
	// 		//if we get an io error, let the client terminate
	// 		fmt.Println("Error, exiting")
	// 		break
	// 	} else {
	// 		if strings.TrimSpace(string(text)) == "STOP" {
	// 			//if we get the string STOP, let the client terminate
	// 			fmt.Println("TCP client exiting...")
	// 			break
	// 		}
	// 		_, err := c.Write([]byte(text + "\n"))
	// 		if err != nil {
	// 			//if the connection is closed, let the client terminate
	// 			fmt.Println(err)
	// 			break
	// 		}
	// 	}
	// }
	_, err = c.Write([]byte(*args.msg + "\n"))
	if err != nil {
		fmt.Println(err)
	}
	c.Close()
}
