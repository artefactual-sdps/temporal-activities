# bagcreate

Creates a BagIt Bag from a given directory path. Allows setting the path where
the Bag should be created. If not set, the Bag will be created at the source
path, replacing the original directory contents. The checksum algorithm used to
generate file checksums can be configured, valid values are "md5", "sha1",
"sha256" and "sha512" (default).

## Registration

The `Name` constant is used as example, use any name to register and execute
the activity that meets your needs. An example registration:

```go
import (
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"

	"github.com/artefactual-sdps/temporal-activities/bagcreate"
)

tw := worker.New(...)

tw.RegisterActivityWithOptions(
    bagcreate.New(bagcreate.Config{ChecksumAlgorithm: "md5"}).Execute,
    activity.RegisterOptions{Name: bagcreate.Name},
)
```

## Execution

An example execution:

```go
import (
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "github.com/artefactual-sdps/temporal-activities/bagcreate"
)

opts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    ScheduleToCloseTimeout: 15 * time.Minute,
    RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 1},
})

var re bagcreate.Result
err := workflow.ExecuteActivity(
    opts,
    bagcreate.Name,
    &bagcreate.Params{
        SourcePath: "/path/to/dir",
        BagPath:    "/path/to/bag",
    },
).Get(opts, &re)
```

`err` may contain any system error. `re.BagPath` will be the final path to the
created Bag.
