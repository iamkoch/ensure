# Ensure

A scenario-based test runner for Go.

Every step runs as its own `t.Run` subtest, so `go test -v` prints the scenario as a tree and a
failure names the step that failed rather than the whole test.

## Installation

```bash
go get github.com/iamkoch/ensure/v2
```

## Basic scenario

```go
package myapp

import (
	"testing"

	"github.com/iamkoch/ensure/v2"
	"github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
	ensure.That("a thing can be set to true", func(s *ensure.Scenario) {
		var aThing bool

		s.Given("a thing is false", func(t *testing.T) {
			aThing = false
		})

		s.When("I set a thing to true", func(t *testing.T) {
			aThing = true
		})

		s.Then("the thing should be true", func(t *testing.T) {
			require.True(t, aThing)
		})
	}, t)
}
```

Each step function takes the subtest's own `*testing.T`. Assert against that one, not against the
`*testing.T` of the enclosing test function, or the failure is attributed to the wrong step.

Declare the variables the steps share inside the scenario function, as above. Declaring them in the
enclosing test function works too, but scoping them to the scenario stops two scenarios in one test
sharing state by accident.

## Output

```
--- PASS: TestFullScenario (0.00s)
    --- PASS: TestFullScenario/Scenario__A_full_scenario_runs_as_expected (0.00s)
        --- PASS: TestFullScenario/Scenario__A_full_scenario_runs_as_expected/Given_a_thing_is_false (0.00s)
        --- PASS: TestFullScenario/Scenario__A_full_scenario_runs_as_expected/And_another_thing_is_true (0.00s)
        --- PASS: TestFullScenario/Scenario__A_full_scenario_runs_as_expected/When_I_do_the_old_swaperoo (0.00s)
        --- PASS: TestFullScenario/Scenario__A_full_scenario_runs_as_expected/Then_the_a_thing_should_be_true (0.00s)
        --- PASS: TestFullScenario/Scenario__A_full_scenario_runs_as_expected/And_anotherThing_should_be_false (0.00s)
        --- PASS: TestFullScenario/Scenario__A_full_scenario_runs_as_expected/Teardown_of_revert_anotherThing (0.00s)
        --- PASS: TestFullScenario/Scenario__A_full_scenario_runs_as_expected/Teardown_of_tearDown (0.00s)
```

Because steps are subtests, `go test -run` selects them individually and any tool that reads Go test
output reports each step in its own right.

## Steps

| Step | Use it for |
|---|---|
| `Background` | Setup that is not part of the behaviour being described: starting a container, seeding a database. |
| `Given` | The state the scenario starts from. |
| `When` | The action under test. |
| `Then` | The assertion. |
| `And` | A continuation of the step before it. Valid after any other step. |
| `Teardown` | Cleanup. Chain it onto the step that created the thing being cleaned up. |

## Background, And, Teardown

```go
package myapp

import (
	"context"
	"testing"

	"github.com/iamkoch/ensure/v2"
	"github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
	ctx := context.Background()

	ensure.That("a thing can be set to true", func(s *ensure.Scenario) {
		var aThing bool

		s.Background("prepare the scenario", func(t *testing.T) {
			// start a container, open a connection
		})

		s.Given("a thing is false", func(t *testing.T) {
			aThing = false
		})

		s.And("another thing happens", func(t *testing.T) {
			// ...
		})

		s.When("I set a thing to true", func(t *testing.T) {
			aThing = true
		})

		s.Then("the thing should be true", func(t *testing.T) {
			require.True(t, aThing)
		}).Teardown("revert a thing", ctx, func(ctx context.Context) {
			aThing = false
		})
	}, t)
}
```

`Teardown` takes a context and passes it to the teardown function untouched. Cancelling it is the
caller's business; the scenario does not cancel it.

Teardown functions run after the scenario's last step, whether the scenario passed or failed, and in
the order they were registered.

## Unwritten scenarios

`NotImplemented` fails the scenario immediately, so a scenario you have named but not yet written
does not sit in the suite silently passing.

```go
ensure.That("a expired licence is rejected", func(s *ensure.Scenario) {
	s.NotImplemented()
}, t)
```

## Step failures do not stop the scenario

A failing step does not skip the steps after it, in the same way that one failing `t.Run` does not
stop the next. A scenario whose `Given` fails still runs its `When` and `Then`, which can produce
several failures with one cause. Read the first failing step.

## Full example

See [`lib_test.go`](lib_test.go). It is the library's own test and it exercises every step type.

## Licence

MIT. See [LICENSE](LICENSE).
