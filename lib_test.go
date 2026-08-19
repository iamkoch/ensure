package ensure

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// recordingT stands in for *testing.T so that a step can be reported as failed
// without failing the test that is doing the checking.
type recordingT struct {
	t       *testing.T
	failOn  string
	steps   []string
	skipped []string
}

func (r *recordingT) Run(name string, stepFunc func(t *testing.T)) bool {
	r.steps = append(r.steps, name)

	r.t.Run(strings.ReplaceAll(name, " ", "_"), func(t *testing.T) {
		t.Cleanup(func() {
			if t.Skipped() {
				r.skipped = append(r.skipped, name)
			}
		})
		stepFunc(t)
	})

	return name != r.failOn
}

func (r *recordingT) Fatal(args ...any) { r.t.Fatal(args...) }

func TestAFailedStepSkipsTheStepsAfterIt(t *testing.T) {
	recorder := &recordingT{t: t, failOn: "When the action is performed"}
	scenario := &Scenario{t: recorder}

	var executed []string
	record := func(name string) func(*testing.T) {
		return func(t *testing.T) { executed = append(executed, name) }
	}

	scenario.Given("the state is set up", record("given"))
	scenario.When("the action is performed", record("when"))
	scenario.Then("the outcome is asserted", record("then"))
	scenario.And("and so is this one", record("and"))

	require.Equal(t, []string{"given", "when"}, executed,
		"the steps after a failure must not run")
	require.Equal(t, []string{"Then the outcome is asserted", "And and so is this one"},
		recorder.skipped, "the steps after a failure must be reported as skipped")
	require.Len(t, recorder.steps, 4,
		"every step must still appear in the output, skipped rather than absent")
}

func TestTeardownsRunInReverseOrderOfRegistration(t *testing.T) {
	ctx := context.Background()
	recorder := &recordingT{t: t}
	scenario := &Scenario{t: recorder}

	var order []string
	teardown := func(name string) func(context.Context) {
		return func(context.Context) { order = append(order, name) }
	}

	scenario.Background("the container starts", func(t *testing.T) {}).
		Teardown("stop the container", ctx, teardown("container"))
	scenario.Given("a client connects to it", func(t *testing.T) {}).
		Teardown("close the client", ctx, teardown("client"))

	scenario.runTeardowns(recorder)

	require.Equal(t, []string{"client", "container"}, order,
		"a resource must be released before whatever it was created from")
}

func TestTeardownsRunWhenAStepFailed(t *testing.T) {
	ctx := context.Background()
	recorder := &recordingT{t: t, failOn: "When it goes wrong"}
	scenario := &Scenario{t: recorder}

	tornDown := false

	scenario.Given("something is created", func(t *testing.T) {}).
		Teardown("destroy it", ctx, func(context.Context) { tornDown = true })
	scenario.When("it goes wrong", func(t *testing.T) {})

	scenario.runTeardowns(recorder)

	require.True(t, tornDown, "cleanup must happen even when the scenario failed")
}

func TestTeardownIsNotRegisteredForASkippedStep(t *testing.T) {
	ctx := context.Background()
	recorder := &recordingT{t: t, failOn: "Given the setup fails"}
	scenario := &Scenario{t: recorder}

	var order []string

	scenario.Given("the setup fails", func(t *testing.T) {}).
		Teardown("undo the setup", ctx, func(context.Context) {
			order = append(order, "setup")
		})
	scenario.When("this never runs", func(t *testing.T) {}).
		Teardown("undo what never happened", ctx, func(context.Context) {
			order = append(order, "never")
		})

	scenario.runTeardowns(recorder)

	require.Equal(t, []string{"setup"}, order,
		"a skipped step has nothing to undo, so its teardown must not be registered")
}
