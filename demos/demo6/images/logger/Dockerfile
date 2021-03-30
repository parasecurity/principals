FROM centos:8

COPY logger.py logger_server.py /home/

RUN yum install -y python3

WORKDIR "/home"
CMD ["python3 logger.py"]
