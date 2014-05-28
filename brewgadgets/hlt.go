package brewgadgets

import (
	"bitbucket.org/cswank/gogadgets/input"
	"bitbucket.org/cswank/gogadgets/output"
	"bitbucket.org/cswank/gogadgets/models"
	"log"
)

type HLT struct {
	input.InputDevice
	GPIO        input.Poller
	Value       float64
	Units       string
	Volume      float64
	startVolume float64
	out         chan<- models.Value
}

func NewHLT(pin *models.Pin) (input.InputDevice, error) {
	var err error
	var h *HLT
	pin.Edge = "rising"
	gpio, err := output.NewGPIO(pin)
	if err == nil {
		h = &HLT{
			GPIO:  gpio.(input.Poller),
			Value: pin.Value.(float64),
			Units: pin.Units,
		}
	}
	return h, err
}

func (h *HLT) Start(in <-chan models.Message, out chan<- models.Value) {
	h.out = out
	value := make(chan float64)
	err := make(chan error)
	go h.wait(value, err)
	for {
		select {
		case msg := <-in:
			h.readMessage(msg)
		case val := <-value:
			h.Volume = val
			h.startVolume = val
			h.sendValue()
		case e := <-err:
			log.Println(e)
		}
	}
}

func (h *HLT) GetValue() *models.Value {
	return &models.Value{
		Value: h.Volume,
		Units: h.Units,
	}
}

func (h *HLT) wait(out chan<- float64, err chan<- error) {
	for {
		val, e := h.GPIO.Wait()
		if e != nil {
			err <- e
		} else {
			if val {
				out <- h.Value
			} else {
				out <- 0.0
			}
		}
	}
}

func (h *HLT) readMessage(msg models.Message) {
	if msg.Sender == "mash tun volume" && msg.Value.Value.(float64) > 0.0 {
		h.Volume = h.startVolume - msg.Value.Value.(float64)
		h.sendValue()
	} else if msg.Sender == "mash tun volume" && msg.Value.Value.(float64) == 0.0 {
		h.startVolume = h.Volume
	}
}

func (h *HLT) sendValue() {
	h.out <- models.Value{
		Value: h.Volume,
		Units: h.Units,
	}
}
