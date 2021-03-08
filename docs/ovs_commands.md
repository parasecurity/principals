# Base commands 

OVS is feature rich with different configuration commands, but the majority of your configuration and troubleshooting can be accomplished with the following 3 commands:

  - ovs-vsctl : Used for configuring the ovs-vswitchd configuration database (known as ovs-db)
  - ovs-ofctl : A command line tool for monitoring and administering OpenFlow switches
  - ovs-dpctl : Used to administer Open vSwitch datapaths

## ovs-vsctl

This tool is used for configuration and viewing OVS switch operations. Port configuration, bridge additions/deletions, bonding, and VLAN tagging are just some of the options that are available with this command.

  - ovs-vsctl -V : Prints the current version of openvswitch.
  - ovs-vsctl show: Prints the current bridge configuration
  - ovs-vsctl list-br: Prints the names of all configured bridges 
  - ovs-vsctl list-ports \<_bridge_\>: Prints a list of ports connected on a specific bridge
  - ovs-vsctl list interface: Prints a list of interfaces
  - ovs-vsctl add-br \<_bridge_\>: Adds a new bridge
  - ovs-vsctl add-port \<_bridge_\> \<_interface_\>: Adds a new port on selected bridge
  - ovs-vsctl del-port \<_bridge_\> \<_interface_\>: Deletes a specific port on a bridge

## ovs-ofctl

This tool is used for administering and monitoring OpenFlow switches. Even if OVS isn’t configured for centralized administration, ovs-ofctl can be used to show the current state of OVS including features, configuration, and table entries.

  - ovs-ofctl show \<_bridge_\>: Shows OpenFlow features and port descriptions
  - ovs-ofctl dump-flows \<_bridge_\>: Prints flow entries of specified bridge. 
  - ovs-ofctl dump-ports-desc \<_bridge_\>: Prints port statistics. This will show detailed information about interfaces in this bridge, include the state, peer, and speed information
  - ovs-ofctl add-flow \<_bridge_\> \<_flow_\>: Add a static flow to the specified bridge
  - ovs-ofctl del-flows \<_bridge_\> \<_flow_\>: Delete the flow entries from flow table of stated bridge. If the flow is omitted, all flows in specified bridge will be deleted

## ovs-dpctl

This tool is very similar to ovs-ofctl in that they both show flow table entries. The flows that ovs-dpctl prints are always an exact match and reflect packets that have actually passed through the system within the last few seconds. ovs-dpctl queries a kernel datapath and not an OpenFlow switch. This is why it’s useful for debugging flow data.
  - ovs-dpctl dump-flows: Prints flow table data

