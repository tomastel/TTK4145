package fault

import (
	"elevator_project/config"
	"elevator_project/network/bcast"
	"elevator_project/data/types"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
) 

func rebootSystem() {
	config.Resets++
	idString := "-id="+strconv.Itoa(config.ID)
	resetString := "-resets="+strconv.Itoa(config.Resets)

	if config.Resets < 5 {
		err := exec.Command("gnome-terminal", "--", "go", "run", "main.go", idString, "-port=15657", resetString).Run()
		if err != nil {
			fmt.Println("Failed to reboot, error: ", err, ". Shutting down elevator")
			os.Exit(0)
		}
	} else {log.Fatalf("I have reset too many times, shutting down elevator")}
	os.Exit(0)
}

func Faults_motorStop(elevBehaviour <-chan types.ElevatorBehaviour) {
	alreadyStarted := false
	var start time.Time
	ticker := time.NewTicker(100*time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case behaviour := <-elevBehaviour:
			if behaviour == types.EB_Moving {
				if !alreadyStarted  {
					alreadyStarted = true
					start = time.Now()
				}
			} else {
				bcast.UpdateOnlineStatus(true)
				alreadyStarted = false
			} 

		case <-ticker.C:
			if alreadyStarted && time.Since(start) > config.MotorStopTimer {
				fmt.Println("Motor failure, rebooting elevator")
				alreadyStarted = false
				bcast.UpdateOnlineStatus(false)
			}
		}
	}
}

func Faults_totalObstruction(doorClosedChan <-chan bool, doorObstructionChan <-chan bool) {
	alreadyStarted := false
	onlineStatus := true
	var start time.Time
	timeTol := 5*time.Second
	ticker := time.NewTicker(100*time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-doorClosedChan:
			if !onlineStatus {
				fmt.Println("Door closed, reconnecting")
				onlineStatus = true
				bcast.UpdateOnlineStatus(onlineStatus)
			}
			alreadyStarted = false

		case <-doorObstructionChan:
			if !alreadyStarted {
				alreadyStarted = true
				start = time.Now()
			}

		case <-ticker.C:
			if alreadyStarted && time.Since(start) > timeTol && onlineStatus{
				fmt.Println("Obstruction, disconnecting from network")
				onlineStatus = false
				bcast.UpdateOnlineStatus(onlineStatus)
			}
		}
	}
}
