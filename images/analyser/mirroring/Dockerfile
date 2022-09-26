FROM projects.registry.vmware.com/antrea/antrea-ubuntu:v1.0.0

ARG DEBIAN_FRONTEND=noninteractive
RUN  apt-get clean
RUN  apt-get -y update
RUN  apt-get -y upgrade
COPY mirroring.js /home/mirroring.js
RUN chmod +x /home/mirroring.js
RUN apt-get install -y npm
RUN  apt-get install -y nodejs
RUN  /usr/share/openvswitch/scripts/ovs-ctl start
CMD ["/bin/sh" ,"-c" ,"node /home/mirroring.js"]
