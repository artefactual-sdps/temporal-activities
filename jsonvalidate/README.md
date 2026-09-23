# jsonvalidate

Validates a JSON document against a JSON Schema.

The activity checks that both files exist, then delegates validation to a
`JSONValidator`. This package does not ship a validator implementation.
Supply one when constructing the activity.

## Registration

The `Name` constant is used as example, use any name to register and execute
the activity that meets your needs.

```go
import (
    "go.temporal.io/sdk/activity"
    "go.temporal.io/sdk/worker"

    "github.com/artefactual-sdps/temporal-activities/jsonvalidate"
)

tw := worker.New(...)

tw.RegisterActivityWithOptions(
    jsonvalidate.New(validator).Execute,
    activity.RegisterOptions{Name: jsonvalidate.Name},
)
```

`validator` must implement `jsonvalidate.JSONValidator`.

## Execution

An example execution:

```go
import (
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "github.com/artefactual-sdps/temporal-activities/jsonvalidate"
)

opts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    ScheduleToCloseTimeout: 5 * time.Minute,
    RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 1},
})

var re jsonvalidate.Result
err := workflow.ExecuteActivity(
    opts,
    jsonvalidate.Name,
    &jsonvalidate.Params{
        JSONPath:   "/path/to/document.json",
        SchemaPath: "/path/to/schema.json",
    },
).Get(opts, &re)
```

`err` may contain any non validation error. `re.Failures` contains the
validator output as `[]string`.
