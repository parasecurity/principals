import socket
import argparse
import json

def create_send_obj(action, argument, malicious_ip):
    argument['malicious_ip'] = malicious_ip
    obj = {
        "action": action,
        "argument": argument,
    }

    return json.dumps(obj)
    

def request(action, argument, hostlocal="192.168.1.201", host="192.168.49.2", port=2378, portlocal=8080):
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:

        # Ip, port to listen for incoming tcp packets (dga)
        print("[+] Listening on {0}:{1}".format(hostlocal, portlocal))
        sock.bind((hostlocal, portlocal))
        sock.listen(5)
        conn, addr = sock.accept()
    
        # Ip, port to send tcp packets (antrea-agent)
        sock_2 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock_2.connect((host, port))

        with conn as c:
            while True:
                request = c.recv(4096)
                print("[+] Received ip address")
                malicious_ip = request.decode('utf-8')
                send_obj = create_send_obj(action, argument, malicious_ip) 

                print("[+] Forwarding to {}:{}".format(host, port))
                sock_2.sendall(send_obj.encode('utf-8'))

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Flow controller")
    parser.add_argument('-i','--input', help='Input necessary data in JSON format',required=True)

    # Parse the arguments
    args = parser.parse_args()
    raw_data = args.input

    # Parse the json data
    data = json.loads(raw_data, strict=False)

    action = data['action']
    argument = data['argument']
    listen_ip = data['listen_ip']
    send_ip = data['send_ip']

    request(action, argument, listen_ip, send_ip)

