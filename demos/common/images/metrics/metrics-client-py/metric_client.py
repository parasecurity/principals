#!/bin/python3
# main is a tester
# TODO main should be single session log

import socket
import argparse
from time import time

class delay_log(object):
    @classmethod
    def init_logger(cls, name, server_ip='10.104.60.10'):
        cls.id = name
        cls.conn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        cls.conn.connect((server_ip, 4321)) 

    @classmethod
    def stamp(cls, msg):
        t = round(time()*1000000000)
        b = t.to_bytes(8, byteorder='little')
        buff = b + bytes(cls.id, 'ascii') + bytes(' ', 'ascii') + bytes(msg, 'ascii') + bytes('\n', 'ascii')
        cls.conn.send(buff)

def tester(server_ip):
    log = delay_log("python_tester", server_ip)
    log.stamp("action1")
    log.stamp("action2")
    log.stamp("action3")
    log.conn.close()

if __name__ == '__main__' :
    parser = argparse.ArgumentParser(description="Metrics client tester (python)")
    parser.add_argument("-a", "--server_ip", help="Ip address of metrics server to send logs", required=False)
    parser.add_argument("-n", "--log_name", help="Name of log entry", required=False)

    args = parser.parse_args()

    if args.server_ip == None:
        server_ip = '10.104.60.10'
    else:
        server_ip = args.server_ip

    if args.log_name == None:
        log_name = 'logger'
    else:
        server_ip = args.log_name
