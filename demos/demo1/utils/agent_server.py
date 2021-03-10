import socket
import os
import json

def execute_command(action, argument):
    if action == 'block':
        os.system('ovs-ofctl add-flow br-int ip,nw_src=' + argument + ',actions=drop')
    elif action == 'unblock':
        os.system('ovs-ofctl del-flows --strict br-int ip,nw_src=' + argument)

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

                # Receive a json array from flow-controller
                raw_data = request.decode('utf-8')
                data = json.loads(raw_data)

                print("[+] Received command", data['action'])
                execute_command(data['action'], data['argument'])

                sock.listen(5)

if __name__ == "__main__":
    server()


