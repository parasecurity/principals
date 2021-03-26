import socket
import argparse
import json
import logging

def request(hostlocal="192.168.1.201", portlocal=8080):
    logging.basicConfig(filename='data.log' ,level=logging.INFO, format='%(message)s')
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:

        # Ip, port to listen for incoming tcp packets (dga)
        print("[+] Listening on {0}:{1}".format(hostlocal, portlocal))
        sock.bind((hostlocal, portlocal))
        sock.listen(5)
        conn, addr = sock.accept()
    
        with conn as c:
            while True:
                request = c.recv(4096)
                print("[+] Received ip address")
                raw_data = request.decode('utf-8')
                try:
                    data = json.loads(raw_data, strict=False)
                    logging.info(raw_data)
                except ValueError:
                    print("Decoding JSON has failed")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Flow controller")
    parser.add_argument('-i','--input', help='Input necessary data in JSON format',required=True)

    # Parse the arguments
    args = parser.parse_args()
    raw_data = args.input

    # Parse the json data
    data = json.loads(raw_data, strict=False)
    listen_ip = data['listen_ip']
    request(listen_ip)