package brewery_test

import (
	"github.com/cswank/brewery"
	"github.com/cswank/gogadgets"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Carboy Volume", func() {
	var (
		out, in chan gogadgets.Message
		carboy  *brewery.Carboy
	)

	BeforeEach(func() {
		carboy, _ = brewery.NewCarboy()

		out = make(chan gogadgets.Message)
		in = make(chan gogadgets.Message)

		go carboy.Start(out, in)
		//capture the initial values from startup
		msg := <-in
		Expect(msg.Value.Value.(float64)).To(Equal(0.0))
	})

	It("keeps track of the carboy volume", func() {
		out <- gogadgets.Message{
			Type:   "update",
			Sender: "boiler volume",
			Value: gogadgets.Value{
				Value: 6.0,
			},
		}

		out <- gogadgets.Message{
			Type:   "update",
			Sender: "carboy pump",
			Value: gogadgets.Value{
				Value: false,
			},
		}

		msg := <-in
		Expect(msg.Value.Value.(float64)).To(Equal(6.0))
	})
})
