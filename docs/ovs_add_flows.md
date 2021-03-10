# Quick guide to ovs-ofctl add-flow command

One of the most usefull commands provided by OVS is __ovs-ofctl add-flow__.  In each case, flow specifies a flow entry in the format described in Flow Syntax, below, file is a text file that contains zero or more flows in the same syntax, one per line.

- add-flow \<_switch_\> \<_flow_\>
- add-flow \<_switch_> - < \<_file_\>

## Flow Syntax

Add-flow command accept an argument that describes a flow or flows. Such flow descriptions comprise a series of __field=value__ assignements, seperated by commas or white space.

## Example of Flow commands

> Block all outgoing traffic from a given IP address
```sh
ovs-ofctl add-flow \<name of bridge\> ip,nw_src=\<ip\>,actions=drop
```

> Unblock all outgoing traffic from a given IP address
```sh
ovs-ofctl del-flows --strict \<name of bridge\> ip,nw_src=\<ip\>
```

> Block all incoming traffic to a given IP address
```sh
ovs-ofctl add-flow \<name of bridge\> ip,nw_dst=\<ip\>,actions=drop
```

> Unblock all incoming traffic to a given IP address
```sh
ovs-ofctl del-flows --strict \<name of bridge\> ip,nw_dst=\<ip\>
```

> Block all outgoing/incoming traffic from an specific bridge port
```sh
ovs-ofctl add-flow \<name of bridge\> in_port=\<bridge port\>,actions=drop
``

> Unblock all outgoing/incoming traffic from an specific bridge port
```sh
ovs-ofctl del-flows --strict \<name of bridge\> in_port=\<bridge port\>,actions=drop
``

> Block all outgoing traffic by matching field dl\_src with an Ethernet source address. This value uses 6 pairs of hexadecimal digits to specify, eg: 00:0B:C4:A8:22:B0.
```sh
ovs-ofctl add-flow \<bridge\> dl_src=\<mac\>,actions=drop
```

> Unblock all outgoing traffic by matching field dl\_src with an Ethernet source address. 
```sh
ovs-ofctl del-flows --strict \<bridge\> dl_src=\<mac\>
```
> Block all incoming traffic by matching field dl\_dst with an Ethernet source address. This value uses 6 pairs of hexadecimal digits to specify, eg: 00:0B:C4:A8:22:B0.
```sh
ovs-ofctl add-flow \<bridge\> dl_dst=\<mac\>,actions=drop
```

> Unblock all outgoing traffic by matching field dl\_dst with an Ethernet source address. 
```sh
ovs-ofctl del-flows --strict \<bridge\> dl_dst=\<mac\>
```

