import socket
import fcntl
import struct
import sys

import requests


# https://stackoverflow.com/questions/24196932/how-can-i-get-the-ip-address-from-nic-in-python
def get_ip_address(ifname):
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    return socket.inet_ntoa(fcntl.ioctl(
        s.fileno(),
        0x8915,  # SIOCGIFADDR
        struct.pack(b'256s', ifname[:15].encode('utf-8'))
    )[20:24])


if __name__ == '__main__':
    client = get_ip_address('eth1')
    print('Starting dns monitor for {}'.format(client))
    MANAGEMENT_URL = 'http://10.0.0.8:5000/ips'
    for line in sys.stdin:
        line = line.split()[4]
        if '.' in line:
            dst_ip = '.'.join(line.split('.')[:4])
            data = {'client': client, 'dst_ip': dst_ip}
            print(data)
            r = requests.post(MANAGEMENT_URL, json=data)
            if r.status_code != 200:
                print(f'POST request returned error code: {r.status_code}')

