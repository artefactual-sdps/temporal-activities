# ffvalidate

Identifies file formats in the given path and validates them against a
list of allowed formats.

## Registration

The `Name` constant is used as example, use a different name to register and
execute the activity if that doesn't suit your needs. An example registration:

```go
import (
    "go.temporal.io/sdk/activity"
    "go.temporal.io/sdk/worker"

    "github.com/artefactual-sdps/temporal-activities/ffvalidate"
)

tw := worker.New(...)
cfg := ffvalidate.Config{AllowlistPath: "/path/to/allowed_file_formats.csv"}

tw.RegisterActivityWithOptions(
    ffvalidate.New(cfg).Execute,
    activity.RegisterOptions{Name: ffvalidate.Name},
)
```

## Execution

An example execution:

```go
import (
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "github.com/artefactual-sdps/temporal-activities/ffvalidate"
)

ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    ScheduleToCloseTimeout: 5 * time.Minute,
    RetryPolicy: &temporal.RetryPolicy{MaximumAttempts: 1},
})

var result ffvalidate.Result
err := workflow.ExecuteActivity(
    ctx,
    ffvalidate.Name,
    &ffvalidate.Params{Path: "/path/to/files"},
).Get(ctx, &result)
```

`err` may contain any non validation error.
`result.Failures` contains a list of files that are not an allowed format.
