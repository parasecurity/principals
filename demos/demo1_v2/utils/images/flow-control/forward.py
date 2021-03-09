import socket
import argparse
import json

def request(action, argument, server = "192.168.49.2", port = 2378):
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

        # We pass to be blocked the ip through data
        print("[+] Forwarding to {}:{}".format(server, port))
        sock.sendall(send_obj.encode('utf-8'))

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Flow controller")
    parser.add_argument('-i','--input', help='Input commands in JSON format',required=True)

    # Parse the arguments
    args = parser.parse_args()
    raw_data = args.input

    # Parse the json data
    data = json.loads(raw_data)

    action = data['action']
    argument = data['argument']
    server_ip = data['server_ip']

    request(action, argument, server_ip)

