package brewery

import (
	"time"

	"github.com/cswank/gogadgets"
)

type HLT struct {
	volumes  map[string]float64
	capacity float64
	isFull   chan bool
	poller   gogadgets.Poller
	out      chan<- gogadgets.Message
}

type HLTConfig struct {
	HLTCapacity float64
	Poller      gogadgets.Poller
}

func NewHLT(cfg *HLTConfig) (*HLT, error) {
	return &HLT{
		volumes: map[string]float64{
			"tun": 0.0,
			"hlt": 0.0,
		},
		isFull:   make(chan bool),
		capacity: cfg.HLTCapacity,
		poller:   cfg.Poller,
	}, nil
}

func (h *HLT) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
	h.out = out
	h.sendUpdate()

	for {
		select {
		case msg := <-input:
			h.readMessage(msg)
		case f := <-h.isFull:
			h.updateVolume(f)
		}
	}
}

func (h *HLT) GetDirection() string {
	return "input"
}

func (h *HLT) GetUID() string {
	return "hlt volume"
}

func (h *HLT) readMessage(msg gogadgets.Message) {
	if msg.Type == "command" && msg.Body == "update" {
		h.sendUpdate()
	} else if msg.Type == "update" && msg.Sender == "hlt valve" && msg.Value.Value == true {
		go h.waitForFloatSwitch()
	} else if msg.Type == "update" && msg.Sender == "tun volume" {
		h.readTunVolume(msg)
	}
}

func (h *HLT) readTunVolume(msg gogadgets.Message) {
	v := msg.Value.Value.(float64)
	diff := v - h.volumes["tun"]
	h.volumes["tun"] = v
	if diff > 0.0 {
		h.volumes["hlt"] -= diff
		h.sendUpdate()
	}
}

func (h *HLT) sendUpdate() {
	h.out <- gogadgets.Message{
		UUID:      gogadgets.GetUUID(),
		Sender:    "hlt volume",
		Location:  "hlt",
		Name:      "volume",
		Type:      "update",
		Timestamp: time.Now().UTC(),
		Value: gogadgets.Value{
			Value: h.volumes["hlt"],
			Units: "gallons",
		},
		Info: gogadgets.Info{
			Direction: "input",
		},
	}
}

func (h *HLT) waitForFloatSwitch() {
	f, _ := h.poller.Wait()
	h.isFull <- f
}

func (h *HLT) updateVolume(full bool) {
	if !full {
		return
	}
	h.volumes["hlt"] = h.capacity
	h.sendUpdate()
}
