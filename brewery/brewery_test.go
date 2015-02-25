package main

import (
	"math/rand"
	"testing"
	"time"

	"bitbucket.org/cswank/brewery/brewgadgets"
	"bitbucket.org/cswank/gogadgets"
)

func TestBrewery(t *testing.T) {
	app, spyChan, spyChanOut, trigger := getTestApp()
	stop := make(chan gogadgets.Message)
	go app.GoStart(stop)

	time.Sleep(100 * time.Millisecond)

	//the 3 valves should send an update on startup
	//because they are output gadgets
	i := 0
	for i < 3 {
		msg := <-spyChanOut
		if msg.Type != "update" || msg.Value.Value != false {
			t.Fatal(msg.Sender, msg.Type, msg.Value)
		}
		i += 1
	}

	spyChan <- gogadgets.Message{
		Type: "command",
		Body: "fill hlt to 7 gallons",
	}

	msg := <-spyChanOut
	if msg.Body != "fill hlt to 7 gallons" {
		t.Fatal(msg)
	}

	msg = <-spyChanOut
	if msg.Type != "update" || msg.Sender != "hlt valve" || msg.Value.Value != true {
		t.Fatal(msg)
	}

	trigger <- true

	msg = <-spyChanOut
	if msg.Type != "update" ||
		msg.Sender != "hlt volume" ||
		msg.Value.Value.(float64) != 7.0 {
		t.Fatal(msg)
	}

	trigger <- true

	spyChan <- gogadgets.Message{
		Type: "command",
		Body: "fill mash tun to 1 gallon",
	}

	for {
		msg = <-spyChanOut
		if msg.Sender == "mash tun valve" && msg.Type == "update" && msg.Value.Value == false {
			break
		}
	}

	//flush all messages out
	i = 0
	for i < 3 {
		<-spyChanOut
		i++
	}

	updates := getUpdates(spyChan, spyChanOut, 7)

	if updates["hlt valve"].Value.Value != false {
		t.Fatal("hlt valve should be off")
	}

	if updates["mash tun valve"].Value.Value != false {
		t.Fatal("mash tun valve should be off")
	}

	if updates["boiler valve"].Value.Value != false {
		t.Fatal("boiler valve should be off")

	}
	mtv := updates["mash tun volume"].Value.Value.(float64)
	hltv := updates["hlt volume"].Value.Value.(float64)
	if !isBetween(hltv, 5.97, 6.0) {
		t.Fatal("hlt volume", hltv)
	}
	if !isBetween(mtv, 1.00, 1.03) {
		t.Fatal("mash tun volume", mtv)
	}
	if hltv+mtv != 7.00 {
		t.Fatal("total volume", mtv+hltv)
	}

	spyChan <- gogadgets.Message{
		Type: "command",
		Body: "fill boiler to 1 gallon",
	}

	for {
		msg = <-spyChanOut
		if msg.Sender == "boiler valve" && msg.Type == "update" && msg.Value.Value == false {
			break
		}
	}

	//flush all messages out
	msg = <-spyChanOut
	if msg.Sender != "mash tun volume" || msg.Type != "update" || msg.Value.Value.(float64) != 0.0 {
		t.Fatal(msg)
	}

	updates = getUpdates(spyChan, spyChanOut, 8)

	if updates["hlt valve"].Value.Value != false {
		t.Fatal("hlt valve should be off")
	}

	if updates["mash tun valve"].Value.Value != false {
		t.Fatal("mash tun valve should be off")
	}

	if updates["boiler valve"].Value.Value != false {
		t.Fatal("boiler valve should be off")

	}
	hltv = updates["hlt volume"].Value.Value.(float64)
	if !isBetween(hltv, 5.97, 6.0) {
		t.Fatal("hlt volume", hltv)
	}
	if updates["mash tun volume"].Value.Value.(float64) != 0.0 {
		t.Fatal("mash tun volume", updates["mash tun volume"].Value.Value.(float64))
	}
	bv := updates["boiler volume"].Value.Value.(float64)
	if bv != mtv {
		t.Fatal("boiler volume", bv)
	}

	spyChan <- gogadgets.Message{
		Type: "command",
		Body: "fill hlt to 7 gallons",
	}

	msg = <-spyChanOut
	if msg.Body != "fill hlt to 7 gallons" {
		t.Fatal(msg)
	}

	msg = <-spyChanOut
	if msg.Type != "update" || msg.Sender != "hlt valve" || msg.Value.Value != true {
		t.Fatal(msg)
	}

	trigger <- true

	msg = <-spyChanOut
	if msg.Type != "update" ||
		msg.Sender != "hlt volume" ||
		msg.Value.Value.(float64) != 7.0 {
		t.Fatal(msg)
	}

	msg = <-spyChanOut
	if msg.Type != "update" ||
		msg.Sender != "hlt valve" ||
		msg.Value.Value != false {
		t.Fatal(msg)
	}
	trigger <- true //simulate the water level going down in the hlt

	updates = getUpdates(spyChan, spyChanOut, 8)

	if updates["hlt valve"].Value.Value != false {
		t.Fatal("hlt valve should be off")
	}

	if updates["mash tun valve"].Value.Value != false {
		t.Fatal("mash tun valve should be off")
	}

	if updates["boiler valve"].Value.Value != false {
		t.Fatal("boiler valve should be off")

	}
	hltv = updates["hlt volume"].Value.Value.(float64)
	if hltv != 7.0 {
		t.Fatal("hlt volume", hltv)
	}
	if updates["mash tun volume"].Value.Value.(float64) != 0.0 {
		t.Fatal("mash tun volume", updates["mash tun volume"].Value.Value.(float64))
	}
	bv = updates["boiler volume"].Value.Value.(float64)
	if bv != mtv {
		t.Fatal("boiler volume", bv)
	}

	spyChan <- gogadgets.Message{
		Type: "command",
		Body: "fill mash tun to 1 gallon",
	}

	for {
		msg = <-spyChanOut
		if msg.Sender == "mash tun valve" && msg.Type == "update" && msg.Value.Value == false {
			break
		}
	}

	//flush all messages out
	i = 0
	for i < 3 {
		<-spyChanOut
		i++
	}

	updates = getUpdates(spyChan, spyChanOut, 7)

	if updates["hlt valve"].Value.Value != false {
		t.Fatal("hlt valve should be off")
	}

	if updates["mash tun valve"].Value.Value != false {
		t.Fatal("mash tun valve should be off")
	}

	if updates["boiler valve"].Value.Value != false {
		t.Fatal("boiler valve should be off")

	}
	mtv = updates["mash tun volume"].Value.Value.(float64)
	hltv = updates["hlt volume"].Value.Value.(float64)
	if !isBetween(hltv, 5.97, 6.0) {
		t.Fatal("hlt volume", hltv)
	}
	if !isBetween(mtv, 1.00, 1.03) {
		t.Fatal("mash tun volume", mtv)
	}
	if hltv+mtv != 7.00 {
		t.Fatal("total volume", mtv+hltv)
	}

	spyChan <- gogadgets.Message{
		Type: "command",
		Body: "fill mash tun to 1.5 gallons",
	}

	for {
		msg = <-spyChanOut
		if msg.Sender == "mash tun valve" && msg.Type == "update" && msg.Value.Value == false {
			break
		}
	}

	//flush all messages out
	i = 0
	for i < 3 {
		<-spyChanOut
		i++
	}

	updates = getUpdates(spyChan, spyChanOut, 7)

	if updates["hlt valve"].Value.Value != false {
		t.Fatal("hlt valve should be off")
	}

	if updates["mash tun valve"].Value.Value != false {
		t.Fatal("mash tun valve should be off")
	}

	if updates["boiler valve"].Value.Value != false {
		t.Fatal("boiler valve should be off")

	}
	mtv = updates["mash tun volume"].Value.Value.(float64)
	hltv = updates["hlt volume"].Value.Value.(float64)
	if !isBetween(hltv, 5.45, 5.50) {
		t.Fatal("hlt volume", hltv)
	}
	if !isBetween(mtv, 1.50, 1.55) {
		t.Fatal("mash tun volume", mtv)
	}
}

func getUpdates(spyChan chan gogadgets.Message, spyChanOut chan gogadgets.Message, n int) map[string]gogadgets.Message {
	spyChan <- gogadgets.Message{
		Type: "command",
		Body: "update",
	}

	updates := map[string]gogadgets.Message{}
	i := 0
	for i < n {
		msg := <-spyChanOut
		if msg.Type == "update" {
			updates[msg.Sender] = msg
			i++
		}
	}
	return updates
}

func isBetween(x, low, high float64) bool {
	return x <= high && x >= low
}

func getTestApp() (*gogadgets.App, chan gogadgets.Message, chan gogadgets.Message, chan bool) {
	pubPort := 1024 + rand.Intn(65535-1024)
	subPort := pubPort + 1
	config := &gogadgets.Config{
		Host:    "localhost",
		PubPort: pubPort,
		SubPort: subPort,
	}
	trigger := make(chan bool)
	fakePoller := &FakePoller{trigger: trigger}
	mashConfig := &brewgadgets.MashConfig{
		TankRadius:  7.5 * 2.54,
		ValveRadius: 2.0 * 2.54,
		Coefficient: 0.43244,
		Interval:    10 * time.Millisecond,
	}
	boilerConfig := &brewgadgets.BoilerConfig{
		WaitTime: 100 * time.Millisecond,
	}
	app, err := getApp(config, mashConfig, fakePoller, boilerConfig)
	if err != nil {
		panic(err)
	}

	spyChan := make(chan gogadgets.Message, 100)
	spyChanOut := make(chan gogadgets.Message, 100)
	spy := &Spy{
		input:    spyChan,
		output:   spyChanOut,
		messages: map[string][]gogadgets.Message{}}
	app.AddGadget(spy)

	hltGPIO := &FakeOutput{}

	hltValve := &gogadgets.Gadget{
		Location:   "hlt",
		Name:       "valve",
		Output:     hltGPIO,
		Direction:  "output",
		OnCommand:  "fill hlt",
		OffCommand: "stop filling hlt",
		UID:        "hlt valve",
		Operator:   ">=",
	}

	app.AddGadget(hltValve)

	mashGPIO := &FakeOutput{}

	mashValve := &gogadgets.Gadget{
		Location:   "mash tun",
		Name:       "valve",
		Output:     mashGPIO,
		Direction:  "output",
		OnCommand:  "fill mash tun",
		OffCommand: "stop filling mash tun",
		UID:        "mash tun valve",
		Operator:   ">=",
	}

	app.AddGadget(mashValve)

	boilerGPIO := &FakeOutput{}

	boilerValve := &gogadgets.Gadget{
		Location:   "boiler",
		Name:       "valve",
		Output:     boilerGPIO,
		Direction:  "output",
		OnCommand:  "fill boiler",
		OffCommand: "stop filling boiler",
		UID:        "boiler valve",
		Operator:   ">=",
	}

	app.AddGadget(boilerValve)
	return app, spyChan, spyChanOut, trigger
}

type FakePoller struct {
	val     bool
	trigger chan bool
}

func (f *FakePoller) Wait() (bool, error) {
	<-f.trigger
	f.val = !f.val
	return f.val, nil
}

type FakeOutput struct {
	val    *gogadgets.Value
	status bool
}

func (f *FakeOutput) On(val *gogadgets.Value) error {
	f.status = true
	return nil
}

func (f *FakeOutput) Off() error {
	f.status = false
	return nil
}

func (f *FakeOutput) Update(msg *gogadgets.Message) {

}

func (f *FakeOutput) Status() interface{} {
	return f.status
}

func (f *FakeOutput) Config() gogadgets.ConfigHelper {
	return gogadgets.ConfigHelper{}
}

type Spy struct {
	messages map[string][]gogadgets.Message
	input    chan gogadgets.Message
	output   chan gogadgets.Message
}

func (s *Spy) GetUID() string {
	return "spy"
}

func (s *Spy) Start(input <-chan gogadgets.Message, output chan<- gogadgets.Message) {
	for {
		select {
		case msg := <-input:
			s.appendMsg(msg)
			s.output <- msg
		case msg := <-s.input:
			output <- msg
		}
	}
}

func (s *Spy) appendMsg(msg gogadgets.Message) {
	msgs, ok := s.messages[msg.Sender]
	if !ok {
		msgs = []gogadgets.Message{}
	}
	msgs = append(msgs, msg)
	s.messages[msg.Sender] = msgs
}
