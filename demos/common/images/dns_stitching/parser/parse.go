package parser

import (
	"crypto/sha256"
	"dns_stitching/color"
	"strconv"

	//"dns_stitching/color"
	"dns_stitching/iohandlers"
	"encoding/csv"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
	"time"
	//"strconv"
)

var (
	DoParseTcp          = true
	DoParseQuestions    = false
	DoParseQuestionsEcs = true
	Source              = ""
	Sensor              = ""
	OutputFormat        = ""
)

type flow struct {
	sourceIp   string
	destIp     string
	sourcePort uint16
	destPort   uint16
	noPackets  uint
	rName      string
}

var flows_arr = []flow{}

type resourceRecord struct {
	rName string
	Ttl   int64
}

//var num = 0
var cache = make(map[string]map[string]*resourceRecord)

//var cache = make(map[string]map[string]string)

func ParseFile(fname string) {
	var (
		handle *pcap.Handle
		err    error
	)

	if "-" == fname {
		handle, err = pcap.OpenOfflineFile(os.Stdin)
	} else {
		handle, err = pcap.OpenOffline(fname)
	}

	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Setup BPF filter on handle
	//bpfFilter := "udp port 53 or (vlan and udp port 53)"
	//if DoParseTcp {
	//	bpfFilter = "port 53 or (vlan and port 53)"
	//}
	//err = handle.SetBPFFilter(bpfFilter)
	//if err != nil {
	//	log.Warnf("Could not set BPF filter: %v\n", err)
	//}

	ParseDns(handle)
}

func ParseDevice(device string, snapshotLen int32, promiscuous bool, timeout time.Duration) {
	handle, err := pcap.OpenLive(device, snapshotLen, promiscuous, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Setup BPF filter on handle
	//bpfFilter := "udp port 53 or (vlan and udp port 53)"
	//if DoParseTcp {
	//	bpfFilter = "port 53 or (vlan and port 53)"
	//}
	//err = handle.SetBPFFilter(bpfFilter)
	//if err != nil {
	//	log.Warnf("Could not set BPF filter: %v\n", err)
	//}

	ParseDns(handle)
}

func ParseDns(handle *pcap.Handle) {
	var (
		schema iohandlers.DnsSchema
		stats  Statistics
		ip4    *layers.IPv4
		ip6    *layers.IPv6
		tcp    *layers.TCP
		udp    *layers.UDP
		msg    *dns.Msg
	)

	var num = 1

	// Set the source and sensor for packet source
	schema.Sensor = Sensor
	schema.Source = Source

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetSource.NoCopy = true
	packetSource.Lazy = true

	// Initialize IO handler for output format
	iohandlers.Initialize(OutputFormat)

	file, err := os.Create("server/summary.csv")
	defer file.Close()
	if err != nil {
		log.Fatalln("failed to open file", err)
	}
	w := csv.NewWriter(file)
	//defer w.Flush()
	//defer func() {
	//	w.Flush()
	//	file.Close()
	//}()

PACKETLOOP:
	for {
		packet, err := packetSource.NextPacket()
		if err == io.EOF {
			break
		}
		stats.PacketTotal += 1

		if err != nil {
			//log.Errorf("Error decoding some part of the packet: %v\n", err)
			stats.PacketErrors += 1
			continue
		}

		// Parse network layer information
		networkLayer := packet.NetworkLayer()
		if networkLayer == nil {
			//log.Error("Unknown/missing network layer for packet")
			stats.PacketErrors += 1
			continue
		}
		switch networkLayer.LayerType() {
		case layers.LayerTypeIPv4:
			ip4 = networkLayer.(*layers.IPv4)
			schema.SourceAddress = ip4.SrcIP.String()
			schema.DestinationAddress = ip4.DstIP.String()
			schema.Ipv4 = true
			stats.PacketIPv4 += 1
		case layers.LayerTypeIPv6:
			ip6 = networkLayer.(*layers.IPv6)
			schema.SourceAddress = ip6.SrcIP.String()
			schema.DestinationAddress = ip6.DstIP.String()
			schema.Ipv4 = false
			stats.PacketIPv6 += 1
		}

		//fmt.Println("Before transport layer")
		//fmt.Println(stats.PacketTotal)
		//fmt.Printf("From %s to %s\n", schema.SourceAddress, schema.DestinationAddress)
		//fmt.Println()

		//Ned to add this for UDP
		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if tcpLayer != nil {
			//fmt.Println("TCP layer detected.")
			num = num + 1
			tcp, _ := tcpLayer.(*layers.TCP)
			schema.DestinationPort = uint16(tcp.DstPort)
		}

		// Parse DNS and transport layer information
		msg = nil
		transportLayer := packet.TransportLayer()
		if transportLayer == nil {
			//log.Error("Unknown/missing transport layer for packet")
			stats.PacketErrors += 1
			continue
		}

		switch transportLayer.LayerType() {
		case layers.LayerTypeTCP:
			tcp = transportLayer.(*layers.TCP)
			stats.PacketTcp += 1

			if !DoParseTcp {
				//continue PACKETLOOP
			}

			msg = new(dns.Msg)
			if err := msg.Unpack(tcp.Payload); err != nil {
				//log.Errorf("Could not decode DNS: %v\n", err)
				stats.PacketErrors += 1
				//continue PACKETLOOP
			}
			stats.PacketDns += 1

			schema.SourcePort = uint16(tcp.SrcPort)
			schema.DestinationPort = uint16(tcp.DstPort)
			schema.Udp = false
			schema.Sha256 = fmt.Sprintf("%x", sha256.Sum256(tcp.Payload))
		case layers.LayerTypeUDP:
			udp = transportLayer.(*layers.UDP)
			stats.PacketUdp += 1

			msg = new(dns.Msg)
			if err := msg.Unpack(udp.Payload); err != nil {
				//log.Errorf("Could not decode DNS: %v\n", err)
				stats.PacketErrors += 1
				//continue PACKETLOOP
			}
			stats.PacketDns += 1

			schema.SourcePort = uint16(udp.SrcPort)
			schema.DestinationPort = uint16(udp.DstPort)
			schema.Udp = true

			// Hash and salt packet for grouping related records
			tsSalt, err := packet.Metadata().Timestamp.MarshalBinary()
			if err != nil {
				log.Errorf("Could not marshal timestamp: #{err}\n")
			}
			schema.Sha256 = fmt.Sprintf("%x", sha256.Sum256(append(tsSalt, packet.Data()...)))
		}

		//fmt.Println("After transport layer")
		//fmt.Printf("From %s to %s\n", schema.SourceAddress, schema.DestinationAddress)
		//fmt.Println()
		fmt.Println(stats.PacketTotal)
		if schema.SourcePort == 53 {
			num = num + 1
			// This means we did not attempt to parse a DNS payload and
			// indicates an unexpected transport layer protocol
			if msg == nil {
				//log.Debug("Unexpected transport layer protocol")
				continue PACKETLOOP
			}

			// Ignore questions unless flag set
			if !msg.Response && !DoParseQuestions && !DoParseQuestionsEcs {
				continue PACKETLOOP
			}

			// Fill out information from DNS headers
			schema.Timestamp = packet.Metadata().Timestamp.Unix()
			schema.Id = msg.Id
			schema.Rcode = msg.Rcode
			schema.Truncated = msg.Truncated
			schema.Response = msg.Response
			schema.RecursionDesired = msg.RecursionDesired

			// Parse ECS information
			schema.EcsClient = nil
			schema.EcsSource = nil
			schema.EcsScope = nil
			if opt := msg.IsEdns0(); opt != nil {
				for _, s := range opt.Option {
					switch o := s.(type) {
					case *dns.EDNS0_SUBNET:
						ecsClient := o.Address.String()
						ecsSource := o.SourceNetmask
						ecsScope := o.SourceScope
						schema.EcsClient = &ecsClient
						schema.EcsSource = &ecsSource
						schema.EcsScope = &ecsScope
					}
				}
			}

			// Reset RR information
			schema.Ttl = nil
			schema.Rname = nil
			schema.Rdata = nil
			schema.Rtype = nil

			// Let's get QUESTION
			// TODO: Throw error if there's more than one question
			for _, qr := range msg.Question {
				schema.Qname = qr.Name
				schema.Qtype = qr.Qtype
			}

			// Get a count of RRs in DNS response
			rrCount := 0
			for _, rr := range append(append(msg.Answer, msg.Ns...), msg.Extra...) {
				if rr.Header().Rrtype != 41 {
					rrCount++
				}
			}

			// Let's output records without RRs records if:
			//   1. Questions flag is set and record is question
			//   2. QuestionsEcs flag is set and question record contains ECS information
			//   4. Any response without any RRs (e.g., NXDOMAIN without SOA, REFUSED, etc.)
			if (DoParseQuestions && !schema.Response) ||
				(DoParseQuestionsEcs && schema.EcsClient != nil && !schema.Response) ||
				(schema.Response && rrCount < 1) {
				schema.Marshal(nil, -1, OutputFormat)
			}

			// Let's get ANSWERS
			for _, rr := range msg.Answer {
				//schema.Marshal(&rr, iohandlers.DnsAnswer, OutputFormat)
				rtype := (rr).Header().Rrtype
				//fmt.Println("RTYPE",rtype)
				if rtype == 1 {
					rdata := strings.TrimPrefix((rr).String(), (rr).Header().String())
					fmt.Println("RDATA", rdata)
					ttl := (rr).Header().Ttl
					ttl64 := int64(ttl)
					timestamp := packet.Metadata().Timestamp.Unix()
					timeLive := timestamp + ttl64
					rname := (rr).Header().Name
					//fmt.Println("RNAME", rName)
					flag1 := 0
					if cache[schema.DestinationAddress] == nil {
						cache[schema.DestinationAddress] = make(map[string]*resourceRecord)
						cache[schema.DestinationAddress][rdata] = &resourceRecord{rName: rname, Ttl: timeLive}
						//fmt.Println("New source added: ", rdata, "    ", timeLive)
						//cache[schema.DestinationAddress] = make(map[string]string)
					} else {
						for key, val := range cache {
							if key == schema.DestinationAddress {
								for key1, _ := range val {
									if key1 == rdata {
										cache[key][key1] = &resourceRecord{rName: rname, Ttl: timeLive}
										flag1 = flag1 + 1
										//println("replaced:", rdata, "    ", timeLive)
										break
									}
								}
							}
							if flag1 == 0 {
								cache[schema.DestinationAddress][rdata] = &resourceRecord{rName: rname, Ttl: timeLive}
								//fmt.Println("Same source added: ", rdata, "    ", timeLive)
							}
							break
						}
					}
					//cache[schema.DestinationAddress][rdata] = rname
					//cache[schema.DestinationAddress][rdata] = &resourceRecord{rName: rname, Ttl: timeLive}
				}

			}

			//fmt.Println("\n")

			// Let's get AUTHORITATIVE information
			//for _, rr := range msg.Ns {
			//	schema.Marshal(&rr, iohandlers.DnsAuthority, OutputFormat)
			//}

			// Let's get ADDITIONAL information
			//for _, rr := range msg.Extra {
			//	schema.Marshal(&rr, iohandlers.DnsAdditional, OutputFormat)
			//}
		} else {
			if schema.DestinationPort == 53 {
				continue
			}

			//--------------------
			//Adding packets to flows array

			//Checking if there is already a flow for the packet
			flag := 0
			//fmt.Println(stats.PacketTotal)
			for i, elem := range flows_arr {
				if elem.sourceIp == schema.SourceAddress && elem.destIp == schema.DestinationAddress || elem.sourceIp == schema.DestinationAddress && elem.destIp == schema.SourceAddress {
					if elem.sourcePort == schema.SourcePort && elem.destPort == schema.DestinationPort || elem.sourcePort == schema.DestinationPort && elem.destPort == schema.SourcePort {
						flows_arr[i].noPackets = flows_arr[i].noPackets + 1
						flag = flag + 1
						//fmt.Println(flows_arr[i])
						//fmt.Println("Existing flow: ", flows_arr[i])
						continue PACKETLOOP
					}
				}
			}

			//If the packet is the first packet in the flow.Create a new flow and append it to the flow array
			if flag == 0 {
				tempFlow := flow{schema.SourceAddress, schema.DestinationAddress, schema.SourcePort, schema.DestinationPort, 1, "Not resolved"}
				//flows_arr = append(flows_arr, tempFlow)
				//fmt.Println(flows_arr)
				//fmt.Println("New flow: ", tempFlow)

				//Cache checking with TTL

				if cache[schema.SourceAddress] != nil {
					if cache[schema.SourceAddress][schema.DestinationAddress] != nil {
						if tempFlow.sourceIp == schema.SourceAddress && tempFlow.destIp == schema.DestinationAddress {
							temp := *(cache[schema.SourceAddress][schema.DestinationAddress])
							tstamp := packet.Metadata().Timestamp.Unix()
							//fmt.Println("Packet time: ", tstamp, "  Ttl: ", temp.Ttl)
							if temp.Ttl >= tstamp {
								//fmt.Println("1 Caught packet :",stats.PacketTotal)
								//fmt.Println("Stitched")
								tempFlow.rName = temp.rName
								//fmt.Println(flows_arr[i])
							} else {
								//fmt.Println("2 Caught packet :",stats.PacketTotal)
								//fmt.Println("Expired")
								tempFlow.rName = "Record expired"
								//fmt.Println(flows_arr[i])
							}
							//fmt.Println(" Flow "+flows_arr[i].sourceIp+" --- "+flows_arr[i].destIp+" Source port: ", flows_arr[i].sourcePort, " Destination port: ", flows_arr[i].destPort, "No. packets: ", flows_arr[i].noPackets, color.Blue+" Rname: "+flows_arr[i].rName+color.Reset)
						} else if tempFlow.sourceIp == schema.DestinationAddress && tempFlow.destIp == schema.SourceAddress {
							temp := *(cache[schema.SourceAddress][schema.DestinationAddress])
							tstamp := packet.Metadata().Timestamp.Unix()
							//fmt.Println("Packet time: ", tstamp, "  Ttl: ", temp.Ttl)
							if temp.Ttl >= tstamp {
								//fmt.Println("1I Caught packet :",stats.PacketTotal)
								//fmt.Println("Stitched")
								tempFlow.rName = temp.rName
								//fmt.Println(flows_arr[i])
							} else {
								//fmt.Println("2I Caught packet :",stats.PacketTotal)
								//fmt.Println("Expired")
								tempFlow.rName = "Record expired"
								//fmt.Println(flows_arr[i])
							}
							//fmt.Println(" Flow "+flows_arr[i].sourceIp+" --- "+flows_arr[i].destIp+" Source port: ", flows_arr[i].sourcePort, " Destination port: ", flows_arr[i].destPort, "No. packets: ", flows_arr[i].noPackets, color.Blue+" Rname: "+flows_arr[i].rName+color.Reset)
						}

					} else {
						//fmt.Println("3 Caught packet :",stats.PacketTotal)
						//fmt.Println(" From " + schema.SourceAddress + " to " + schema.DestinationAddress + color.Red + " --Cache miss " + color.Reset)
					}
				} else if cache[schema.DestinationAddress] != nil {
					if cache[schema.DestinationAddress][schema.SourceAddress] != nil {
						if tempFlow.sourceIp == schema.DestinationAddress && tempFlow.destIp == schema.SourceAddress {
							temp := *(cache[schema.DestinationAddress][schema.SourceAddress])
							tstamp := packet.Metadata().Timestamp.Unix()
							//fmt.Println("Packet time: ", tstamp, "  Ttl: ", temp.Ttl)
							if temp.Ttl >= tstamp {
								//fmt.Println("4 Caught packet :",stats.PacketTotal)
								//fmt.Println("Stitched 2")
								tempFlow.rName = temp.rName
								//fmt.Println(flows_arr[i])
							} else {
								//fmt.Println("5 Caught packet :",stats.PacketTotal)
								//fmt.Println("Expired 2")
								tempFlow.rName = "Record expired"
								//fmt.Println(flows_arr[i])
							}
							//fmt.Println(" Flow "+flows_arr[i].sourceIp+" --- "+flows_arr[i].destIp+" Source port: ", flows_arr[i].sourcePort, " Destination port: ", flows_arr[i].destPort, "No. packets: ", flows_arr[i].noPackets, color.Blue+" Rname: "+flows_arr[i].rName+color.Reset)
						} else if tempFlow.sourceIp == schema.SourceAddress && tempFlow.destIp == schema.DestinationAddress {
							temp := *(cache[schema.DestinationAddress][schema.SourceAddress])
							tstamp := packet.Metadata().Timestamp.Unix()
							//fmt.Println("Packet time: ", tstamp, "  Ttl: ", temp.Ttl)
							if temp.Ttl >= tstamp {
								//fmt.Println("4I Caught packet :",stats.PacketTotal)
								//fmt.Println("Stitched 2")
								tempFlow.rName = temp.rName
								//fmt.Println(flows_arr[i])
							} else {
								//fmt.Println("5I Caught packet :",stats.PacketTotal)
								//fmt.Println("Expired 2")
								tempFlow.rName = "Record expired"
								//fmt.Println(flows_arr[i])
							}
							//fmt.Println(" Flow "+flows_arr[i].sourceIp+" --- "+flows_arr[i].destIp+" Source port: ", flows_arr[i].sourcePort, " Destination port: ", flows_arr[i].destPort, "No. packets: ", flows_arr[i].noPackets, color.Blue+" Rname: "+flows_arr[i].rName+color.Reset)
						}
					} else {
						//fmt.Println("6 Caught packet :",stats.PacketTotal)
						//fmt.Println(" From " + schema.SourceAddress + " to " + schema.DestinationAddress + color.Red + " --Cache miss " + color.Reset)
					}
				} else {
					//fmt.Println("7 Caught packet :",stats.PacketTotal)
					//fmt.Println(" From " + schema.SourceAddress + " to " + schema.DestinationAddress + color.Red + " --Cache miss " + color.Reset)
					//fmt.Println(" From " + schema.SourceAddress + " to " + schema.DestinationAddress)
				}

				flows_arr = append(flows_arr, tempFlow)
				row := []string{tempFlow.sourceIp, tempFlow.destIp, strconv.Itoa(int(tempFlow.sourcePort)), strconv.Itoa(int(tempFlow.destPort)), tempFlow.rName}
				if err := w.Write(row); err != nil {
					log.Fatalln("error writing record to file", err)
				}
				//w.Flush()
			}
		}
		w.Flush()
	}

	fmt.Println("\n\n")
	for _, elem := range flows_arr {
		fmt.Println(elem.sourceIp+" --- "+elem.destIp+" Source port: ", elem.sourcePort, " Destination port: ", elem.destPort, " No. packets: ", elem.noPackets, color.Blue+" Rname: "+elem.rName+color.Reset)
	}

	//Writing summary to file

	//var data [][]string
	//for _, elem := range flows_arr{
	//	if elem.rName != ""{
	//		row := []string{elem.sourceIp, elem.destIp, strconv.Itoa(int(elem.sourcePort)), strconv.Itoa(int(elem.destPort)), strconv.Itoa(int(elem.noPackets)), elem.rName}
	//		data = append(data,row)
	//	} else{
	//		row := []string{elem.sourceIp, elem.destIp, strconv.Itoa(int(elem.sourcePort)), strconv.Itoa(int(elem.destPort)), strconv.Itoa(int(elem.noPackets)), "No resolution"}
	//		data = append(data,row)
	//	}
	//	w.WriteAll(data)
	//}

}
