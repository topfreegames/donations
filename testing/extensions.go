package testing

import "github.com/onsi/ginkgo"

//BeforeOnce runs the before each block only once
func BeforeOnce(beforeBlock func()) {
	hasRun := false

	ginkgo.BeforeEach(func() {
		if !hasRun {
			beforeBlock()
			hasRun = true
		}
	})
}
