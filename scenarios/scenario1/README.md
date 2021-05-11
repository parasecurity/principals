## Scenario 1: DDoS (application-based)

### Requirements

- A dummy Web service 
- X pods that generate legitimate traffic to a Web service, Y pods that generate DDoS traffic, Z nodes
- A detector pod that receives the mirrored traffic and records aggregate statistics (vol in/ vol out) for a destination IP address (the one of the Web service) and measures the uplink of the service and also used for mitigation. The same pod will be used to trigger the detection rules and block traffic. 
- A ‘canary’ client pod periodically health-checks  the Web-service by requesting a specific health-check endpoint (application-specific measurement). The same client can be used to ‘ping’ and record RTT times (non-application specific measurement)

### Web service

In order to create a running web service just:
```sh
cd web_service
docker run -dit --name sample-service -p 8080:80 -v "$PWD":/usr/local/apache2/htdocs/ httpd:2.4
```
Your webpage should be running on `http://localhost:8080/webpage.html`

