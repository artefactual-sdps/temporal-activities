# bucketcopy

Copies a blob within a configured [gocloud.dev/blob] bucket. The activity
accepts source and destination keys and performs an in-bucket copy operation.

This activity will heartbeat each one-third of the configured timeout, if set
in the activity options.

## Registration

The `Name` constant is used as example, use any name to register and execute
the activity that meets your needs. See the [Go CDK guide] and
[go.artefactual.dev/tools/bucket] for options to open a bucket. An example
registration:

```go
import (
    "go.artefactual.dev/tools/bucket"
    "go.temporal.io/sdk/activity"
    "go.temporal.io/sdk/worker"

    "github.com/artefactual-sdps/temporal-activities/bucketcopy"
)

tw := worker.New(...)

b, err := bucket.NewWithConfig(ctx, &bucket.Config{URL: "<driver-url>"})
if err != nil {
    // Handle error.
}
defer b.Close()

tw.RegisterActivityWithOptions(
    bucketcopy.New(b).Execute,
    activity.RegisterOptions{Name: bucketcopy.Name},
)
```

## Execution

An example execution:

```go
import (
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "github.com/artefactual-sdps/temporal-activities/bucketcopy"
)

opts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    StartToCloseTimeout: time.Hour,
    HeartbeatTimeout:    time.Minute,
    RetryPolicy: &temporal.RetryPolicy{
        MaximumAttempts: 3,
        NonRetryableErrorTypes: []string{
            "TemporalTimeout:StartToClose",
        },
    },
})

var re bucketcopy.Result
err := workflow.ExecuteActivity(
    opts,
    bucketcopy.Name,
    &bucketcopy.Params{
        SourceKey: "source.txt",
        DestKey:   "dest.txt",
    },
).Get(opts, &re)
```

`err` may contain any system error. `re` will be empty.

[gocloud.dev/blob]: https://pkg.go.dev/gocloud.dev/blob
[Go CDK guide]: https://gocloud.dev/howto/blob
[go.artefactual.dev/tools/bucket]: https://pkg.go.dev/go.artefactual.dev/tools/bucket
