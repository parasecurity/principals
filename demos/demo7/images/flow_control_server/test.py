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
    sock_2.connect(("localhost", 12345))

    while True:
        val = input("Enter your value: ")
        send_obj = create_send_obj(val, "dsa")
        print(send_obj)
        sock_2.sendall(send_obj.encode('utf-8'))

