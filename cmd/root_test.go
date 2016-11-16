package cmd_test

import (
	. "github.com/topfreegames/donations/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Root Command", func() {
	BeforeEach(func() {
		MockStdout()
	})
	AfterEach(func() {
		ResetStdout()
	})

	It("Should return help", func() {
		Execute(RootCmd)
		ResetStdout()
		result := ReadStdout()
		Expect(result).To(BeEquivalentTo("unknown flag: --test.timeout\n"))
	})
})
