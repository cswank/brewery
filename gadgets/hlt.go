package brewgadgets


import (
	"bitbucket.com/cswank/gogadgets"
	"log"
)


type HLT struct {
	gogadgets.InputDevice
	GPIO gogadgets.Poller
	volume float64
	startVolume float64
	units string
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
			volume: pin.Value.(float64),
			units: pin.Units,
		}
	}
	return h, err
}

func (h *HLT) Start(in <-chan gogadgets.Message, out chan<- gogadgets.Value) {
	h.out = out
	value := make(chan float64)
	err := make(chan error)
	keepGoing := true
	for keepGoing {
		go h.wait(value, err)
		select {
		case msg := <- in:
			keepGoing = h.readMessage(msg)
		case val := <-value:
			h.startVolume = 0.0
			h.volume = val
			h.sendValue()
		case e := <-err:
			log.Println(e)
		}
	}
}

func (h *HLT) wait(out chan<- float64, err chan<- error) {
	val, e := h.GPIO.Wait()
	if e != nil {
		err<- e
	} else {
		if val {
			out<- h.volume
		} else {
			out<- 0.0
		}
	}
}

func (h *HLT) readMessage(msg gogadgets.Message) (keepGoing bool) {
	keepGoing = true
	if msg.Type == "command" && msg.Body == "shutdown" {
		keepGoing = false
	} else if msg.Type == "command" && msg.Body == "update" {
		h.sendValue()
	} else if msg.Type == "update" && msg.Body == "mash volume" {
		if h.startVolume == 0.0 {
			h.startVolume = h.volume
		}
		h.volume = h.startVolume - msg.Value.Value.(float64)
		h.sendValue()
	}
	return keepGoing
}

func (h *HLT) sendValue() {
	h.out<- gogadgets.Value{
		Value: h.volume,
		Units: h.units,
	}
}
