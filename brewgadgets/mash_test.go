package brewgadgets

import (
	"testing"
	"bitbucket.org/cswank/gogadgets"
)

var (
	config = &MashConfig{
		TankRadius: 7.5 * 2.54,
		ValveRadius: 0.1875 * 2.54,
		Coefficient: 0.43244,
	}
)

func TestGetVolume(t *testing.T) {
	dev, _ := NewMash(config)
	mash := dev.(*Mash)
	volume := mash.GetVolume(5000.0, 13.0)
	if volume != 0.36460455256750146 {
		t.Error("incorrect volume", volume)
	}
}

func TestSystem(t *testing.T) {
	hltOut := make(chan gogadgets.Message)
	mashOut := make(chan gogadgets.Message)
	in := make(chan gogadgets.Value)

	poller := &FakePoller{}
	hlt := &HLT{
		GPIO: poller,
		Value: 25.0,
		Units: "liters",
	}
	
	dev, _ := NewMash(config)
	mash := dev.(*Mash)

	go hlt.Start(hltOut, in)
	go mash.Start(mashOut, in)

	val := <-in
	msg := gogadgets.Message{
		Sender: "hlt volume",
		Value: val,
	}
	mashOut<- msg
	
	msg = gogadgets.Message{
		Type: "update",
		Sender: "mash tun valve",
		Value: gogadgets.Value{
		Value: true,
		},
	}
	
	mashOut<- msg
	
	// for {
	// 	val := <- in
	// 	fmt.Println(val)
	// 	if val.Value.(float64) >= 0.55 {
	// 		msg = gogadgets.Message{
	// 			Type: "update",
	// 			Sender: "hlt valve",
	// 			Value: gogadgets.Value{
	// 				Value: false,
	// 			},
	// 		}
	// 		out<- msg
	// 		break
	// 	}
	// }
	// val := <-in
	// fmt.Println(val)
}
