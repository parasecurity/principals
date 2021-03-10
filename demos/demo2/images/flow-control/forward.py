import socket
import argparse

# host / port ~> antrea-agent
# host-local / port-local ~> flow-controller 
def client(host="192.168.49.2", hostlocal="192.168.1.201", port=2378, portlocal=8080):
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
    
        # Port to listen incomming tcp packets
        # This ip and port is for dga
        sock.bind((hostlocal, portlocal))
        print("[+] Listening on {0}:{1}".format(hostlocal, portlocal))
        sock.listen(5)
    
        # Permit to access
        conn, addr = sock.accept()
    
        # Ip to connect to send tcp packets
        # This ip and port is for antrea-agent
        sock_2 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock_2.connect((host, port))
    
        with conn as c:

          while True:

            # We receive the ip from dga
            request = c.recv(4096)
            address = request.decode('utf-8')
            print("[+] Received", repr(address))
        
            # We pass to be blocked the ip through data
            print("[+] Forwarding to {}:{}".format(host, port))
            sock_2.sendall(address.encode('utf-8'))

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Flow controller")
    parser.add_argument("-l", "--listen", help="Ip address to listen incoming packets")
    parser.add_argument("-s", "--send", help="Ip address of antrea-agent to send packet")
    args = parser.parse_args()
    
    local_ip = args.listen
    remote_ip = args.send
    client(remote_ip, local_ip)


