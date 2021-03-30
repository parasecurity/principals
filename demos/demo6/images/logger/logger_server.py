import socket

class Log:
    log_file = None
    
    def get_log(self):
        return self.log_file.read()

    def __init__(self):
        self.log_file = open('data.log', 'r')
        
def client(log):
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        print('[+] Listening on {0}:{1}'.format('localhost', 8006))
        sock.bind(('localhost', 8006))
        
        while True:
            sock.listen()
            conn, addr = sock.accept()
            with conn as c:
                c.recv(4096)
                data = log.get_log()
                print('[+] Sending log data')
                c.sendall(data.encode('utf-8'))
                sock.listen()
                log = Log()

if __name__ == '__main__':
    log = Log()
    client(log)
