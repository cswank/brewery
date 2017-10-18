package brewery

import (
	"github.com/cswank/gogadgets"
)

var (
	vol *volumeManager
)

type Config struct {
	//A and B are the slope of the calibration table
	//as in y = a + bx
	//where y = time(s) and x = volume(mL)
	A               float64
	B               float64
	HLTCapacity     float64
	BoilerFillTime  int //time to drain the mash in seconds
	FloatSwitchPin  string
	FloatSwitchPort string
}

func New(cfg *Config, opts ...func(*volumeManager)) (*Tank, *Tank, *Tank, *Tank, error) {
	var err error
	vol, err = newVolumeManager(cfg, opts...)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	hlt := NewTank("hlt", masterTank)
	tun := NewTank("tun")
	boiler := NewTank("boiler")
	carboy := NewTank("carboy")
	return hlt, tun, boiler, carboy, nil
}

func WithAfter(a Afterer) func(*volumeManager) {
	return func(v *volumeManager) {
		v.after = a
	}
}

func WithTimer(t Timer) func(*volumeManager) {
	return func(v *volumeManager) {
		v.timer = t
	}
}

func WithPoller(p gogadgets.Poller) func(*volumeManager) {
	return func(v *volumeManager) {
		v.poller = p
	}
}
