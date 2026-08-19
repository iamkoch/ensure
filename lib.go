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
//				require.True(t, thing)
//			})
//		}, t)
//	}
package ensure

import (
	"context"
	"testing"
)

// That runs one scenario. Steps run in the order they are declared, each as a
// subtest of the scenario. Teardown functions run after the last step, in the
// order they were registered.
//
// A step that fails does not stop the steps after it, in the same way that one
// failing t.Run does not stop the next.
func That(scenarioName string, scenarioFunc func(s *Scenario), t *testing.T) {
	t.Run("Scenario__"+scenarioName, func(t *testing.T) {
		scenario := &Scenario{t: t}
		scenarioFunc(scenario)
		for _, teardown := range scenario.teardownMethods {
			t.Run("Teardown of "+teardown.name, func(t *testing.T) {
				teardown.f()
			})
		}
	})
}

// Scenario collects the steps of one scenario. That supplies one to the
// scenario function; a Scenario built any other way has no *testing.T and
// panics on first use.
type Scenario struct {
	teardownMethods []tearDown
	t               *testing.T
}

// Given runs a step that establishes the state the scenario starts from.
func (s *Scenario) Given(stepName string, stepFunc func(t *testing.T)) *Scenario {
	s.t.Run("Given "+stepName, func(t *testing.T) {
		stepFunc(t)
	})
	return s
}

// And runs a step that continues from the step before it. Use it after any
// other step.
func (s *Scenario) And(stepName string, stepFunc func(t *testing.T)) *Scenario {
	s.t.Run("And "+stepName, func(t *testing.T) {
		stepFunc(t)
	})
	return s
}

// When runs the step that performs the action under test.
func (s *Scenario) When(stepName string, stepFunc func(t *testing.T)) *Scenario {
	s.t.Run("When "+stepName, func(t *testing.T) {
		stepFunc(t)
	})
	return s
}

// Background runs a setup step that is not part of the behaviour being
// described, such as starting a container or seeding a database.
func (s *Scenario) Background(stepName string, stepFunc func(t *testing.T)) *Scenario {
	s.t.Run("Background of "+stepName, func(t *testing.T) {
		stepFunc(t)
	})
	return s
}

// Then runs a step that asserts the outcome.
func (s *Scenario) Then(stepName string, stepFunc func(t *testing.T)) *Scenario {
	s.t.Run("Then "+stepName, func(t *testing.T) {
		stepFunc(t)
	})
	return s
}

// Teardown registers a function to run once the scenario's steps have
// finished. Chain it onto the step that created whatever needs cleaning up.
//
// ctx is passed through to teardownFunc untouched. Cancelling it is the
// caller's business; the scenario does not cancel it.
//
// Teardown functions run whether the scenario passed or failed, and in the
// order they were registered.
func (s *Scenario) Teardown(stepName string, ctx context.Context, teardownFunc func(ctx context.Context)) *Scenario {
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

// NotImplemented fails the scenario immediately. Use it to name a scenario you
// have not written yet without leaving it silently passing.
func (s *Scenario) NotImplemented() {
	s.t.Fatal("Not implemented")
}

type tearDown struct {
	name string
	f    func()
}
