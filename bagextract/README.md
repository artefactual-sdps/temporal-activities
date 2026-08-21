# bagextract

Removes BagIt bag structure from a given directory by moving the contents of
the `data` directory to the bag root and deleting bag metadata files (e.g.
`bagit.txt`, manifests). If the path is not a bag (`bagit.txt` is missing),
the activity does nothing. Files or directories listed in `Keep` are preserved
at the bag root.

## Registration
The `Name` constant is used as example, use any name to register and execute 
the activity that meets your needs. An example registration:

```go
import (
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"

	"github.com/artefactual-sdps/temporal-activities/bagextract"
)

tw := worker.New(...)

tw.RegisterActivityWithOptions(
    bagextract.New().Execute,
    activity.RegisterOptions{Name: bagextract.Name},
)
```

## Execution

An example execution:

```go
import (
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "github.com/artefactual-sdps/temporal-activities/bagextract"
)

opts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    ScheduleToCloseTimeout: 5 * time.Minute,
    RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 1},
})

var re bagextract.Result
err := workflow.ExecuteActivity(
    opts,
    bagextract.Name,
    &bagextract.Params{
        Path: "/path/to/bag",
        Keep: []string{"metadata"},
    },
).Get(opts, &re)
```

`err` may contain any system error, including a missing `data` directory when
the path is a bag. `re.Path` is the same path passed in `params.Path`.
