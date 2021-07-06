# Demo 1: DDoS Attack

This demo

We assume that a kubernetes cluster is setup and running, with antrea deployed on it.

When all prerequisites are satisfied, you can start the demo with:

```sh
./run
```

For the _DDoS with application level canary_ demo, pass the following arguments:
```sh
create canary -api=<API ip>:8001 -c=tarpit
```

For the _DDoS with link level canary demo_, pass the following arguments:
```sh
create canary-link -api=<API ip>:8001 -c=tarpit
```

