package book

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ginkgo Suite")
}
