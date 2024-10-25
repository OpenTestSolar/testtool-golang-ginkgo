package v1_empty

import (
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"
)

func Add(m int, n int) int {
	return m + n
}

var _ = Describe("Testcase v1", func() {
	// Context("context", func() {
	// 	It("it", func() {
	// 		Expect(Add(100, 100)).To(Equal(200))
	// 	})
	// })
})
