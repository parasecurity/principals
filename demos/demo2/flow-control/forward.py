import socket

# host / port ~> antrea-agent
# host-local / port-local ~> flow-controller 
def client(host="192.168.49.2", port=2378, hostlocal="192.168.1.201", portlocal=8080):
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
    sock.connect((host, port))
    
    with conn as c:

      while True:

        # We receive the ip from dga
        request = c.recv(4096)
        address = request.decode('utf-8')
        print("[+] Received", repr(address))
        
        # We pass to be blocked the ip through data
        print("[+] Forwarding to {}:{}".format(host, port))
        sock.sendall(address.encode('utf-8'))

if __name__ == "__main__":
  client()

