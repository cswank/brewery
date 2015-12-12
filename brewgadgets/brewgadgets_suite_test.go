package brewgadgets_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBrewgadgets(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Brewgadgets Suite")
}
