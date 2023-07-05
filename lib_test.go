package ensure

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

type testPrinter struct {
	entries []string
}

func (t *testPrinter) Write(s string) {
	t.entries = append(t.entries, s)
}

func (t *testPrinter) Output() string {
	var output string
	for _, entry := range t.entries {
		output += entry + "\n"
	}
	return output
}

func init() {
	Printer = &testPrinter{}
}

func TestFullScenario(t *testing.T) {
	var (
		testCtx      = context.Background()
		aThing       bool
		anotherThing bool
		tornDown     = false
	)

	That("A full scenario runs as expected", func(s *Scenario) {
		s.Given("a thing is false", func() {
			aThing = false
		})

		s.And("another thing is true", func() {
			anotherThing = true
		}).Teardown("revert anotherThing", testCtx, func(ctx context.Context) {
			anotherThing = false
		})

		s.When("I do the old swaperoo", func() {
			aThing = true
			anotherThing = false
		})

		s.Then("the a thing should be true", func() {
			require.Equal(t, true, aThing)
		})

		s.And("anotherThing should be false", func() {
			require.Equal(t, false, anotherThing)
		}).Teardown("tearDown", testCtx, func(ctx context.Context) {
			tornDown = true
		})
	}, t)

	require.True(t, tornDown)

	require.Equal(t, `Scenario: A full scenario runs as expected
 Given a thing is false
  And another thing is true
 When I do the old swaperoo
 Then the a thing should be true
  And anotherThing should be false
Tearing down revert anotherThing
Tearing down tearDown
`, Printer.(*testPrinter).Output(), "output should be as expected")
}
