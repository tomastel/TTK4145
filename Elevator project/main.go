package main

import (
	"elevator_project/config"
	"elevator_project/data/processing"
	"flag"
	"fmt"
	"log"
	"time"
)

func main() {
	var restore int

	flag.IntVar(&config.ID, "id", config.ID, "Id of this peer")
	flag.IntVar(&config.ElevPort, "port", config.ElevPort, "Port of this peer")
	flag.IntVar(&config.Resets, "resets", config.Resets, "Number of consecutive resets")
	flag.IntVar(&restore, "restore", 1, "Restoring cab calls")

	flag.Parse()

	if config.ID == -1 {
		log.Fatalf("error: no ID provided, terminating")
	}

	fmt.Println("port = ", config.ElevPort, ", resets = ", config.Resets, ", numfloors = ", config.NumFloors, ", restore = ", restore)
	fmt.Println("Initializing")
	for i := 0; i < 3; i++ {
		fmt.Printf(".")
		time.Sleep(100*time.Millisecond)
	}

	fmt.Println("")

	processing.InitDataModule(restore)
	for {select{}}
}
