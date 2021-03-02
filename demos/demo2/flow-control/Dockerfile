FROM centos:8

COPY forward.py /home/

RUN yum -y install python3-pip python3-devel

WORKDIR "/home"
CMD ["python3 forward.py"]
