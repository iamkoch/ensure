package ensure

import (
	"context"
	"fmt"
	"testing"
)

type ScenarioPrinter interface {
	Write(s string)
}

type fmtPrinter struct {
}

func (d fmtPrinter) Write(s string) {
	fmt.Println(s)
}

var (
	Printer ScenarioPrinter = &fmtPrinter{}
)

func That(s string, f func(s *Scenario), t *testing.T) {
	Printer.Write("Scenario: " + s)
	scn := &Scenario{
		t: t,
	}
	f(scn)
	for _, f2 := range scn.teardownMethods {
		Printer.Write("Tearing down " + f2.name)
		f2.f()
	}
}

type Scenario struct {
	teardownMethods []tearDown
	t               *testing.T
}

func (s2 *Scenario) Given(s string, f func()) *Scenario {
	Printer.Write(" Given " + s)
	f()
	return s2
}

func (s2 *Scenario) And(s string, f func()) *Scenario {
	Printer.Write("  And " + s)
	f()
	return s2
}

func (s2 *Scenario) When(s string, f func()) *Scenario {
	Printer.Write(" When " + s)
	f()
	return s2
}

func (s2 *Scenario) Background(s string, f func()) *Scenario {
	Printer.Write(" Background " + s)
	f()
	return s2
}

func (s2 *Scenario) Then(s string, f func()) *Scenario {
	Printer.Write(" Then " + s)
	f()
	return s2
}

// Teardown adds a function to be called when the scenario ends.
// The function is passed a context that is cancelled when the scenario ends.
func (s2 *Scenario) Teardown(s string, ctx context.Context, f func(ctx context.Context)) *Scenario {
	s2.addTearDown(s, func() {
		ctx.Done()
		f(ctx)
	})
	return s2
}

func (s2 *Scenario) addTearDown(s string, f func()) {
	s2.teardownMethods = append(s2.teardownMethods, tearDown{
		name: s,
		f:    f,
	})
}

func (s2 *Scenario) NotImplemented() {
	s2.t.Fatal("Not implemented")
}

type tearDown struct {
	name string
	f    func()
}
