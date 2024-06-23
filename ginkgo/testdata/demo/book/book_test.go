package book

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func Add(m int, n int) int {
	return m + n
}

var _ = Describe("Testcase Book", func() {
	Context("Read Book", func() {
		It("Read two books", func() {
			By("this is case 1")
			Expect(Add(100, 100)).To(Equal(200))
		})
		It("Read one book", func() {
			By("this is case 2")
			Expect(Add(200, 100)).To(Equal(200))
		})
	})
	Context("Buy Book", func() {
		It("Buy one book", func() {
			By("this is case 3")
			Expect(200).To(Equal(200))
		})
	})
})
