from scapy.all import *
import socket
import argparse
import time
import json

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


class PacketInfo(object):

    def create_send_object(self):
        obj = {
            "ts": self.ts,
            "domain": self.domain,
            "ip_src": self.ip_src,
            "ip_dst": self.ip_dst,
            "resolved_ip": self.resolved_ip,
            "mac_src": self.mac_src,
            "mac_dst": self.mac_dst,
            "port_src": self.port_src,
            "port_dst": self.port_dst
        }
        return json.dumps(obj)

    def __init__(self, ts, domain, ip_src, ip_dst, resolved_ip, mac_src, mac_dst, port_src, port_dst):
        self.ts = ts
        self.domain = domain
        self.ip_src = ip_src
        self.ip_dst = ip_dst
        self.resolved_ip = resolved_ip
        self.mac_src = mac_src
        self.mac_dst = mac_dst
        self.port_src = port_src
        self.port_dst = port_dst

class PacketMonitor:
    iface = None
    detector = None

    def establish_connection(self, host="192.168.1.204", port=8080):
        self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.socket.connect((host, port))
        pass

    def send_data(self, data):
        self.arguments["data"] = data
        obj = {
            "action": self.action,
            "argument": self.arguments,
        }
        # Format json object
        send_obj = json.dumps(obj) + "\n"
        self.socket.sendall(send_obj.encode("utf-8"))
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
            mac_src = str(packet.src)
            mac_dst = str(packet.dst)
            if UDP in packet:
                port_src = str(packet[UDP].sport)
                port_dst = str(packet[UDP].dport)
            if TCP in packet:
                port_src = str(packet[TCP].sport)
                port_dst = str(packet[TCP].dport)

            domain = self.trim_domain_string(str(dns_layer.qd.qname))
            dga_result = self.domains.exist(domain)

            if dga_result == False:
                return

            for x in range(dns_layer.ancount):
                resolved_ip = str(dns_layer.an[x].rdata)
                packet_info = PacketInfo(ts, domain, ip_src, ip_dst, resolved_ip, mac_src, mac_dst, port_src, port_dst)
                send_object = packet_info.create_send_object()
                self.send_data(send_object)

    def sniff_packets(self):
        if iface:
            # `process_packet` is the callback
            sniff(filter="port 53", prn=self.process_packet, iface=iface, store=False)
        else:
            # sniff with default interface
            sniff(filter="port 53", prn=self.process_packet, store=False)

    def __init__(self, iface, domains, address, port, action, arguments):
        self.iface = iface
        self.domains = domains
        self.establish_connection(address, port)
        self.action = action
        self.arguments = arguments
        pass

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="DGA detector")
    parser.add_argument("-i", "--iface", help="Interface to use, default is scapy's default interface")
    parser.add_argument("-d", "--domains", help="Domain list to track")
    parser.add_argument("-a", "--address", help="Ip address of flow controller")
    parser.add_argument("-p", "--port", help="Port of flow controller")
    parser.add_argument("-c", "--command", help="The command we want to repeat to the ovs bridge to execute", required=True)
    parser.add_argument("-arg", "--arguments", help="Input arguments in JSON format", required=False)
    args = parser.parse_args()

    domains_file = args.domains
    domains = DomainList(domains_file)

    iface = args.iface
    address = args.address
    port = int(args.port)

    action = args.command
    if args.arguments == None:
        arguments_json = "\{\}"
    else:
        arguments_json = args.arguments

    try:
        arguments = json.loads(arguments_json)
    except ValueError:
        print("Decoding JSON has failed")
        arguments = {}
        pass

    packet_monitor = PacketMonitor(iface, domains, address, port, action, arguments)
    packet_monitor.sniff_packets()
    packet_monitor.socket.close()
