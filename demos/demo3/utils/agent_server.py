import socket
import os
import json

def execute_command(action, argument):
    if action == 'block':
        os.system('ovs-ofctl add-flow br-int ip,nw_src=' + argument + ',actions=drop')
    elif action == 'unblock':
        os.system('ovs-ofctl del-flows --strict br-int ip,nw_src=' + argument)
    elif action == 'throttle':
        port = argument['port']
        # Maximum rate that a port should be allowed to send data
        limit = int(argument['limit']) * 1000
        # Maximum amount of data that a port can send beyond the policing rate
        barrier = int(argument['limit']) * 100

        os.system('ovs-vsctl set interface ' + port + ' ingress_policing_rate=' + str(limit))
        os.system('ovs-vsctl set interface ' + port + ' ingress_policing_burst=' + str(barrier))

def server(host='192.168.49.2', port=2378):
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        sock.bind((host, port))

        while True:
            print("[+] Listening on {0}:{1}".format(host, port))
            sock.listen(5)
            conn, addr = sock.accept()

            with conn as c:

                request = c.recv(4096)
                
                raw_data = request.decode('utf-8')
                data = json.loads(raw_data)
                print("[+] Received command", data['action'])
                execute_command(data['action'], data['argument'])

                sock.listen(5)

if __name__ == "__main__":
    server()


