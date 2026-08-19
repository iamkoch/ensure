package ensure

import (
	"context"
	"strings"
	"testing"
)

// equalStrings fails the test unless got holds the same strings as want, in the
// same order. The tests here compare the order steps and teardowns ran in, so
// the message prints both slices.
func equalStrings(t *testing.T, got, want []string, msg string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s\ngot:  %v\nwant: %v", msg, got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("%s\ngot:  %v\nwant: %v", msg, got, want)
		}
	}
}

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
			if !aThing {
				t.Error("aThing should be true after the swaperoo")
			}
		})

		s.And("anotherThing should be false", func(t *testing.T) {
			if anotherThing {
				t.Fatal("anotherThing should be false after the swaperoo")
			}
		}).Teardown("tearDown", testCtx, func(ctx context.Context) {
			tornDown = true
		})
	}, t)

	if !tornDown {
		t.Error("the teardown chained onto the last step should have run")
	}
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

	equalStrings(t, executed, []string{"given", "when"},
		"the steps after a failure must not run")
	equalStrings(t, recorder.skipped,
		[]string{"Then the outcome is asserted", "And and so is this one"},
		"the steps after a failure must be reported as skipped")
	if len(recorder.steps) != 4 {
		t.Fatalf("every step must still appear in the output, skipped rather than absent: got %d steps, want 4",
			len(recorder.steps))
	}
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

	equalStrings(t, order, []string{"client", "container"},
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

	if !tornDown {
		t.Error("cleanup must happen even when the scenario failed")
	}
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

	equalStrings(t, order, []string{"setup"},
		"a skipped step has nothing to undo, so its teardown must not be registered")
}
