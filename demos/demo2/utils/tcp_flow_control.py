# -*- conding: utf-8 -*-

import socket

# address and port is arbitrary
def client(host="192.168.49.2", port=2378):
  with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
    sock.connect((host, port))

    while True:
      data = input("[+] Enter string : ")
      sock.sendall(data.encode('utf-8'))
      print("[+] Sending to {}:{}".format(host, port))

      response = sock.recv(4096)

      if not response:
        print("[-] Not Received")
        break

      print("[+] Received", repr(response.decode('utf-8')))

if __name__ == "__main__":
  server()
  client()
