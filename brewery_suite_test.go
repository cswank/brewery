package brewery_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBrewery(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Brewery Suite")
}
