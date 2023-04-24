package timer

import (
	"time"
)

var timerEndTime float64
var timerActive bool

func getWallTime() float64 {
	return float64(time.Now().UnixNano()) / float64(time.Second)
}

func TimerStart(duration float64) {
	timerEndTime = getWallTime() + duration
	timerActive = true
}

func TimerStop() {
	timerActive = false
}

func TimerTimedOut() bool {
	return ((timerActive) && (getWallTime() > timerEndTime))
}
