package ginkgo

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func Add(m int, n int) int {
	return m + n
}

var _ = Describe("Testcase", func() {
	Context("cont", func() {
		It("demo test", func() {
			By("this is case 1")
			Expect(Add(100, 100)).To(Equal(200))
		})
		It("demo test2", func() {
			By("this is case 2")
			Expect(Add(200, 100)).To(Equal(200))
		})
	})
	Context("[cont3]", func() {
		It("[demo test3]", func() {
			By("this is case 3")
			Expect(200).To(Equal(200))
		})
	})
})
