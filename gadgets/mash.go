package brewgadgets


type Mash struct {
	gogadgets.InputDevice
	volume float64
	hltVolume float64
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
	m.volumeChan := make(chan float64)
	for {
		select {
		case msg := <- in:
			h.readMessage(msg)
		case val := <-volumeChan:
			m.volume = val
			m.SendValue()
		}
	}
}

func (m *Mash) monitor(out chan<- float64) {
	startingVolume := m.hltVolume
	
}

func (m *Mash) readMessage(msg gogadgets.Message) {
	keepGoing = true
	if msg.Type == "update" && msg.Sender == "hlt valve" {
		if msg.Value.Value == true {
			go m.monitor(m.volumeChan)
		} else {
			
		}
	} else if msg.Type == "update" && msg.Sender == "hlt volume" {
		m.hltVolume == msg.Value.Value.(float64)
	} else if msg.Type == "command" && msg.Body == "fill mash tun" {
		//get the target volume, if present, so a message can be
		//send when the target volume is achieved.
	}
	return keepGoing
}

func (m *Mash) SendValue() {
	m.out<- gogadgets.Value{
		Value: m.Value,
		Units: m.Units,
	}
}


def __init__(self, tank_radius, valve_radius, coefficient=0.81):
tank_radius = tank_radius * 2.54  # convert to centimeters
valve_radius = valve_radius * 2.54 # convert to centimeters
self.coefficient = coefficient
self.tank_area = math.pi * tank_radius ** 2
valve_area = math.pi * valve_radius ** 2
g = 9.806 * 100.0 #centimeters
x = pow((2 / g), 0.5)
self.k = (self.tank_area * x) / (valve_area * coefficient)
self.valve_area = valve_area
self.x = x

func (m *Mash) getVolume(startingVolume, time) {
	startingVolume := startingVolume * 1000.0 //convert to cubic centimeters
	height := m.getHeight(startingVolume)
	dh := math.fabs(((time / self.k) ** 2) - (2 * (time / self.k) * pow(height, 0.5)))
	return (m.tankArea * dh) / 1000.0
}




    
def get_drain_time(self, starting_volume, volume):
volume = 1000 * volume #convert to cubic centimeters
starting_volume = starting_volume * 1000.0
height_diff = self._get_height(volume)
height = self._get_height(starting_volume)
h2 = height - height_diff
return (pow(height, 0.5) - pow(h2, 0.5)) * self.k
