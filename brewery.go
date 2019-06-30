package brewery

import (
	"github.com/cswank/gogadgets"
)

var (
	vol *volumeManager
)

// Config holds brewery env vars
type Config struct {
	//A, B and C are the coeffcients of a polynomial curve
	//for water flowing out of the hlt into the mash
	//tun, as in y = a + bx + cx^2
	//where x = time (s) and y = vol (ml)
	A           float64
	B           float64
	C           float64
	HLTCapacity float64

	//BoilerFIllTime is the time to drain the mash in seconds
	BoilerFillTime  int
	FloatSwitchPin  string
	FloatSwitchPort string
}

func New(cfg *Config, opts ...func(*volumeManager)) (*Tank, *Tank, *Tank, *Tank, error) {
	var err error
	vol, err = newVolumeManager(cfg, opts...)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return NewTank("hlt", masterTank), NewTank("tun"), NewTank("boiler"), NewTank("carboy"), nil
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
