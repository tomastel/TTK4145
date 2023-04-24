package types

import (
	"elevator_project/config"
	"elevator_project/elevio"
	"io/ioutil"
	"strconv"
)

type ElevatorBehaviour int

type AckedList [3]bool
type CabOrderTable [config.NumElevators][]bool
type OrderAckTable [][2]AckedList
type OrderTable [][2]bool

type Elevator struct {
	Floor 						int
	Dirn 						elevio.MotorDirection
	Requests 					[config.NumFloors][3]bool
	Behaviour					ElevatorBehaviour
	DoorOpenDuration_s			float64
}

type NodeData struct {
	ElevatorID 			int
	Elev       			Elevator
	MyHallOrders  		OrderTable
	AllHallOrders 		OrderTable
	AckTable 			OrderAckTable
	CabOrders  			[]bool
}

type PeersOverview struct {
	Peers []string
	New   string
	Lost  []string
}

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

var NewNodeData = NodeData{
	ElevatorID: config.ID,
	Elev:       NewElevator(),
	MyHallOrders: EmptyOrders(),
	AllHallOrders: EmptyOrders(),
	CabOrders:  EmptyCabOrders(),
	AckTable:   EmptyAckTable(),
}

func CombineOrderAckTables(table1, table2 OrderAckTable) OrderAckTable {
	combinedTable := make(OrderAckTable, len(table1))
	for i := range table1 {
		combinedTable[i] = [2]AckedList{{false,false,false}, {false, false, false}}
		for j := 0; j < 2; j++ {
			for k := 0; k < 3; k++ {
				if table1[i][j][k] || table2[i][j][k] {
					combinedTable[i][j][k] = true
				}
			}
		}
	}
	return combinedTable
}

func CountTrue(list AckedList) int {
	count := 0
	for _, value := range list {
		if value {
			count++
		}
	}
	return count
}

func Eb_toString(eb ElevatorBehaviour) string {
	if eb == EB_Idle {
		return "idle"
	} else if eb == EB_DoorOpen {
		return "doorOpen"
	} else if eb == EB_Moving {
		return "moving"
	} else {
		return "undefined"
	}
}

func Elevator_uninitialized() Elevator {
	elev := Elevator{
		Floor:					0,
		Dirn: 					elevio.MD_Stop,
		Behaviour: 				EB_Idle,
		DoorOpenDuration_s: 	3.0,
		Requests:       	    [config.NumFloors][3]bool{},
	}	

	return elev
}

func EmptyAckTable() OrderAckTable {
	ackTable := make(OrderAckTable, 4)
	for i := range ackTable {
		ackTable[i] = [2]AckedList{{false, false, false}, {false, false, false}}
	}
	return ackTable
}

func EmptyCabOrders() []bool {
	cabOrders := make([]bool, config.NumFloors)
	for i := 0; i < config.NumFloors; i++ {
		cabOrders[i] = false
	}
	return cabOrders
}

func EmptyOrders() OrderTable {
	orderTable := make(OrderTable, config.NumFloors)
	for i := range orderTable {
		orderTable[i] = [2]bool{false, false}
	}
	return orderTable
}

func EmptyPeriodicMsg() NodeData {
	return NodeData{
		ElevatorID: config.ID,
		Elev:    NewElevator(),
		MyHallOrders: EmptyOrders(),
		AllHallOrders: EmptyOrders(),
		AckTable: EmptyAckTable(),
		CabOrders: make([]bool, 4),
	}
}

func NewElevator() Elevator {
    return Elevator{
        Floor:                0,
        Dirn:                 elevio.MD_Stop,
        Behaviour:            EB_Idle,
        DoorOpenDuration_s:   3.0,
    }
}

func ReadCabOrdersFromFile() ([]bool, error) {
    byteSlice, err := ioutil.ReadFile(config.CabOrderBackupFile)
    if err != nil {
        return nil, err
    }

    var cabOrders []bool
    for _, b := range byteSlice {
        order, err := strconv.ParseBool(string(b))
        if err != nil {
            return nil, err
        }
        cabOrders = append(cabOrders, order)
    }

    return cabOrders, nil
}

func UpdateAllActiveOrders(a, b OrderTable) (OrderTable, bool) {
	c := make(OrderTable, config.NumFloors)
	newActiveOrder := false
	for i := range a {
		for j := 0; j < 2; j++ {
			if b[i][j] {
				c[i][j] = true
				if !a[i][j] {
					newActiveOrder = true
				}
			} else {
				c[i][j] = a[i][j]
			}
		}
	}
	return c, newActiveOrder
}

func WriteCabOrdersToFile(cabOrders []bool) error {
    var byteSlice []byte
    for _, order := range cabOrders {
        if order {
            byteSlice = append(byteSlice, '1')
        } else {
            byteSlice = append(byteSlice, '0')
        }
    }

    return ioutil.WriteFile(config.CabOrderBackupFile, byteSlice, 0644)
}
