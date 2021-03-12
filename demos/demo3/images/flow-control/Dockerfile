FROM centos:8

COPY forward.py /home/

RUN yum install -y python3

WORKDIR "/home"
CMD ["python3 forward.py"]
