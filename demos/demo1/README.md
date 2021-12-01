# Demo 1: DDoS Attack

This demo

We assume that a kubernetes cluster is setup and running, with antrea deployed on it.


Before running you need to configure the demo.
First edit conf/demo.conf. Set restistry's ip and port, and make sure images are in correct version.
Then run
```sh
./configure.sh
```
 
When all prerequisites are satisfied, you can start the demo with:

```sh
./run
```

For the _DDoS with application level canary_ demo, pass the following arguments:
```sh
# This command creates the canary
{\"action\":\"create\", \"target\": \"canary\", \"arguments\": []}

# This command creates the detector
{\"action\":\"create\", \"target\": \"detector\", \"arguments\": [\"-c=block\"]}

```

For the _DDoS with link level canary demo_, pass the following arguments:
```sh
# This command creates the canary-link
{\"action\":\"create\", \"target\": \"canary-link\", \"arguments\": []}

# This command creates the detector-link
{\"action\":\"create\", \"target\": \"detector-link\", \"arguments\": [\"-c=block\"]}

```

For more information run:
```sh
./run --help
```
