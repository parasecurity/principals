from scapy.all import *
import socket
import argparse
import time

class DomainList:
    domains = None

    def exist(self, domain):
        if domain in self.domains:
            return True
        else:
            return False
        pass

    def load_domains(self, path):
        self.domains = set(line.strip() for line in open(path))
        pass

    def __init__(self, path):
        self.load_domains(path)

class PacketMonitor:
    iface = None
    detector = None

    def establish_connection(self, host="192.168.1.204", port=8080):
        self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.socket.connect((host, port))
        pass

    def send_ip(self, address):
        self.socket.sendall(address.encode('utf-8'))
        pass

    def trim_domain_string(self, domain):
        domain = domain[:-2]
        domain = domain[2:]
        return domain

    def process_packet(self, packet):
        """
        This function is executed whenever a packet is sniffed
        """
        if IP not in packet:
            return

        if not packet.haslayer(DNS):
            return

        domains = None
        dns_layer = packet.getlayer(DNS)

        if dns_layer.ancount > 0 and dns_layer.qd:
            ts = time.time()
            ip_src = str(packet[IP].src)
            ip_dst = str(packet[IP].dst)
            if UDP in packet:
                port_src=packet[UDP].sport
                port_dst=packet[UDP].dport
            if TCP in packet:
                port_src=packet[TCP].sport
                port_dst=packet[TCP].dport


            domain = self.trim_domain_string(str(dns_layer.qd.qname))

            dga_result = self.domains.exist(domain)
            print(ip_src, ip_dst, domain, dga_result)

            if dga_result == False:
                return

            for x in range(dns_layer.ancount):
                resolved_ip = str(dns_layer.an[x].rdata)
                print("Blocking: " + resolved_ip)
                #self.send_ip(resolved_ip)

    def sniff_packets(self):
        if iface:
            # `process_packet` is the callback
            sniff(filter="port 53", prn=self.process_packet, iface=iface, store=False)
        else:
            # sniff with default interface
            sniff(filter="port 53", prn=self.process_packet, store=False)

    def __init__(self, iface, domains, address):
        self.iface = iface
        self.domains = domains
        #self.establish_connection(address)
        pass

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="DGA detector")
    parser.add_argument("-i", "--iface", help="Interface to use, default is scapy's default interface")
    parser.add_argument("-d", "--domains", help="Domain list to track")
    parser.add_argument("-a", "--address", help="Ip address of flow controller")
    args = parser.parse_args()

    domains_file = args.domains
    domains = DomainList(domains_file)

    iface = args.iface
    address = args.address
    packet_monitor = PacketMonitor(iface, domains, address)
    packet_monitor.sniff_packets()
    packet_monitor.socket.close()
