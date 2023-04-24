package request

import (
	"elevator_project/config"
	"elevator_project/data/types"
	"elevator_project/elevio"
)

type DirnBehaviourPair struct {
	Dirn      elevio.MotorDirection
	Behaviour types.ElevatorBehaviour
}

func requestsAbove(e types.Elevator) bool {
	for f := e.Floor + 1; f < config.NumFloors; f++ {
		for btn := 0; btn < config.NumButtons; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e types.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < config.NumButtons; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(e types.Elevator) bool {
	for btn := 0; btn < config.NumButtons; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func ClearSingleRequest(req elevio.ButtonEvent, e types.Elevator) types.Elevator {
	e.Requests[req.Floor][req.Button] = false
	return e
}

func ClearRequests(req []elevio.ButtonEvent, e types.Elevator) types.Elevator {
	for _, button := range req {
		e.Requests[button.Floor][button.Button] = false
	}
	return e
}

func RequestsChooseDirection(e types.Elevator) DirnBehaviourPair {
	switch e.Dirn {
	case elevio.MD_Up:
		if requestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, types.EB_Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Down, types.EB_DoorOpen}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, types.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, types.EB_Idle}
		}

	case elevio.MD_Down:
		if requestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, types.EB_Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Up, types.EB_DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, types.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, types.EB_Idle}
		}

	case elevio.MD_Stop:
		if requestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Stop, types.EB_DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, types.EB_Moving}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, types.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, types.EB_Idle}
		}

	default:
		return DirnBehaviourPair{elevio.MD_Stop, types.EB_Idle}
	}
}

func RequestsShouldClearImmediately(e types.Elevator) (int, elevio.ButtonType) {
	for btnType := elevio.ButtonType(0); btnType < elevio.ButtonType(config.NumButtons); btnType++ {
		if e.Requests[e.Floor][btnType] && ((e.Dirn == elevio.MD_Up && btnType == elevio.BT_HallUp) || (e.Dirn == elevio.MD_Down && btnType == elevio.BT_HallDown) || btnType == elevio.BT_Cab || e.Dirn == elevio.MD_Stop) {
			return e.Floor, btnType
		}
	}
	return -1, elevio.BT_HallUp
}

func RequestsShouldStop(e types.Elevator) bool {
	switch e.Dirn {
	case elevio.MD_Down:
		if (e.Requests[e.Floor][elevio.BT_HallDown]) || (e.Requests[e.Floor][elevio.BT_Cab]) || (!requestsBelow(e)) {
			return true
		}

	case elevio.MD_Up:
		if (e.Requests[e.Floor][elevio.BT_HallUp]) || (e.Requests[e.Floor][elevio.BT_Cab]) || !requestsAbove(e) {
			return true
		}

	case elevio.MD_Stop:
		fallthrough

	default:
		return true
	}
	return false
}

func RequestsToClearAtCurrentFloor(e types.Elevator) []elevio.ButtonEvent {
	buttonsToClear := make([]elevio.ButtonEvent, 0)

	if e.Requests[e.Floor][elevio.BT_Cab] {
		buttonsToClear = append(buttonsToClear, elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_Cab})
	}

	switch e.Dirn {
	case elevio.MD_Up:
		if !requestsAbove(e) && e.Requests[e.Floor][elevio.BT_HallDown] {
			buttonsToClear = append(buttonsToClear, elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown})
		}
		if e.Requests[e.Floor][elevio.BT_HallUp] {
			buttonsToClear = append(buttonsToClear, elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp})
		}

	case elevio.MD_Down:
		if !requestsBelow(e) && e.Requests[e.Floor][elevio.BT_HallUp] {
			buttonsToClear = append(buttonsToClear, elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp})
		}
		if e.Requests[e.Floor][elevio.BT_HallDown] {
			buttonsToClear = append(buttonsToClear, elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown})
		}

	case elevio.MD_Stop:
		if e.Requests[e.Floor][elevio.BT_HallUp] {
			buttonsToClear = append(buttonsToClear, elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp})
		}
		if e.Requests[e.Floor][elevio.BT_HallDown] {
			buttonsToClear = append(buttonsToClear, elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown})
		}

	default:
		if e.Requests[e.Floor][elevio.BT_HallUp] {
			buttonsToClear = append(buttonsToClear, elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp})
		}
		if e.Requests[e.Floor][elevio.BT_HallDown] {
			buttonsToClear = append(buttonsToClear, elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown})
		}
	}
	return buttonsToClear
}
