package testing

import (
	"time"

	"github.com/onsi/ginkgo"
)

//MockClock abstracts time
type MockClock struct {
	Time int64
}

//GetUTCTime returns the mocked time
func (m *MockClock) GetUTCTime() time.Time {
	return time.Unix(m.Time, 0)
}

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
