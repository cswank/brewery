package brewgadgets


type Mash struct {
	gogadgets.InputDevice
	Value float64
	Units string
	out chan<- gogadgets.Value
	stop chan bool
}

func NewMash(pin *gogadgets.Pin) (gogadgets.InputDevice, error) {
	var err error
	return &Mash{
		Value: pin.Value.(float64),
		Units: pin.Units,
	}, nil
}

func (m *Mash) Start(in <-chan gogadgets.Message, out chan<- gogadgets.Value) {
	m.out = out
	m.volume := make(chan float64)
	for {
		select {
		case msg := <- in:
			h.readMessage(msg)
		case val := <-volume:
			m.Value = val
			m.SendValue()
		}
	}
}

func (m *Mash) monitor(
	
}

func (m *Mash) readMessage(msg gogadgets.Message) {
	keepGoing = true
	if msg.Type == "status" && msg.Sender == "hlt valve" {
		if msg.Value.Value == true {
			go m.monitor(m.volume)
		} else {
			
		}
	}
	return keepGoing
}

func (m *Mash) SendValue() {
	m.out<- gogadgets.Value{
		Value: m.Value,
		Units: m.Units,
	}
}
