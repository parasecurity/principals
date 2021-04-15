import socket
import argparse
import json


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
        send_obj = json.dumps(obj)

        print(send_obj)
        # We pass to be blocked the ip through data
        print("[+] Forwarding to {}:{}".format(server, port))
        sock.sendall(send_obj.encode("utf-8"))


def client(host="192.168.49.2", hostlocal="192.168.1.201", port=2378, portlocal=8080, action="block"):
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
                if not request:
                    break
                address = request.decode("utf-8")
                print("[+] Received", repr(address))

                argument = {}
                if action == "block":
                    argument["ip"] = address
                elif action == "unblock":
                    argument["ip"] = address
                elif action == "tarpit":
                    argument["ip"] = address

                obj = {
                    "action": action,
                    "argument": argument,
                }

                send_obj = json.dumps(obj)

                # We pass to be blocked the ip through data
                print("[+] Forwarding to {}:{}".format(host, port))
                sock_2.sendall(send_obj.encode("utf-8"))


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Flow controller")
    parser.add_argument("-l", "--listen", help="Ip address to listen incoming packets", required=True)
    parser.add_argument("-lp", "--listenport", help="Port to listen incoming packets", required=True)
    parser.add_argument("-s", "--send", help="Ip address of antrea-agent to send packet", required=True)
    parser.add_argument("-sp", "--sendport", help="Port of antrea-agent to send packet", required=True)
    parser.add_argument("-i", "--input", help="Input commands in JSON format", required=False)

    # Parse the arguments
    args = parser.parse_args()
    raw_data = args.input
    local_ip = args.listen
    local_port = int(args.listenport)
    remote_ip = args.send
    remote_port = int(args.sendport)

    if raw_data != None and raw_data != "":
        try:
            data = json.loads(raw_data)

            action = data["action"]
            argument = data["argument"]

            request(action, argument, remote_ip, remote_port)

        except ValueError:
            print("Decoding JSON has failed")
        except KeyError:
            print("Wrong command format")
    else:
        client(remote_ip, local_ip, remote_port, local_port)
