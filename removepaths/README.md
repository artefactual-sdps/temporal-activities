# removepaths

Removes all the given directory/file paths and any children they may contain
and returns any errors encountered. If a given path doesn't exist, no error
will be returned.

## Registration

The `Name` constant is used as example, use a different name to register and
execute the activity if that doesn't suit your needs. An example registration:

```go
import (
    "go.temporal.io/sdk/activity"
    "go.temporal.io/sdk/worker"

	"github.com/artefactual-sdps/temporal-activities/removepaths"
)

tw := worker.New(...)

tw.RegisterActivityWithOptions(
    removepaths.New().Execute,
    activity.RegisterOptions{Name: removepaths.Name},
)
```

## Execution

An example execution:

```go
import (
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "github.com/artefactual-sdps/temporal-activities/removepaths"
)

opts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    ScheduleToCloseTimeout: 5 * time.Minute,
    RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 1},
})

var re removepaths.Result
err := workflow.ExecuteActivity(
    opts,
    removepaths.Name,
    &removepaths.Params{
        Paths: []string{
            "/path/to/file.txt",
            "/path/to/dir",
            "/path/to/with/children",
            "/path/to/missing/dir/file.txt",
        },
    },
).Get(opts, &re)
```

`err` may contain any system error. It will try to delete all paths before
returning all errors joined.
