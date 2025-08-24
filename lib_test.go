package ensure

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFullScenario(t *testing.T) {
	var (
		testCtx      = context.Background()
		aThing       bool
		anotherThing bool
		tornDown     = false
	)

	That("A full scenario runs as expected", func(s *Scenario) {
		s.Given("a thing is false", func(t *testing.T) {
			aThing = false
		})

		s.And("another thing is true", func(t *testing.T) {
			anotherThing = true
		}).Teardown("revert anotherThing", testCtx, func(ctx context.Context) {
			anotherThing = false
		})

		s.When("I do the old swaperoo", func(t *testing.T) {
			aThing = true
			anotherThing = false
		})

		s.Then("the a thing should be true", func(t *testing.T) {
			assert.Equal(t, true, aThing)
		})

		s.And("anotherThing should be false", func(t *testing.T) {
			require.Equal(t, false, anotherThing)
		}).Teardown("tearDown", testCtx, func(ctx context.Context) {
			tornDown = true
		})
	}, t)

	require.True(t, tornDown)

}
