import json
import socket
import argparse

def request(send_ip, send_port):
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        print('[+] Querying logger data from {0}:{1}'.format(send_ip, send_port))
        sock.connect((send_ip, send_port))
        sock.sendall(b'Get logs')
        data = sock.recv(2048).decode('utf-8')  
        print(data)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Logger Client')
    parser.add_argument('-i','--input', help='Input necessary data in JSON format',required=True)

    # Parse the arguments
    args = parser.parse_args()
    raw_data = args.input

    # Parse the json data
    data = json.loads(raw_data, strict=False)

    send_ip = data['send_ip']
    send_port = int(data['send_port'])
    request(send_ip, send_port)

