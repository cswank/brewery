package brewgadgets

import (
	"bitbucket.com/cswank/gogadgets"
	"log"
)


type HLT struct {
	gogadgets.InputDevice
	GPIO gogadgets.Poller
	Value float64
	Units string
	Volume float64
	startVolume float64
	out chan<- gogadgets.Value
}

func NewHLT(pin *gogadgets.Pin) (gogadgets.InputDevice, error) {
	var err error
	var h *HLT
	pin.Edge = "rising"
	gpio, err := gogadgets.NewGPIO(pin)
	if err == nil {
		h = &HLT{
			GPIO:gpio.(gogadgets.Poller),
			Value: pin.Value.(float64),
			Units: pin.Units,
		}
	}
	return h, err
}

func (h *HLT) Start(in <-chan gogadgets.Message, out chan<- gogadgets.Value) {
	h.out = out
	value := make(chan float64)
	err := make(chan error)
	go h.wait(value, err)
	for {
		select {
		case msg := <- in:
			h.readMessage(msg)
		case val := <-value:
			h.startVolume = 0.0
			h.Volume = val
			h.sendValue()
		case e := <-err:
			log.Println(e)
		}
	}
}

func (h *HLT) GetValue() *gogadgets.Value {
	return &gogadgets.Value{
		Value: h.Volume,
		Units: h.Units,
	}
}

func (h *HLT) wait(out chan<- float64, err chan<- error) {
	for {
		val, e := h.GPIO.Wait()
		if e != nil {
			err<- e
		} else {
			if val {
				out<- h.Value
			} else {
				out<- 0.0
			}
		}
	}
}

func (h *HLT) readMessage(msg gogadgets.Message) {
	if msg.Sender == "mash volume" {
		if h.startVolume == 0.0 {
			h.startVolume = h.Volume
		}
		h.Volume = h.startVolume - msg.Value.Value.(float64)
		h.sendValue()
	}
}

func (h *HLT) sendValue() {
	h.out<- gogadgets.Value{
		Value: h.Volume,
		Units: h.Units,
	}
}
