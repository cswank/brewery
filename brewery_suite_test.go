package brewery_test

import (
	"fmt"
	"time"

	"github.com/cswank/brewery"
	"github.com/cswank/gogadgets"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"testing"
)

func TestBrewery(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Brewery Suite")
}

type FakePoller struct {
	gogadgets.Poller
	trigger <-chan bool
}

func (f *FakePoller) Wait() (bool, error) {
	v := <-f.trigger
	return v, nil
}

type FakeAfter struct {
	trigger <-chan bool
}

func (f *FakeAfter) After(t time.Duration) <-chan time.Time {
	c := make(chan time.Time)
	go func() {
		<-f.trigger
		c <- time.Now()
	}()
	return c
}

type fakeTimer struct {
	i int
}

func (f *fakeTimer) Start() {
	f.i = 0
}

func (f *fakeTimer) Since() time.Duration {
	f.i += 1
	return time.Duration(f.i) * time.Second
}

var _ = Describe("Brewery", func() {
	var (
		poller                    *FakePoller
		afterTrigger, pollTrigger chan bool
		out, in                   map[string]chan gogadgets.Message
		after                     *FakeAfter
		timer                     *fakeTimer
		hlt, tun, boiler, carboy  *brewery.Tank
		cfg                       *brewery.Config
	)

	BeforeEach(func() {
		out = map[string]chan gogadgets.Message{
			"hlt":    make(chan gogadgets.Message),
			"tun":    make(chan gogadgets.Message),
			"boiler": make(chan gogadgets.Message),
			"carboy": make(chan gogadgets.Message),
		}

		in = map[string]chan gogadgets.Message{
			"hlt":    make(chan gogadgets.Message),
			"tun":    make(chan gogadgets.Message),
			"boiler": make(chan gogadgets.Message),
			"carboy": make(chan gogadgets.Message),
		}

		afterTrigger = make(chan bool)
		after = &FakeAfter{
			trigger: afterTrigger,
		}

		timer = &fakeTimer{}

		pollTrigger = make(chan bool)
		poller = &FakePoller{
			trigger: pollTrigger,
		}

		cfg = &brewery.Config{
			HLTCapacity:    7.0,
			HLTRadius:      10.0,
			TunValveRadius: 0.25,
			HLTCoefficient: 0.4,
		}

		var err error
		hlt, tun, boiler, carboy, err = brewery.New(cfg, brewery.WithAfter(after.After), brewery.WithTimer(timer), brewery.WithPoller(poller))
		Expect(err).To(BeNil())
	})

	Context("hlt", func() {

		BeforeEach(func() {

			go hlt.Start(out["hlt"], in["hlt"])
			//capture the initial values from startup
			msg := <-in["hlt"]
			Expect(msg.Value.Value.(float64)).To(Equal(0.0))
		})

		It("fills the hlt", func() {
			out["hlt"] <- gogadgets.Message{
				Type:   "update",
				Sender: "hlt valve",
				Value: gogadgets.Value{
					Value: true,
				},
			}

			//fill hlt
			pollTrigger <- true

			msg := <-in["hlt"]
			Expect(msg.Value.Value.(float64)).To(Equal(cfg.HLTCapacity))
		})
	})

	Context("tun", func() {

		BeforeEach(func() {
			go hlt.Start(out["hlt"], in["hlt"])
			go tun.Start(out["tun"], in["tun"])
			//capture the initial values from startup
			msg := <-in["hlt"]
			Expect(msg.Value.Value.(float64)).To(Equal(0.0))
			msg = <-in["tun"]
			Expect(msg.Value.Value.(float64)).To(Equal(0.0))
		})

		It("fills the tun", func() {
			out["hlt"] <- gogadgets.Message{
				Type:   "update",
				Sender: "hlt valve",
				Value: gogadgets.Value{
					Value: true,
				},
			}

			//fill hlt
			pollTrigger <- true

			msg := <-in["hlt"]
			Expect(msg.Value.Value.(float64)).To(Equal(cfg.HLTCapacity))

			out["hlt"] <- gogadgets.Message{
				Type:   "update",
				Sender: "tun valve",
				Value: gogadgets.Value{
					Value: true,
				},
			}

			afterTrigger <- true
			msg = <-in["hlt"]
			Expect(msg.Value.Value.(float64)).To(Equal(6.4904387056582635))
			msg = <-in["tun"]
			Expect(msg.Value.Value.(float64)).To(Equal(0.5095612943417365))

			afterTrigger <- true
			msg = <-in["hlt"]
			Expect(msg.Value.Value.(float64)).To(Equal(6.000131447292215))
			msg = <-in["tun"]
			Expect(msg.Value.Value.(float64)).To(Equal(0.9998685527077844))

			out["hlt"] <- gogadgets.Message{
				Type:   "update",
				Sender: "tun valve",
				Value: gogadgets.Value{
					Value: false,
				},
			}

			msg = <-in["hlt"]
			Expect(msg.Value.Value.(float64)).To(Equal(5.529078224901856))
			msg = <-in["tun"]
			Expect(msg.Value.Value.(float64)).To(Equal(1.4709217750981438))
		})
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
		err = fmt.Errorf("%f is not close to %f", val, c.expected)
	}
	return a, err
}

func (c *closeTo) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto be close to\n\t%#v", actual, c.expected)
}

func (c *closeTo) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nnot to be close to\n\t%#v", actual, c.expected)
}
