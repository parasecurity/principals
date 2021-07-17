# Demo 1: DDoS Attack

This demo

We assume that a kubernetes cluster is setup and running, with antrea deployed on it.

When all prerequisites are satisfied, you can start the demo with:

```sh
./run
```

For the _DDoS with application level canary_ demo, pass the following arguments:
```sh
# This command creates the canary
./client -arg "{\"action\":\"create\", \"target\": \"canary\", \"arguments\": []}"

# This command creates the detector
./client -arg "{\"action\":\"create\", \"target\": \"detector\", \"arguments\": [\"-c=tarpit\"]}"

```

For the _DDoS with link level canary demo_, pass the following arguments:
```sh
# This command creates the canary-link
./client -arg "{\"action\":\"create\", \"target\": \"canary-link\", \"arguments\": []}"

# This command creates the detector-link
./client -arg "{\"action\":\"create\", \"target\": \"detector-link\", \"arguments\": [\"-c=tarpit\"]}"

```

