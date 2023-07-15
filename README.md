# Ensure
A scenario-based test runner for Go

## Installation
```bash
go get github.com/iamkoch/ensure
```

## Basic scenario

```go
package myapp 

import (
    "testing"
    "github.com/iamkoch/ensure"
)

func TestExample(t *testing.T) {
    var aThing bool

    ensure.That("Testing a thing", func(s *ensure.Scenario) {
        s.Given("a thing is false", func() {
            aThing = false
        })
		
        s.When("I set a thing to true", func() {
            aThing = true
        })
		
        s.Then("the thing should be true", func() {
            if !aThing {
                t.Error("aThing should be true")
            }
        })
    }, t)
}

```

## Background, And, Teardown
This library also supports Background, And, and Teardown steps:

```go
package myapp

import (
	"testing"
	"context"
	"github.com/iamkoch/ensure"
)

func TestExample(t *testing.T) {
    var aThing bool

    ensure.That("Testing a thing", func(s *ensure.Scenario) {
        s.Background("Prepare the scenario", func() {
            // Do something here
        })
        
        s.Given("a thing is false", func() {
            aThing = false
        })
		
        s.And("another thing happens", func() {
            // Do another thing here
        })
		
        s.When("I set a thing to true", func() {
            aThing = true
        })
        
        s.Then("the thing should be true", func() {
            if !aThing {
                t.Error("aThing should be true")
            }
        }).Teardown("revert a thing", context.Background(), func(ctx context.Context) {
            aThing = false
        })
    }, t)
}

```

## Printing 
Scenario text is, by default, printed to stdout using `fmt.Println`. This can be overridden by setting the `Printer` variable to a custom implementation of the `Printer` interface.

This can be overridden by setting the `Printer` variable to a custom implementation of the `Printer` interface.

```go
package myapp

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
	ensure.Printer = &testPrinter{}
}

func TestMyThing(t *testing.T) {
	...
}
```



## Full Example
This is taken from our own test in `lib_test.go`:

```go
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

```
