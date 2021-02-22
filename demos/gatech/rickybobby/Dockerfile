FROM ubuntu:18.04

RUN apt update -y && apt install -y libpcap-dev

COPY rickybobby /bin/rickybobby
RUN chmod +x /bin/rickybobby

ENTRYPOINT ["/bin/rickybobby", "--questions", "live", "eth0"]
