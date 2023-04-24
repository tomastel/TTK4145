package fsm

import (
	"elevator_project/config"
	"elevator_project/data/types"
	"elevator_project/elevio"
	"elevator_project/fault"
	"elevator_project/request"
	"elevator_project/timer"
	"fmt"
	"strconv"
)

var doorClosedChan = make(chan bool)
var doorObstructionChan = make(chan bool)
var elevatorBehaviourChan = make(chan types.ElevatorBehaviour)
var requestFinishedChan = make(chan elevio.ButtonEvent)
var sendElevDataChan = make(chan bool)

var my_elevator types.Elevator

func cabOrderToRequests(orders []bool) {
	for i, val := range orders {
		my_elevator.Requests[i][2] = val
	}
}

func hallOrdersToRequests(orders types.OrderTable) {
	for i := 0; i < config.NumFloors; i++ {
		for j := 0; j < 2; j++ {
			my_elevator.Requests[i][j] = orders[i][j]
		}
	}
}

func obstruction() {
	if my_elevator.Behaviour == types.EB_DoorOpen {
		timer.TimerStart(my_elevator.DoorOpenDuration_s)
		doorObstructionChan <- true	
	}
}

func onDoorTimeout() {
	switch my_elevator.Behaviour {
	case types.EB_DoorOpen:
		pair := request.RequestsChooseDirection(my_elevator)
		my_elevator.Dirn = pair.Dirn
		my_elevator.Behaviour = pair.Behaviour
		elevatorBehaviourChan <- my_elevator.Behaviour

		switch my_elevator.Behaviour {
		case types.EB_DoorOpen:
			timer.TimerStart(my_elevator.DoorOpenDuration_s)
			requestsToClear := request.RequestsToClearAtCurrentFloor(my_elevator)
			my_elevator = request.ClearRequests(requestsToClear, my_elevator)
			sendRequestToClear(requestsToClear)

		case types.EB_Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(my_elevator.Dirn)
			doorClosedChan <- true

		case types.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(my_elevator.Dirn)
			doorClosedChan <- true
		}
	}
}

func onFloorArrival(newFloor int) {
	my_elevator.Floor = newFloor
	elevio.SetFloorIndicator(my_elevator.Floor)

	switch my_elevator.Behaviour {
	case types.EB_Moving:
		if request.RequestsShouldStop(my_elevator) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			timer.TimerStart(my_elevator.DoorOpenDuration_s)
			elevio.SetDoorOpenLamp(true)
			requestsToClear := request.RequestsToClearAtCurrentFloor(my_elevator)
			my_elevator = request.ClearRequests(requestsToClear, my_elevator)
			sendRequestToClear(requestsToClear)
			my_elevator.Behaviour = types.EB_DoorOpen
			elevatorBehaviourChan <- my_elevator.Behaviour
		}
	}
}

func onRequestUpdate() {
	var btnEvent elevio.ButtonEvent
	switch my_elevator.Behaviour {
	case types.EB_DoorOpen:
		if floor, btnType := request.RequestsShouldClearImmediately(my_elevator); floor > -1 {
			timer.TimerStart(my_elevator.DoorOpenDuration_s)
			btnEvent.Floor = floor
			btnEvent.Button = btnType
			my_elevator = request.ClearSingleRequest(btnEvent, my_elevator)
			sendRequestToClear(btnEvent)
		}

	case types.EB_Idle:
		pair := request.RequestsChooseDirection(my_elevator)
		my_elevator.Dirn = pair.Dirn
		my_elevator.Behaviour = pair.Behaviour
		elevatorBehaviourChan <- my_elevator.Behaviour
		switch pair.Behaviour {
		case types.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.TimerStart(my_elevator.DoorOpenDuration_s)
			requestsToClear := request.RequestsToClearAtCurrentFloor(my_elevator)
			my_elevator = request.ClearRequests(requestsToClear, my_elevator)
			sendRequestToClear(requestsToClear)

		case types.EB_Moving:
			elevio.SetMotorDirection(my_elevator.Dirn)
		}
	}
}

func sendRequestToClear(req interface{}) {
	switch v := req.(type) {
	case elevio.ButtonEvent:
		requestFinishedChan <- v

	case []elevio.ButtonEvent:
		for _, r := range v {
			requestFinishedChan <- r
		}

	default:
		fmt.Println("There were no requests to clear")
	}
}

func Fsm_elevatorModule(hallOrdersChan chan types.OrderTable, cabOrdersChan chan []bool) {
	localport := "localhost:" + strconv.Itoa(config.ElevPort)
	elevio.Init(localport)

	my_elevator = types.Elevator_uninitialized()
	my_elevator.Behaviour = types.EB_Moving
	elevio.SetMotorDirection(elevio.MD_Up)

	drv_floors := make(chan int)
	drv_timer := make(chan bool)
	drv_obstr := make(chan bool)

	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollTimer(drv_timer)
	go elevio.PollObstructionSwitch(drv_obstr)
	go fault.Faults_totalObstruction(doorClosedChan, doorObstructionChan)
	go fault.Faults_motorStop(elevatorBehaviourChan)

	for {
		select {
		case hallOrders := <-hallOrdersChan:
			hallOrdersToRequests(hallOrders)
			onRequestUpdate()

		case cabOrders := <-cabOrdersChan:
			cabOrderToRequests(cabOrders)
			onRequestUpdate()

		case a := <-drv_floors:
			onFloorArrival(a)

		case <-drv_timer:
			timer.TimerStop()
			onDoorTimeout()

		case a := <-drv_obstr:
			if a {
				obstruction()
			}
		}
		sendElevDataChan <- true
	}
}

func SendDataToDP(elevDataChan chan<- types.Elevator, orderFinishedChan chan<- elevio.ButtonEvent) {
	for {
		select {
		case finishedRequest := <-requestFinishedChan:
			orderFinishedChan <- finishedRequest

		case <-sendElevDataChan:
			elevDataChan <- my_elevator
		}
	}
}
