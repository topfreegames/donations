package cmd_test

import (
	"fmt"

	. "github.com/topfreegames/donations/cmd"
	. "github.com/topfreegames/donations/metadata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version Command", func() {
	BeforeEach(func() {
		MockStdout()
	})
	AfterEach(func() {
		ResetStdout()
	})

	It("Should write proper version", func() {
		VersionCmd.Run(VersionCmd, []string{})
		ResetStdout()
		result := ReadStdout()
		Expect(result).To(BeEquivalentTo(fmt.Sprintf("donations v%s\n", GetVersion())))
	})
})
