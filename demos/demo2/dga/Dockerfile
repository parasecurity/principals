FROM centos:8

COPY dga.model monitor.py requirements.txt /tmp/
RUN yum -y install python3-pip python3-devel
RUN yum -y install gcc gcc-c++ 
RUN pip3 install -r /tmp/requirements.txt

CMD ["python3 /tmp/monitory.py -m /tmp/dga.model"]
