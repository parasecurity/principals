from scapy.all import *
import socket
import argparse
import time
import tensorflow as tf
import numpy as np
from tensorflow.python.keras.preprocessing import sequence
from tensorflow import keras
from typing import List, Tuple


def as_keras_metric(method):
    """ from https://stackoverflow.com/questions/43076609/how-to-calculate-precision-and-recall-in-keras """
    import functools

    @functools.wraps(method)
    def wrapper(self, args, **kwargs):
        """ Wrapper for turning tensorflow metrics into keras metrics """
        value, update_op = method(self, args, **kwargs)
        tf.compat.v1.keras.backend.get_session().run(tf.compat.v1.local_variables_initializer())
        with tf.control_dependencies([update_op]):
            value = tf.identity(value)
        return value

    return wrapper

class DGADetector:
    domain_name_dictionary = {'0': 0, '1': 1, '2': 2, '3': 3, '4': 4, '5': 5, '6': 6, '7': 7, '8': 8, '9': 9, ':': 10,
                          '-': 11, '.': 12, '/': 13, '_': 14, 'a': 15, 'b': 16, 'c': 17, 'd': 18, 'e': 19, 'f': 20,
                          'g': 21, 'h': 22, 'i': 23, 'j': 24, 'k': 25, 'l': 26, 'm': 27, 'n': 28, 'o': 29, 'p': 30,
                          'q': 31, 'r': 32, 's': 33, 't': 34, 'u': 35, 'v': 36, 'w': 37, 'x': 38, 'y': 39, 'z': 40,
                          np.NaN: 41}

    model = None

    def domain_to_ints(self, domain: str) -> List[int]:
        """
        Converts the given domain into a list of ints, given the static dictionary defined above.
        Converts the domain to lower case, and uses a set value (mapped to np.NaN) for unknown characters.
        """
        return [
             self.domain_name_dictionary.get(y, self.domain_name_dictionary.get(np.NaN))
             for y in domain.lower()
        ]

    def prep_data(self, data: np.ndarray, max_length=75) -> np.ndarray:
        return sequence.pad_sequences(
            np.array([self.domain_to_ints(x) for x in data]), maxlen=max_length)

    def load_model(self):
        if self.model:
            return

        self.model = keras.models.load_model(self.model_path,  custom_objects={
            "precision": as_keras_metric(tf.metrics.precision),
            "recall": as_keras_metric(tf.metrics.recall)
        })  

        self.graph = tf.compat.v1.get_default_graph()

    def is_dga(self, domain):
        real_x = self.prep_data([domain], 75)
        with self.graph.as_default():
            prediction = self.model.predict(real_x, batch_size=256, verbose=0)
            return prediction[0][0] > 0.5

    def __init__(self, model_path):
        self.model_path = model_path
        self.load_model()

class PacketMonitor:
    iface = None
    detector = None

    def establish_connection(self, host="192.168.1.204", port=8080):
        # Ip and port of flow-controller
        self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM) 
        self.socket.connect((host, port))        
        pass

    def send_ip(self, address):
        # Send bad ip address
        self.socket.sendall(address.encode('utf-8'))
        pass

    def process_packet(self, packet):
        """
        This function is executed whenever a packet is sniffed
        """
        if IP not in packet:
            return
        
        if not packet.haslayer(DNS):
            return

        dns_layer = packet.getlayer(DNS)
        
        if dns_layer.ancount > 0 and dns_layer.qd:
            ip_src = str(packet[IP].src)
            ip_dst = str(packet[IP].dst)
            domain = str(dns_layer.qd.qname)
            
            if str(ip_src) != "172.16.0.1" and str(ip_src) != "172.16.0.2":
                dga_result = self.detector.is_dga(domain)
                if (domain == "b'speedtest.wdc01.softlayer.com.'"):
                    dga_result = True

                print(ip_src, ip_dst, domain, dga_result)

                if dga_result == False:
                    return

                for x in range(dns_layer.ancount):
                    resolved_ip = str(dns_layer.an[x].rdata)
                    print("Blocking: " + resolved_ip)
                    self.send_ip(resolved_ip)
    
    def sniff_packets(self):
        if iface:
            # `process_packet` is the callback
            sniff(filter="port 53", prn=self.process_packet, iface=iface, store=False)
        else:
            # sniff with default interface
            sniff(filter="port 53", prn=self.process_packet, store=False)

    def __init__(self, iface, detector, address):
        self.iface = iface
        self.detector = detector
        self.establish_connection(address)
        pass

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="DGA detector")
    parser.add_argument("-i", "--iface", help="Interface to use, default is scapy's default interface")
    parser.add_argument("-m", "--model", help="DGA detection model to load")
    parser.add_argument("-a", "--address", help="Ip address of flow controller")
    args = parser.parse_args()

    model_path = args.model
    detector = DGADetector(model_path)

    iface = args.iface
    address = args.address
    packet_monitor = PacketMonitor(iface, detector, address)
    packet_monitor.sniff_packets()
    packet_monitor.socket.close()
