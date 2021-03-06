package main

import (
	"./commhandler"
	"./fsm"
	"./go-nonblockingchan"
	"./orderhandler"
	"flag"
	"fmt"
	"os"
	"sync"
)

var id_ptr = flag.String("id", "noid", "ID for node")
var elevServerAddr_ptr = flag.String("addr", "localhost:15657", "Port for node")
var commonPort_ptr = flag.Int("bport", 20010, "Port for all broadcasts")

var wg sync.WaitGroup

const N_FLOORS = 4 //import
const N_BUTTONS = 3

func main() {
	flag.Parse()

	if *id_ptr == "noid" {
		fmt.Println("Specify id")
		os.Exit(1)
	}

	// Three modules in wait group
	wg.Add(3)

	// Channels: FSM -> OrderHandler
	elevatorStatusCh := nbc.New()              //make(chan fsm.Elevator)
	placedHallOrderCh := nbc.New()             //make(chan fsm.OrderEvent)
	completedHallOrdersThisElevCh := nbc.New() //make(chan []fsm.OrderEvent)

	// Channels: OrderHandler -> FSM
	addHallOrderCh := nbc.New()    //make(chan fsm.OrderEvent)
	deleteHallOrderCh := nbc.New() //make(chan fsm.OrderEvent)
	updateLightsCh := nbc.New()    //make(chan [N_FLOORS][N_BUTTONS]bool)

	// Channels: OrderHandler -> Network
	assignOrderCh := nbc.New()    //make(chan msgs.TakeOrderMsg)
	placedOrderCh := nbc.New()           //make(chan msgs.Order)
	completedOrderCh := nbc.New()        //make(chan msgs.Order)
	thisElevatorHeartbeatCh := nbc.New() //make(chan msgs.Heartbeat)

	// Channels: Network -> OrderHandler
	allElevatorsHeartbeatCh := nbc.New()   //make(chan []msgs.Heartbeat)
	redundantOrderCh := nbc.New()               //make(chan msgs.RedundantOrderMsg)
	takeOrderCh := nbc.New()           //make(chan msgs.TakeOrderMsg)
	downedElevatorsCh := nbc.New()         //make(chan []msgs.Heartbeat)
	completedHallOrderOtherElevCh := nbc.New() //make(chan msgs.Order)
	lastKnownOrdersCh := nbc.New() //make(chan msgs.Heartbeat)

	// Channels: Network -> FSM
	// (none)

	// FSM -> Network
	// (none)

	go commhandler.CommHandler(*id_ptr, *commonPort_ptr,
		thisElevatorHeartbeatCh, downedElevatorsCh, placedOrderCh,
		assignOrderCh, completedOrderCh,
		allElevatorsHeartbeatCh, takeOrderCh, redundantOrderCh,
		completedHallOrderOtherElevCh, lastKnownOrdersCh, &wg)

	go orderhandler.OrderHandler(*id_ptr,
		placedHallOrderCh, redundantOrderCh, takeOrderCh,
		completedHallOrdersThisElevCh, completedHallOrderOtherElevCh,
		downedElevatorsCh, elevatorStatusCh, allElevatorsHeartbeatCh,
		lastKnownOrdersCh,
		placedOrderCh, assignOrderCh, addHallOrderCh, completedOrderCh,
		deleteHallOrderCh, thisElevatorHeartbeatCh, updateLightsCh, &wg)

	go fsm.FSM(*elevServerAddr_ptr,
		addHallOrderCh, deleteHallOrderCh, updateLightsCh,
		placedHallOrderCh, completedHallOrdersThisElevCh, elevatorStatusCh,
		&wg)

	for {
		select {}
	}
}
