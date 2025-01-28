# removefiles

Deletes any file or directory (and children) within a given path whose name
matches one of the values passed as names or patterns, and returns a count of
deleted items.

## Registration

The `Name` constant is used as example, use a different name to register and
execute the activity if that doesn't suit your needs. An example registration:

```go
import (
    "go.temporal.io/sdk/activity"
    "go.temporal.io/sdk/worker"

	"github.com/artefactual-sdps/temporal-activities/removefiles"
)

tw := worker.New(...)

tw.RegisterActivityWithOptions(
    removefiles.New().Execute,
    activity.RegisterOptions{Name: removefiles.Name},
)
```

## Execution

An example execution:

```go
import (
	"regexp"
	"time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "github.com/artefactual-sdps/temporal-activities/removefiles"
)

opts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    ScheduleToCloseTimeout: 5 * time.Minute,
    RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 1},
})

var re removefiles.Result
err := workflow.ExecuteActivity(
    opts,
    removefiles.Name,
    &removefiles.Params{
        Path: "/path/to/dir",
        RemoveNames: []string{
            "Thumbs.db",
            ".DS_Store",
        },
        RemovePatterns: []*regexp.Regexp{
            regexp.MustCompile("premis.xml$"),
            regexp.MustCompile("(?i)mets.xml$"),
        },
    },
).Get(opts, &re)
```

`err` may contain any system error. `re.Count` contains the number of deleted
directories/files as `int`. A deleted directory is counted as one deleted item
no matter how many items the directory contains.
