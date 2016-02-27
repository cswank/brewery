package brewery_test

import (
	"fmt"

	"github.com/cswank/brewery"
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

var _ = Describe("HLT Volume", func() {
	var (
		trigger chan bool
		out, in chan gogadgets.Message
		poller  *FakePoller
		hlt     *brewery.HLT
	)

	BeforeEach(func() {
		trigger = make(chan bool)
		poller = &FakePoller{
			trigger: trigger,
		}
		cfg := &brewery.HLTConfig{
			HLTCapacity: 7.0,
			Poller:      poller,
		}

		hlt, _ = brewery.NewHLT(cfg)

		out = make(chan gogadgets.Message)
		in = make(chan gogadgets.Message)

		go hlt.Start(out, in)
		//capture the initial values from startup
		msg := <-in
		Expect(msg.Value.Value.(float64)).To(Equal(0.0))
	})

	It("keeps track of hlt volume", func() {
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
			Sender: "tun valve",
			Value: gogadgets.Value{
				Value: true,
			},
		}

		out <- gogadgets.Message{
			Type:   "update",
			Sender: "tun volume",
			Value: gogadgets.Value{
				Value: 0.1,
			},
		}

		msg = <-in
		Expect(msg.Value.Value.(float64)).To(BeCloseTo(6.9))

		out <- gogadgets.Message{
			Type:   "update",
			Sender: "tun volume",
			Value: gogadgets.Value{
				Value: 0.3,
			},
		}

		msg = <-in
		Expect(msg.Value.Value.(float64)).To(BeCloseTo(6.7))
	})
})

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
