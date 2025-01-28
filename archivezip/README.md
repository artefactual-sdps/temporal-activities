# archivezip

Creates a Zip archive from a given directory. Allows setting the destination
path, if not set then the source directory path + ".zip" will be used.

## Registration

The `Name` constant is used as example, use a different name to register and
execute the activity if that doesn't suit your needs. An example registration:

```go
import (
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"

	"github.com/artefactual-sdps/temporal-activities/archivezip"
)

tw := worker.New(...)

tw.RegisterActivityWithOptions(
    archivezip.New().Execute,
    activity.RegisterOptions{Name: archivezip.Name},
)
```

## Execution

An example execution:

```go
import (
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "github.com/artefactual-sdps/temporal-activities/archivezip"
)

opts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    ScheduleToCloseTimeout: 15 * time.Minute,
    RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 1},
})

var re archivezip.Result
err := workflow.ExecuteActivity(
    opts,
    archivezip.Name,
    &archivezip.Params{
        SourceDir: "/path/to/example",
        DestPath:  "/path/to/example.zip",
    },
).Get(opts, &re)
```

`err` may contain any system error. `re.Path` will be the final path to the
created Zip archive.
