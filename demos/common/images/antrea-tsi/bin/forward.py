import socket
import argparse
import json


class Buffer:
    def __init__(self, sock):
        self.sock = sock
        self.buffer = b""

    def get_line(self):
        while b"\n" not in self.buffer:
            data = self.sock.recv(1024)
            if not data:  # socket closed
                return None
            self.buffer += data
        line, sep, self.buffer = self.buffer.partition(b"\n")
        return line.decode()


def request(action, argument, server="192.168.49.2", port=2378):
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:

        # Ip to connect to send tcp packets
        # This ip and port is for antrea-agent
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.connect((server, port))

        # Create object to send
        obj = {
            "action": action,
            "argument": argument,
        }
        send_obj = json.dumps(obj) + "\n"

        print(send_obj)
        # We pass to be blocked the ip through data
        print("[+] Forwarding to {}:{}".format(server, port))
        sock.sendall(send_obj.encode("utf-8"))


def client(host="192.168.49.2", hostlocal="192.168.1.201", port=2378, portlocal=8080, action="block", argument={}):
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:

        # Ip to connect to send tcp packets
        # This ip and port is for antrea-agent
        sock_2 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock_2.connect((host, port))

        # Port to listen incomming tcp packets
        # This ip and port is for dga
        sock.bind((hostlocal, portlocal))
        print("[+] Listening on {0}:{1}".format(hostlocal, portlocal))
        sock.listen(5)

        # Permit to access
        conn, addr = sock.accept()

        with conn as c:
            b = Buffer(c)

            while True:

                # We receive the ip from dga
                address = b.get_line()
                if address is None:
                    break

                print("[+] Received", repr(address))

                argument["ip"] = address

                obj = {
                    "action": action,
                    "argument": argument,
                }

                send_obj = json.dumps(obj) + "\n"

                # We pass to be blocked the ip through data
                print("[+] Forwarding to {}:{}".format(host, port))
                sock_2.sendall(send_obj.encode("utf-8"))


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Flow controller")
    parser.add_argument("-l", "--listen", help="Ip address to listen incoming packets", required=False)
    parser.add_argument("-lp", "--listenport", help="Port to listen incoming packets", required=False)
    parser.add_argument("-s", "--send", help="Ip address of antrea-agent to send packet", required=True)
    parser.add_argument("-sp", "--sendport", help="Port of antrea-agent to send packet", required=True)
    parser.add_argument("-a", "--action", help="The action we want to repeat to the ovs bridge", required=True)
    parser.add_argument("-i", "--arguments", help="Input arguments in JSON format", required=False)

    # Parse the arguments
    args = parser.parse_args()

    local_ip = args.listen
    if args.listenport == None:
        local_port = int(0)
    else:
        local_port = int(args.listenport)

    remote_ip = args.send
    remote_port = int(args.sendport)

    action = args.action
    if args.arguments == None:
        arguments_json = "\{\}"
    else:
        arguments_json = args.arguments

    try:
        arguments = json.loads(arguments_json)
    except ValueError:
        print("Decoding JSON has failed")
        arguments = {}
        pass

    if local_ip == None:
        request(action, arguments, remote_ip, remote_port)
    else:
        while True:
            try:
                client(remote_ip, local_ip, remote_port, local_port, action, arguments)
            except: 
                pass

