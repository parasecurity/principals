import socket

def receive_udp_data(host='0.0.0.0', port=9999):
    # Create a UDP socket
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    # Bind the socket to the address and port
    server_address = (host, port)
    print(f'Starting UDP server on {server_address}')
    sock.bind(server_address)

    try:
        while True:
            # Receive data
            data, address = sock.recvfrom(4096)
            print(f'Received {len(data)} bytes from {address}')
            print(data.decode('utf-8', errors='replace'))
    except KeyboardInterrupt:
        print('Server stopped by user')
    finally:
        sock.close()

if __name__ == '__main__':
    receive_udp_data()
