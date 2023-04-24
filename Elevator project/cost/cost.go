package cost

import (
	"encoding/json"
	"elevator_project/config"
	"elevator_project/data/types"
	"elevator_project/elevio"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
)

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests types.OrderTable        `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func isUnassignedOrdersEmpty(orders types.OrderTable) bool {
	for i := 0; i < config.NumFloors; i++ {
		for j := 0; j < 2; j++ {
			if orders[i][j] == true {
				return false
			}
		}
	}
	return true
}

func RunCostFunc(myNodeData types.NodeData, peers types.PeersOverview, peersData map[int]types.NodeData) types.OrderTable {
	if isUnassignedOrdersEmpty(myNodeData.AllHallOrders) {
		fmt.Println("Returned empty hall order list")
		return myNodeData.AllHallOrders
	}

	myIdStr := strconv.Itoa(config.ID)
	activeNodes := len(peers.Peers)
	elevStates := make(map[string]HRAElevState, activeNodes)
	peersData[config.ID] = myNodeData

	for node := 0; node < activeNodes; node++ {
		id, _ := strconv.Atoi(peers.Peers[node])
		data := peersData[id]
		elevStates[strconv.Itoa(id)] = HRAElevState{
			Behavior:    types.Eb_toString(data.Elev.Behaviour),
			Floor:       data.Elev.Floor,
			Direction:   elevio.ElevioDirnToString(data.Elev.Dirn),
			CabRequests: data.CabOrders,
		}
	}

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	input := HRAInput{
		HallRequests: myNodeData.AllHallOrders,
		States:       elevStates,
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
	}

	ret, err := exec.Command(hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
	}

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
	}

	myOrders := (*output)[myIdStr]

	return myOrders
}
