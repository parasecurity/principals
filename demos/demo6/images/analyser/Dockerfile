FROM centos:8

COPY domains.txt monitor.py requirements.txt /tmp/

RUN yum -y install python3-pip python3-devel
RUN yum -y install gcc gcc-c++ 
RUN pip3 install -r /tmp/requirements.txt

WORKDIR "/tmp"
CMD ["python3 monitory.py -d domains.txt"]
