package bcast

import (
	"elevator_project/data/types"
	"elevator_project/network/connection"
	"elevator_project/config"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"time"
)


const bufSize = 65507
var onlineStatus = true

type typeTaggedJSON struct {
	TypeId string
	JSON   []byte
}

func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func UpdateOnlineStatus(status bool) {
	onlineStatus = status
}

func Receiver(dataChan chan<- types.NodeData) {
	conn := connection.DialBroadcastUDP(config.NodePort)
	localIP, err := getLocalIP()

	if err != nil {
		fmt.Printf("bcast.Receiver(%d, ...): Failed to get local IP: \"%+v\"\n", config.NodePort, err)
		return
	}

	for {
		var buf [bufSize]byte
		conn.SetDeadline(time.Now().Add(100*time.Millisecond))
		n, addr, e := conn.ReadFrom(buf[0:])
		if e != nil {
			fmt.Printf("bcast.Receiver(%d, ...):ReadFrom() failed: \"%+v\"\n", config.NodePort, e)
			continue
		}

		if addr.(*net.UDPAddr).IP.String() == localIP {
			continue
		}

		var ttj typeTaggedJSON
		json.Unmarshal(buf[0:n], &ttj)
		if ttj.TypeId != reflect.TypeOf(types.NodeData{}).String() {
			continue
		}

		var data types.NodeData
		json.Unmarshal(ttj.JSON, &data)
		dataChan <- data
	}
}

func Transmitter(updatePerMsg <-chan types.NodeData) {
	periodicMsg := types.EmptyPeriodicMsg()
	conn := connection.DialBroadcastUDP(config.NodePort)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", config.NodePort))

	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if onlineStatus {
				jsonstr, _ := json.Marshal(periodicMsg)
				ttj, _ := json.Marshal(typeTaggedJSON{
					TypeId: reflect.TypeOf(periodicMsg).String(),
					JSON:   jsonstr,
				})

				if len(ttj) > bufSize {panic(fmt.Sprintf("Tried to send a message longer than the buffer size"))}
				conn.WriteTo(ttj, addr)
			}

		case newData := <- updatePerMsg:
			periodicMsg = newData
		}
	}
}
