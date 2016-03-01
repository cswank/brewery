package brewery_test

import (
	"time"

	"github.com/cswank/brewery"
	"github.com/cswank/gogadgets"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

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

var _ = Describe("HLT Volume", func() {
	var (
		trigger chan bool
		out, in chan gogadgets.Message
		after   *FakeAfter
		timer   *fakeTimer
		tun     *brewery.Tun
	)

	BeforeEach(func() {
		trigger = make(chan bool)
		after = &FakeAfter{
			trigger: trigger,
		}

		timer = &fakeTimer{}

		cfg := &brewery.TunConfig{
			HLTRadius:      10.0,
			TunValveRadius: 0.25,
			Coefficient:    0.4,
			After:          after.After,
			Timer:          timer,
		}

		tun, _ = brewery.NewTun(cfg)
		out = make(chan gogadgets.Message)
		in = make(chan gogadgets.Message)

		go tun.Start(out, in)
		//capture the initial values from startup
		msg := <-in
		Expect(msg.Value.Value.(float64)).To(Equal(0.0))
	})

	It("keeps track of tun volume", func() {
		out <- gogadgets.Message{
			Type:   "update",
			Sender: "hlt volume",
			Value: gogadgets.Value{
				Value: 7.0,
			},
		}

		out <- gogadgets.Message{
			Type:   "update",
			Sender: "tun valve",
			Value: gogadgets.Value{
				Value: true,
			},
		}

		trigger <- true
		msg := <-in
		Expect(msg.Value.Value.(float64)).To(Equal(0.008436018793712464))

		trigger <- true
		msg = <-in
		Expect(msg.Value.Value.(float64)).To(Equal(0.016866951206827928))

		out <- gogadgets.Message{
			Type:   "update",
			Sender: "tun valve",
			Value: gogadgets.Value{
				Value: false,
			},
		}

		msg = <-in
		Expect(msg.Value.Value.(float64)).To(Equal(0.025292797239346396))

		out <- gogadgets.Message{
			Type:   "update",
			Sender: "boiler volume",
			Value: gogadgets.Value{
				Value: 2.0,
			},
		}

		msg = <-in
		Expect(msg.Value.Value.(float64)).To(Equal(0.0))
	})
})
