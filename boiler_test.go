package brewery_test

import (
	"time"

	"github.com/cswank/brewery"
	"github.com/cswank/gogadgets"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Boiler Volume", func() {
	var (
		out, in chan gogadgets.Message
		boiler  *brewery.Boiler
	)

	BeforeEach(func() {
		boiler, _ = brewery.NewBoiler(10 * time.Millisecond)

		out = make(chan gogadgets.Message)
		in = make(chan gogadgets.Message)

		go boiler.Start(out, in)
		//capture the initial values from startup
		msg := <-in
		Expect(msg.Value.Value.(float64)).To(Equal(0.0))
	})

	It("keeps track of the boiler volume", func() {
		out <- gogadgets.Message{
			Type:   "update",
			Sender: "tun volume",
			Value: gogadgets.Value{
				Value: 6.0,
			},
		}

		out <- gogadgets.Message{
			Type:   "update",
			Sender: "boiler valve",
			Value: gogadgets.Value{
				Value: true,
			},
		}

		msg := <-in
		Expect(msg.Value.Value.(float64)).To(Equal(6.0))

		out <- gogadgets.Message{
			Type:   "update",
			Sender: "carboy volume",
			Value: gogadgets.Value{
				Value: 6.0,
			},
		}

		msg = <-in
		Expect(msg.Value.Value.(float64)).To(Equal(0.0))
	})
})
