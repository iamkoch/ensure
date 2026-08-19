// Package ensure is a scenario-based test runner for Go.
//
// A scenario is a named group of Given/When/Then steps. Every step runs as its
// own subtest, so "go test -v" prints the scenario as a tree and a failure
// names the step that failed rather than the whole test.
//
//	func TestSomething(t *testing.T) {
//		ensure.That("a thing can be set", func(s *ensure.Scenario) {
//			var thing bool
//
//			s.Given("the thing is false", func(t *testing.T) {
//				thing = false
//			})
//
//			s.When("the thing is set", func(t *testing.T) {
//				thing = true
//			})
//
//			s.Then("the thing is true", func(t *testing.T) {
//				if !thing {
//					t.Error("the thing should be true")
//				}
//			})
//		}, t)
//	}
//
// The scenario is a subtest of the test function and each step is a subtest of
// the scenario:
//
//	--- PASS: TestSomething (0.00s)
//	    --- PASS: TestSomething/a_thing_can_be_set (0.00s)
//	        --- PASS: TestSomething/a_thing_can_be_set/Given_the_thing_is_false (0.00s)
//	        --- PASS: TestSomething/a_thing_can_be_set/When_the_thing_is_set (0.00s)
//	        --- PASS: TestSomething/a_thing_can_be_set/Then_the_thing_is_true (0.00s)
//
// Go replaces the spaces in a subtest name with underscores, so keep scenario
// and step names short. The step keyword is the only text the library adds.
//
// Steps are sequential and each one assumes the state the one before it left
// behind, so a failing step skips the steps after it. Assertions that are
// independent of each other belong in one step, reported with t.Error rather
// than t.Fatal, so that all of them report.
//
// The library has no dependencies. Each step function takes the subtest's own
// *testing.T, so any assertion library works with it.
package ensure

import (
	"context"
	"testing"
)

// scenarioT is the part of *testing.T that a Scenario uses. It exists so that
// the skip-after-failure behaviour can be tested without a failing test.
type scenarioT interface {
	Run(name string, f func(t *testing.T)) bool
	Fatal(args ...any)
}

// That runs one scenario. Steps run in the order they are declared, each as a
// subtest of the scenario.
//
// Once a step fails, every step after it is skipped. Teardown functions still
// run, in reverse order of registration, so that a resource is released before
// whatever it was created from.
func That(scenarioName string, scenarioFunc func(s *Scenario), t *testing.T) {
	t.Run(scenarioName, func(t *testing.T) {
		scenario := &Scenario{t: t}
		scenarioFunc(scenario)
		scenario.runTeardowns(t)
	})
}

// Scenario collects the steps of one scenario. That supplies one to the
// scenario function; a Scenario built any other way has no *testing.T and
// panics on first use.
type Scenario struct {
	teardownMethods []tearDown
	t               scenarioT
	failed          bool
	lastStepRan     bool
}

// Given runs a step that establishes the state the scenario starts from.
func (s *Scenario) Given(stepName string, stepFunc func(t *testing.T)) *Scenario {
	return s.step("Given ", stepName, stepFunc)
}

// And runs a step that continues from the step before it. Use it after any
// other step.
func (s *Scenario) And(stepName string, stepFunc func(t *testing.T)) *Scenario {
	return s.step("And ", stepName, stepFunc)
}

// When runs the step that performs the action under test.
func (s *Scenario) When(stepName string, stepFunc func(t *testing.T)) *Scenario {
	return s.step("When ", stepName, stepFunc)
}

// Background runs a setup step that is not part of the behaviour being
// described, such as starting a container or seeding a database.
func (s *Scenario) Background(stepName string, stepFunc func(t *testing.T)) *Scenario {
	return s.step("Background ", stepName, stepFunc)
}

// Then runs a step that asserts the outcome.
func (s *Scenario) Then(stepName string, stepFunc func(t *testing.T)) *Scenario {
	return s.step("Then ", stepName, stepFunc)
}

func (s *Scenario) step(prefix, stepName string, stepFunc func(t *testing.T)) *Scenario {
	name := prefix + stepName

	if s.failed {
		s.lastStepRan = false
		s.t.Run(name, func(t *testing.T) {
			t.SkipNow()
		})
		return s
	}

	s.lastStepRan = true
	if !s.t.Run(name, stepFunc) {
		s.failed = true
	}
	return s
}

// Teardown registers a function to run once the scenario's steps have
// finished. Chain it onto the step that created whatever needs cleaning up.
//
// ctx is passed through to teardownFunc untouched. Cancelling it is the
// caller's business; the scenario does not cancel it.
//
// Teardown functions run in reverse order of registration, so a resource is
// released before whatever it was created from. A teardown chained onto a step
// that ran, whether that step passed or failed, will run. A teardown chained
// onto a step that was skipped will not, because there is nothing to undo.
func (s *Scenario) Teardown(stepName string, ctx context.Context, teardownFunc func(ctx context.Context)) *Scenario {
	if !s.lastStepRan {
		return s
	}
	s.addTearDown(stepName, func() {
		teardownFunc(ctx)
	})
	return s
}

func (s *Scenario) addTearDown(name string, f func()) {
	s.teardownMethods = append(s.teardownMethods, tearDown{
		name: name,
		f:    f,
	})
}

func (s *Scenario) runTeardowns(t scenarioT) {
	for i := len(s.teardownMethods) - 1; i >= 0; i-- {
		teardown := s.teardownMethods[i]
		t.Run("Teardown "+teardown.name, func(t *testing.T) {
			teardown.f()
		})
	}
}

// NotImplemented fails the scenario immediately. Use it to name a scenario you
// have not written yet without leaving it silently passing.
func (s *Scenario) NotImplemented() {
	s.t.Fatal("Not implemented")
}

type tearDown struct {
	name string
	f    func()
}
