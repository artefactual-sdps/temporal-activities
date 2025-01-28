# bagvalidate

Checks if a given directory is a valid BagIt Bag.

## Requirements

The activity doesn't have any extra requirements. However, the [bagit-gython]
validator used in the example below requires `glibc` and supports the following
operating systems and architectures:

- darwin-amd64
- darwin-arm64
- linux-amd64
- linux-arm64
- windows-amd64

## Registration

The `Name` constant is used as example, use a different name to register and
execute the activity if that doesn't suit your needs. An example registration
using the `bagit-gython` validator:

```go
import (
	bagit_gython "github.com/artefactual-labs/bagit-gython"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"

	"github.com/artefactual-sdps/temporal-activities/bagvalidate"
)

tw := worker.New(...)

validator, err := bagit_gython.NewBagIt()
if err != nil {
    // Handle error.
}
defer func() {
    if err = validator.Cleanup(); err != nil {
        // Handle error.
    }
}()

tw.RegisterActivityWithOptions(
    bagvalidate.New(validator).Execute,
    activity.RegisterOptions{Name: bagvalidate.Name},
)
```

## Execution

An example execution:

```go
import (
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "github.com/artefactual-sdps/temporal-activities/bagvalidate"
)

opts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    ScheduleToCloseTimeout: 5 * time.Minute,
    RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 1},
})

var re bagvalidate.Result
err := workflow.ExecuteActivity(
    opts,
    bagvalidate.Name,
    &bagvalidate.Params{
        Path: "/path/to/bag",
    },
).Get(opts, &re)
```

`err` may contain any non validation error. `re.Valid` will be true if the Bag
is valid and `re.Error` is a message indicating why validation failed, and will
always be empty when `re.Valid` is true.

[bagit-gython]: https://github.com/artefactual-labs/bagit-gython
