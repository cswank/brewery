package brewery

import (
	"math"

	"github.com/cswank/gogadgets"
)

var (
	vol *volumeManager
)

type Config struct {
	HLTRadius       float64
	TunValveRadius  float64
	HLTCoefficient  float64
	HLTCapacity     float64
	BoilerFillTime  int //time to drain the mash in seconds
	FloatSwitchPin  string
	FloatSwitchPort string
}

func (b *Config) getK() (float64, float64) {
	hltArea := math.Pi * math.Pow(b.HLTRadius, 2)
	valveArea := math.Pi * math.Pow(b.TunValveRadius, 2)
	g := 9.806 * 100.0 //centimeters
	x := math.Pow((2.0 / g), 0.5)
	return (hltArea * x) / (valveArea * b.HLTCoefficient), hltArea
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
