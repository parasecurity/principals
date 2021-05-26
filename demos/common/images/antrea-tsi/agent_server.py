import socket
import os
import json
import argparse
import logging


def execute_command(action, argument):
    try:
        if action == "block":
            malicious_ip = argument["ip"]
            comand_1 = "ovs-ofctl add-flow br-int ip,nw_dst=" + malicious_ip + ",actions=drop"
            comand_2 = "ovs-ofctl add-flow br-int ip,nw_src=" + malicious_ip + ",actions=drop"
            os.system(comand_1)
            os.system(comand_2)
            logging.info("[+] Executing {0}".format(comand_1))
            logging.info("[+] Executing {0}".format(comand_2))
        elif action == "unblock":
            ip = argument["ip"]
            comand_1 = "ovs-ofctl del-flows --strict br-int ip,nw_src=" + ip
            comand_2 = "ovs-ofctl del-flows --strict br-int ip,nw_dst=" + ip
            os.system(comand_1)
            os.system(comand_2)
            logging.info("[+] Executing {0}".format(comand_1))
            logging.info("[+] Executing {0}".format(comand_2))
        elif action == "throttle":
            port = argument["port"]
            # Maximum rate that a port should be allowed to send data
            limit = int(argument["limit"]) * 1000
            # Maximum amount of data that a port can send beyond the policing rate
            barrier = int(argument["limit"]) * 100
            comand_1 = "ovs-vsctl set interface " + port + " ingress_policing_rate=" + str(limit)
            comand_2 = "ovs-vsctl set interface " + port + " ingress_policing_burst=" + str(barrier)
            os.system(comand_1)
            os.system(comand_2)
            logging.info("[+] Executing {0}".format(comand_1))
            logging.info("[+] Executing {0}".format(comand_2))
        elif action == "forward":
            malicious_ip = argument["ip"]
            honeypot_ip = argument["honeypot_ip"]
            honeypot_mac = argument["honeypot_mac"]

            # Block all outgoing traffic to malicious domain
            comand_1 = "ovs-ofctl add-flow br-int table=70,ip,nw_dst=" + malicious_ip + ",priority=300,actions=drop"
            # Forward all tcp:80 connections to malicious domain to honeypot
            comand_2 = (
                "ovs-ofctl add-flow br-int table=70,tcp,tcp_dst=80,nw_dst="
                + malicious_ip
                + ",actions=mod_nw_dst:"
                + honeypot_ip
                + ",mod_dl_dst:"
                + honeypot_mac
                + ",goto_table:71"
            )
            # Mask honeypot responces with original malicious ip
            comand_3 = (
                "ovs-ofctl add-flow br-int table=10,ip,dl_src="
                + honeypot_mac
                + ",nw_src="
                + honeypot_ip
                + ",actions=mod_nw_src:"
                + malicious_ip
                + ",goto_table:29"
            )
            os.system(comand_1)
            os.system(comand_2)
            os.system(comand_3)
            logging.info("[+] Executing {0}".format(comand_1))
            logging.info("[+] Executing {0}".format(comand_2))
            logging.info("[+] Executing {0}".format(comand_3))
        elif action == "tarpit":
            malicious_ip = argument["ip"]
            comand_1 = "ovs-ofctl add-flow br-int ip,nw_dst=" + malicious_ip + ",action=set_queue:100,goto_table:10"
            comand_2 = "ovs-ofctl add-flow br-int ip,nw_src=" + malicious_ip + ",action=set_queue:100,goto_table:10"
            os.system(comand_1)
            os.system(comand_2)
            logging.info("[+] Executing {0}".format(comand_1))
            logging.info("[+] Executing {0}".format(comand_2))
        elif action == "log":
            logging.info(argument)

    except KeyError:
        logging.warning("Wrong Arguments")
        pass


def server(host="192.168.49.2", port=2378, listen=True):
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        if listen == True:
            sock.bind((host, port))
            sock.listen(5)
        else:
            sock.connect((host, port))
            conn = sock

        logging.info("[+] Listening on {0}:{1}".format(host, port))

        while True:
            if listen == True:
                conn, addr = sock.accept()
            
            with conn as c:
                while True:
                    request = c.recv(4096)
                    if not request:
                        break
                    json_data = request.decode("utf-8")

                    try:
                        data = json.loads(json_data)
                        action = data["action"]
                        argument = data["argument"]

                        logging.info("[+] Received command {0}".format(action))

                        if action == "exit":
                            logging.info("Close the program")
                            response = "exit"
                            c.sendall(response.encode("utf-8"))
                            break

                        execute_command(action, argument)

                        response = "ok"
                        c.sendall(response.encode("utf-8"))

                    except ValueError:
                        logging.warning("Decoding JSON has failed")
                        pass
                    except KeyError:
                        logging.warning("Wrong command format")
                        pass


if __name__ == "__main__":
    logging.basicConfig(filename="agent-server.log", level=logging.INFO)

    parser = argparse.ArgumentParser(description="Agent Server")
    parser.add_argument("-i", "--ip", help="Server ip address", required=True)
    parser.add_argument("-p", "--port", help="Server port address", required=True)
    parser.add_argument("-l", "--listen", help="Server port address", required=True, type=lambda x: (str(x).lower() == 'true'))

    args = parser.parse_args()
    server_ip = args.ip
    server_port = int(args.port)
    server_listen = bool(args.listen)

    server(server_ip, server_port, server_listen)
