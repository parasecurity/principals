import socket
import os

# address and port is arbitrary
def server(host='192.168.49.2', port=2378):
  # create socket
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        sock.bind((host, port))

        while True:

            print("[+] Listening on {0}:{1}".format(host, port))
            sock.listen(5)
            conn, addr = sock.accept()

            with conn as c:

                request = c.recv(4096)
                address = request.decode('utf-8')
                print("[+] Received", repr(address))
                os.system('ovs-ofctl add-flow br-int ip,nw_src=' + address + ',actions=drop')
                sock.listen(5)

if __name__ == "__main__":
    server()

