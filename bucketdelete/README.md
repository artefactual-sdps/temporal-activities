# bucketdelete

Deletes a file/blob from a configured [gocloud.dev/blob] bucket.

This activity will heartbeat each one-third of the configured timeout, if set
in the activity options.

## Registration

The `Name` constant is used as example, use a different name to register and
execute the activity if that doesn't suit your needs. See this [Go CDK guide]
and [go.artefactual.dev/tools/bucket] for options to open a bucket. An example
registration:

```go
import (
	"go.artefactual.dev/tools/bucket"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"

	"github.com/artefactual-sdps/temporal-activities/bucketdelete"
)

tw := worker.New(...)

b, err := bucket.NewWithConfig(ctx, &bucket.Config{URL: "<driver-url>"})
if err != nil {
    // Handle error.
}
defer b.Close()

tw.RegisterActivityWithOptions(
    bucketdelete.New(b).Execute,
    activity.RegisterOptions{Name: bucketdelete.Name},
)
```

## Execution

An example execution:

```go
import (
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "github.com/artefactual-sdps/temporal-activities/bucketdelete"
)

opts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    StartToCloseTimeout: time.Minute * 10,
    HeartbeatTimeout:    time.Second * 5,
    RetryPolicy: &temporal.RetryPolicy{
        MaximumAttempts: 3,
    },
})

var re bucketdelete.Result
err := workflow.ExecuteActivity(
    opts,
    bucketdelete.Name,
    &bucketdelete.Params{Key: "file.zip"},
).Get(opts, &re)
```

`err` may contain any system error. `re` will be empty.

[gocloud.dev/blob]: https://pkg.go.dev/gocloud.dev/blob
[Go CDK guide]: https://gocloud.dev/howto/blob
[go.artefactual.dev/tools/bucket]: https://pkg.go.dev/go.artefactual.dev/tools/bucket
