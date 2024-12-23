package ensure

import (
	"context"
	"testing"
)

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

type Scenario struct {
	teardownMethods []tearDown
	t               *testing.T
}

func (s *Scenario) Given(stepName string, stepFunc func(t *testing.T)) *Scenario {
	s.t.Run("Given "+stepName, func(t *testing.T) {
		stepFunc(t)
	})
	return s
}

func (s *Scenario) And(stepName string, stepFunc func(t *testing.T)) *Scenario {
	s.t.Run("And "+stepName, func(t *testing.T) {
		stepFunc(t)
	})
	return s
}

func (s *Scenario) When(stepName string, stepFunc func(t *testing.T)) *Scenario {
	s.t.Run("When "+stepName, func(t *testing.T) {
		stepFunc(t)
	})
	return s
}

func (s *Scenario) Background(stepName string, stepFunc func(t *testing.T)) *Scenario {
	s.t.Run("Background of "+stepName, func(t *testing.T) {
		stepFunc(t)
	})
	return s
}

func (s *Scenario) Then(stepName string, stepFunc func(t *testing.T)) *Scenario {
	s.t.Run("Then "+stepName, func(t *testing.T) {
		stepFunc(t)
	})
	return s
}

// Teardown adds a function to be called when the scenario ends.
// The function is passed a context that is canceled when the scenario ends.
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

func (s *Scenario) NotImplemented() {
	s.t.Fatal("Not implemented")
}

type tearDown struct {
	name string
	f    func()
}
