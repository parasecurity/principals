# DNS Flow Stitching

For stitching DNS resolutions with flows. Currently DNS flow stitching only works with IPv4 traffic. IPv6 flows will be marked as unresolved.
  
You can manually clone the repository and build the binaries yourself.

    $ cd dns_stitching
    $ go build 

## Usage

DNS flow stitching can work with PCAP as well as a live interface.

### Parsing PCAP File
    
       $./dns_stitching pcap [file...]

As shown above, the `pcap` command takes an argument which is the path to an uncompressed PCAP file. 

### Parsing Live Interface

       $./dns_stitching live [command options] [interface]
    
    OPTIONS:
       --promiscuous    set promiscuous mode for traffic collection
       --timeout value  set timeout value for traffic collection (default: 30)
       
As shown above, the `live` command accepts the name of an interface to parse
from as its sole argument. There are also a number of command specific flags
that can be set.

The following shows an example of how you can parse from `eth0` in promiscuous
mode.

    $ ./dns_stitching live --promiscuous eth0

## Docker commands
To build docker: `docker build -t test_dns_stitching .`
To run rocker: `docker run -i -t --rm -p 8000:8000 test_dns_stitching`
