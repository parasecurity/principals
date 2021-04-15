import socket
import argparse
import json

def create_send_obj(action, malicious_ip):
    argument = malicious_ip
    obj = {
        "action": action,
        "argument": argument,
    }

    return json.dumps(obj)

if __name__ == "__main__":
    sock_2 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock_2.connect(("localhost", 23456))

    while True:
        request = sock_2.recv(4096)
        print(type(request))
        if not request: break
        print("[+] Received ip address")
        print(request.decode('utf-8'))

