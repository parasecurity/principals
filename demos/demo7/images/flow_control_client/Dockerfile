FROM centos:8

COPY flow_client.py /home/

RUN yum install -y python3

WORKDIR "/home"
CMD ["python3 forward.py"]
