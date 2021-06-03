package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	rpio "github.com/stianeikeland/go-rpio"
)

type options struct {
	mode     string
	interval uint8
}

var validModes = map[string]bool{
	"red":    true,
	"yellow": true,
	"green":  true,
	"random": true,
}

func getOptions() (options, error) {
	mode := flag.String("mode", "random", "Which traffic light to turn on. Can be one of the following: red,yellow,green,random")
	interval := flag.Int("interval", 5, "The interval at which the light will flash at a rate of (input * 100)ms. If 0, the light will just stay on")

	var intervalMin int

	flag.Parse()

	_, ok := validModes[*mode]
	if !ok {
		return options{}, errors.New("mode can only be one of the following: red,yellow,green,random")
	}

	if *mode == "random" {
		intervalMin = 1
	}

	if *interval < intervalMin || *interval > 100 {
		return options{}, errors.New("interval can only be between 0 to 100, for random mode interval must be at least 1.")
	}

	return options{
		mode:     *mode,
		interval: uint8(*interval * 100),
	}, nil
}

func main() {
	// Shows useful information when user enters --help option
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] \nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	options, err := getOptions()
	check(err)

	err = rpio.Open()
	check(err)

	// Get the pin for each of the lights
	redPin := rpio.Pin(9)
	yellowPin := rpio.Pin(10)
	greenPin := rpio.Pin(11)

	// Set the pins to output mode
	redPin.Output()
	yellowPin.Output()
	greenPin.Output()

	// Clean up on ctrl-c and turn lights out
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		redPin.Low()
		yellowPin.Low()
		greenPin.Low()
		fmt.Println("\nShutting down...")
		os.Exit(0)
	}()

	// Close the pins at the end of the program execution
	defer rpio.Close()

	// Turn lights off to start.
	redPin.Low()
	yellowPin.Low()
	greenPin.Low()

	// Execute behavior based on mode option
	switch options.mode {
	case "red":
		for {
			flashPin(redPin, options.interval)
		}
	case "yellow":
		for {
			flashPin(yellowPin, options.interval)
		}
	case "green":
		for {
			flashPin(greenPin, options.interval)
		}
	case "random":
		pins := []rpio.Pin{
			redPin,
			yellowPin,
			greenPin,
		}
		for {
			pin := pins[getBoundedRandNum(0, len(pins)-1)]
			flashPin(pin, options.interval)
		}
	}
}

func flashPin(pin rpio.Pin, interval uint8) {
	pin.High()
	if interval > 0 {
		time.Sleep(time.Millisecond * time.Duration(interval))
		pin.Low()
		time.Sleep(time.Millisecond * time.Duration(interval))
	}
}

func getBoundedRandNum(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

func check(e error) {
	if e != nil {
		exitGracefully(e)
	}
}
func exitGracefully(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
