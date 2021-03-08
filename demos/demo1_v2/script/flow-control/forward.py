import socket
import argparse

# host / port ~> antrea-agent
def client(block_ip, host="192.168.49.2", port=2378):
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
    
        # Ip to connect to send tcp packets
        # This ip and port is for antrea-agent
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.connect((host, port))
    
        # We pass to be blocked the ip through data
        print("[+] Forwarding to {}:{}".format(host, port))
        sock.sendall(block_ip.encode('utf-8'))

        sock.close()

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Flow controller")
    parser.add_argument("-b", "--block", help="Pods ip address to be blocked")
    parser.add_argument("-s", "--send", help="Ip address of antrea-agent to send packet")
    
    # Parse the arguments
    args = parser.parse_args()
    block_ip = args.block
    remote_ip = args.send
    
    client(block_ip, remote_ip)
