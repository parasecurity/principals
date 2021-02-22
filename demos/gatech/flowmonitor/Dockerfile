FROM python:3.7.2

COPY flow_monitor.py /flow_monitor.py
COPY flow_monitor.sh /flow_monitor.sh

RUN chmod +x /flow_monitor.sh
RUN apt update -y && apt install -y tcpdump
RUN pip install requests

ENTRYPOINT ["/flow_monitor.sh"]
