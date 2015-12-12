package brewgadgets_test

import (
	"fmt"
	"time"

	"bitbucket.org/cswank/brewery/brewgadgets"

	"github.com/cswank/gogadgets"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

type FakePoller struct {
	gogadgets.Poller
	trigger <-chan bool
}

func (f *FakePoller) Wait() (bool, error) {
	v := <-f.trigger
	return v, nil
}

var _ = Describe("Volume", func() {
	var (
		trigger chan bool
		out, in chan gogadgets.Message
		poller  *FakePoller
		bv      *brewgadgets.BrewVolume
	)

	BeforeEach(func() {

		trigger = make(chan bool)
		poller = &FakePoller{
			trigger: trigger,
		}
		cfg := &brewgadgets.BrewConfig{
			MashRadius:      20.0,
			MashValveRadius: 10.0,
			Coefficient:     0.5,
			HLTCapacity:     26.5,
			Poller:          poller,
			BoilerFillTime:  1,
		}

		bv, _ = brewgadgets.NewBrewVolume(cfg)

		out = make(chan gogadgets.Message)
		in = make(chan gogadgets.Message)

		go bv.Start(out, in)

	})
	Describe("when all's good", func() {
		It("does everything", func() {
			out <- gogadgets.Message{
				Type:   "update",
				Sender: "hlt valve",
				Value: gogadgets.Value{
					Value: true,
				},
			}

			trigger <- true
			msg := <-in
			Expect(msg.Value.Value.(float64)).To(BeCloseTo(7.0))

			out <- gogadgets.Message{
				Type:   "update",
				Sender: "mash tun valve",
				Value: gogadgets.Value{
					Value: true,
				},
			}

			tunVolume := 0.0
			hltVolume := 0.0
			for tunVolume < 4.0 {
				msg = <-in
				if msg.Sender == "mash tun volume" {
					tunVolume = msg.Value.Value.(float64)
				} else if msg.Sender == "hlt volume" {
					hltVolume = msg.Value.Value.(float64)
				}
			}

			Expect(hltVolume + tunVolume).To(BeCloseTo(7.0))

			out <- gogadgets.Message{
				Type:   "update",
				Sender: "mash tun valve",
				Value: gogadgets.Value{
					Value: false,
				},
			}

			//clear out all messages
			var stop bool
			for !stop {
				select {
				case msg = <-in:
				case <-time.After(200 * time.Millisecond):
					stop = true
				}
			}

			out <- gogadgets.Message{
				Type: "command",
				Body: "update",
			}

			v := map[string]float64{}
			for len(v) < 3 {
				msg = <-in
				v[msg.Sender] = msg.Value.Value.(float64)
			}
			Expect(v["boiler volume"]).To(Equal(0.0))
			Expect(v["hlt volume"] + v["mash tun volume"]).To(BeCloseTo(7.0))

			out <- gogadgets.Message{
				Type:   "update",
				Sender: "boiler valve",
				Value: gogadgets.Value{
					Value: true,
				},
			}

			//wait for boiler volume update
			tunVolume = v["mash tun volume"]
			hltVolume = v["hlt volume"]
			v = map[string]float64{}
			for v["boiler volume"] == 0.0 {
				msg = <-in
				v[msg.Sender] = msg.Value.Value.(float64)
			}

			out <- gogadgets.Message{
				Type:   "update",
				Sender: "boiler valve",
				Value: gogadgets.Value{
					Value: false,
				},
			}

			//clear out all messages
			stop = false
			for !stop {
				select {
				case msg = <-in:
				case <-time.After(200 * time.Millisecond):
					stop = true
				}
			}

			Expect(v["boiler volume"]).To(Equal(tunVolume))
			Expect(v["mash tun volume"]).To(Equal(0.0))
			Expect(v["hlt volume"]).To(Equal(hltVolume))

			//fill hlt again
			out <- gogadgets.Message{
				Type:   "update",
				Sender: "hlt valve",
				Value: gogadgets.Value{
					Value: true,
				},
			}

			trigger <- true
			msg = <-in
			Expect(msg.Value.Value.(float64)).To(BeCloseTo(7.0))
			Expect(msg.Sender).To(Equal("hlt volume"))

			out <- gogadgets.Message{
				Type:   "update",
				Sender: "mash tun valve",
				Value: gogadgets.Value{
					Value: true,
				},
			}

			boilerVolume := v["boiler volume"]
			v = map[string]float64{}
			for v["mash tun volume"] <= 3.5 {
				msg = <-in
				v[msg.Sender] = msg.Value.Value.(float64)
			}

			out <- gogadgets.Message{
				Type:   "update",
				Sender: "mash tun valve",
				Value: gogadgets.Value{
					Value: false,
				},
			}

			//clear out all messages
			stop = false
			for !stop {
				select {
				case msg = <-in:
				case <-time.After(200 * time.Millisecond):
					stop = true
				}
			}

			out <- gogadgets.Message{
				Type: "command",
				Body: "update",
			}

			v = map[string]float64{}
			for len(v) != 3 {
				msg = <-in
				v[msg.Sender] = msg.Value.Value.(float64)
			}

			Expect(v["hlt volume"] + v["mash tun volume"]).To(BeCloseTo(7.0))
			Expect(v["boiler volume"]).To(Equal(boilerVolume))
		})
	})
})

type step struct {
}

type closeTo struct {
	expected float64
}

func BeCloseTo(expected interface{}) types.GomegaMatcher {
	return &closeTo{
		expected: expected.(float64),
	}
}

func (c *closeTo) Match(actual interface{}) (bool, error) {
	val := actual.(float64)
	a := c.expected-0.05 <= val && val <= c.expected+0.05
	var err error
	if !a {
		err = fmt.Errorf("%d is not close to %d", val, c.expected)
	}
	return a, err
}

func (c *closeTo) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto be close to\n\t%#v", actual, c.expected)
}

func (c *closeTo) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nnot to be close to\n\t%#v", actual, c.expected)
}
