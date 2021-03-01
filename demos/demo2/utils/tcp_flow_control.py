import socket

# address and port is arbitrary
def client(host="192.168.49.2", port=2378):
  with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
    sock.connect((host, port))

    while True:

      # We pass to be blocked the ip through data
      data = input("[+] Enter string : ")
      sock.sendall(data.encode('utf-8'))
      print("[+] Sending to {}:{}".format(host, port))

      response = sock.recv(4096)
      if response.decode('utf-8') == "exit":
        print("Close the program")
        break

if __name__ == "__main__":
  client()

