import socket
import os
import json

def execute_command(action, argument):
    if action == 'block':
        os.system('ovs-ofctl add-flow br-int ip,nw_src=' + argument + ',actions=drop')
    elif action == 'unblock':
        os.system('ovs-ofctl del-flows --strict br-int ip,nw_src=' + argument)
    elif action == 'throttle':
        port = argument['port']
        # Maximum rate that a port should be allowed to send data
        limit = int(argument['limit']) * 1000
        # Maximum amount of data that a port can send beyond the policing rate
        barrier = int(argument['limit']) * 100
        
        os.system('ovs-vsctl set interface ' + port + ' ingress_policing_rate=' + str(limit))
        os.system('ovs-vsctl set interface ' + port + ' ingress_policing_burst=' + str(barrier))
    elif action == 'forward':
        malicious_ip=argument['malicious_ip']
        honeypot_ip=argument['honeypot_ip']
        honeypot_mac=argument['honeypot_mac']
        
        # Block all outgoing traffic to malicious domain
        os.system('ovs-ofctl add-flow br-int table=70,ip,nw_dst=' + malicious_ip + ',priority=300,actions=drop')
        # Forward all tcp:80 connections to malicious domain to honeypot
        os.system('ovs-ofctl add-flow br-int table=70,tcp,tcp_dst=80,nw_dst=' + malicious_ip + ',actions=mod_nw_dst:' + honeypot_ip + ',mod_dl_dst:' + honeypot_mac + ',goto_table:71')
        # Mask honeypot responces with original malicious ip
        os.system('ovs-ofctl add-flow br-int table=10,ip,dl_src=' + honeypot_mac + ',nw_src=' + honeypot_ip + ',actions=mod_nw_src:' + malicious_ip + ',goto_table:29')

def server(host='192.168.49.2', port=2378):
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        sock.bind((host, port))

        while True:
            print("[+] Listening on {0}:{1}".format(host, port))
            sock.listen(5)
            conn, addr = sock.accept()

            with conn as c:

                request = c.recv(4096)
                
                raw_data = request.decode('utf-8')
                data = json.loads(raw_data)
                print("[+] Received command", data['action'])
                execute_command(data['action'], data['argument'])

                sock.listen(5)

if __name__ == "__main__":
    server()


