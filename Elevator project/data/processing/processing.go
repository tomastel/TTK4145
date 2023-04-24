package processing

import (
	"elevator_project/config"
	"elevator_project/cost"
	"elevator_project/data/types"
	"elevator_project/elevio"
	"elevator_project/fsm"
	"elevator_project/network/bcast"
	"fmt"
	"sort"
	"strconv"
	"time"
)

const timeout = time.Second

var myNodeData types.NodeData
var peers types.PeersOverview

var lastSeen = make(map[string]time.Time)
var peersData = make(map[int]types.NodeData)

var updateCabOrdersChan = make(chan []bool, 1)
var updatePerMsgChan = make(chan types.NodeData)

func setAllLights() {
	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := 0; btn < 2; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, myNodeData.AllHallOrders[floor][btn])
		}

		elevio.SetButtonLamp(elevio.BT_Cab, floor, myNodeData.CabOrders[floor])
	}
}

func updatePeersOverview(msgInID int) bool {
	msgIDStr := strconv.Itoa(msgInID)
	updated := false
	peers.New = ""

	if msgIDStr != "" {
		if _, idExists := lastSeen[msgIDStr]; !idExists {
			fmt.Println("Node Added: ", msgInID)
			peers.New = msgIDStr
			updated = true
		}

		lastSeen[msgIDStr] = time.Now()
	}

	peers.Lost = make([]string, 0)
	for k, v := range lastSeen {
		if k == strconv.Itoa(config.ID) {
			continue
		}

		if time.Now().Sub(v) > timeout {
			fmt.Println("Node Deleted: ", k)
			updated = true
			peers.Lost = append(peers.Lost, k)
			delete(lastSeen, k)
		}
	}

	if updated {
		peers.Peers = make([]string, 0, len(lastSeen))

		for k, _ := range lastSeen {
			peers.Peers = append(peers.Peers, k)
		}
		sort.Strings(peers.Peers)
		sort.Strings(peers.Lost)
	}

	return updated
}

func updatePeersData(msg types.NodeData) bool {
	newActiveOrder := false
	peersData[msg.ElevatorID] = msg
	myNodeData.AllHallOrders, newActiveOrder = types.UpdateAllActiveOrders(myNodeData.AllHallOrders, msg.AllHallOrders)
	return newActiveOrder
}

func shouldRemoveActiveOrder(ordersToRemove types.OrderTable, ackTable types.OrderAckTable, activeOrders types.OrderTable) (types.OrderAckTable, types.OrderTable) {
	for i := 0; i < len(ordersToRemove); i++ {
		for j := 0; j < len(ordersToRemove[i]); j++ {
			if ordersToRemove[i][j] {
				activeOrders[i][j] = false
				ordersToRemove[i][j] = false
			}
		}
	}
	return ackTable, activeOrders
}

func shouldUpdateAckCounter(numPeers int, peerAckTable, myAckTable types.OrderAckTable) (types.OrderAckTable, types.OrderTable) {
	ordersToRemove := make(types.OrderTable, len(peerAckTable))
	var inCounter int
	var myCounter int

	for i := range ordersToRemove {
		ordersToRemove[i] = [2]bool{false, false}
	}

	for i, table := range peerAckTable {
		for j, list := range table {
			inCounter = types.CountTrue(peerAckTable[i][j])
			myCounter = types.CountTrue(myAckTable[i][j])

			if myCounter == 0 && inCounter == numPeers {
				continue
			}

			if myCounter == numPeers && inCounter == 0 {
				ordersToRemove[i][j] = true
				myAckTable[i][j] = [3]bool{false, false, false}
				continue
			}

			for _, value := range list {
				if value {
					myAckTable[i][j] = peerAckTable[i][j]
					myAckTable[i][j][config.ID] = true
				}
			}

			if types.CountTrue(peerAckTable[i][j]) == numPeers {
				ordersToRemove[i][j] = true
				myAckTable[i][j] = [3]bool{false, false, false}
			}
		}
	}
	return myAckTable, ordersToRemove
}

func updateAckTable(msg types.OrderAckTable) {
	var ordersToRemove = make(types.OrderTable, 4)

	myNodeData.AckTable, ordersToRemove = shouldUpdateAckCounter(len(peers.Peers), msg, myNodeData.AckTable)
	myNodeData.AckTable, myNodeData.AllHallOrders = shouldRemoveActiveOrder(ordersToRemove, myNodeData.AckTable, myNodeData.AllHallOrders)
}

func processingDataFlow(cabOrdersChan chan []bool) {
	var updateElevStateChan = make(chan types.Elevator)
	var orderFinishedChan = make(chan elevio.ButtonEvent)
	var hallOrdersChan = make(chan types.OrderTable)

	bcastRx := make(chan types.NodeData)
	drv_buttons := make(chan elevio.ButtonEvent)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	go elevio.PollButtons(drv_buttons)
	go bcast.Transmitter(updatePerMsgChan)
	go bcast.Receiver(bcastRx)
	go fsm.Fsm_elevatorModule(hallOrdersChan, cabOrdersChan)
	go fsm.SendDataToDP(updateElevStateChan, orderFinishedChan)

	for {
		select {
		case <-ticker.C:
			updated := updatePeersOverview(myNodeData.ElevatorID)
			if updated {
				myNodeData.MyHallOrders = cost.RunCostFunc(myNodeData, peers, peersData)
				hallOrdersChan <- myNodeData.MyHallOrders
			}

		case msg := <-bcastRx:
			updated := updatePeersOverview(msg.ElevatorID)
			newActiveOrder := updatePeersData(msg)
			if updated || newActiveOrder {
				myNodeData.MyHallOrders = cost.RunCostFunc(myNodeData, peers, peersData)
				hallOrdersChan <- myNodeData.MyHallOrders
			}
			updateAckTable(msg.AckTable)

		case btnEvent := <-drv_buttons:
			if btnEvent.Button == elevio.BT_Cab {
				myNodeData.CabOrders[btnEvent.Floor] = true
				types.WriteCabOrdersToFile(myNodeData.CabOrders)
				cabOrdersChan <- myNodeData.CabOrders
			} else {
				myNodeData.AllHallOrders[btnEvent.Floor][btnEvent.Button] = true
				myNodeData.MyHallOrders = cost.RunCostFunc(myNodeData, peers, peersData)
				hallOrdersChan <- myNodeData.MyHallOrders
			}

		case elevData := <-updateElevStateChan:
			myNodeData.Elev = elevData

		case orderFinished := <-orderFinishedChan:
			if orderFinished.Button == elevio.BT_Cab {
				myNodeData.CabOrders[orderFinished.Floor] = false
				types.WriteCabOrdersToFile(myNodeData.CabOrders)
			} else {
				myNodeData.MyHallOrders[orderFinished.Floor][orderFinished.Button] = false
				myNodeData.AckTable[orderFinished.Floor][orderFinished.Button][config.ID] = true
				updateAckTable(myNodeData.AckTable)
			}
		}

		setAllLights()
		updatePerMsgChan <- myNodeData
	}
}

func InitDataModule(restore int) {
	myNodeData = types.NewNodeData
	myNodeData.ElevatorID = config.ID
	myIDstr := strconv.Itoa(config.ID)
	peersData[config.ID] = myNodeData
	peers.New = myIDstr
	peers.Peers = append(peers.Peers, myIDstr)
	lastSeen[myIDstr] = time.Now()

	if restore == 1 {
		if cabOrdersFromFile, err := types.ReadCabOrdersFromFile(); err == nil {
			myNodeData.CabOrders = cabOrdersFromFile
			updateCabOrdersChan <- myNodeData.CabOrders
		} else {
			fmt.Println("Failed to read from file, no cab orders restored")
		}
	}

	fmt.Println("Starting program")
	go processingDataFlow(updateCabOrdersChan)
	updatePerMsgChan <- myNodeData
}
