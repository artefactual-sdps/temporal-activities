# bucketupload

Uploads a local file to a configured [gocloud.dev/blob] bucket. Allows setting
the object key with the `Key` parameter, using the filename if not set. The
`BufferSize` parameter changes the default size in bytes of the chunks that
will be uploaded in a single request. If 0, the driver will choose a reasonable
default and it could be ignored by some drivers. For example, MinIO supports a
maximum of 10,000 chuncks per upload, which may require to increase the buffer
size to upload big files.

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

	"github.com/artefactual-sdps/temporal-activities/bucketupload"
)

tw := worker.New(...)

b, err := bucket.NewWithConfig(ctx, &bucket.Config{URL: "<driver-url>"})
if err != nil {
    // Handle error.
}
defer b.Close()

tw.RegisterActivityWithOptions(
    bucketupload.New(b).Execute,
    activity.RegisterOptions{Name: bucketupload.Name},
)
```

## Execution

An example execution:

```go
import (
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "github.com/artefactual-sdps/temporal-activities/bucketupload"
)

opts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    StartToCloseTimeout: time.Hour * 2,
    HeartbeatTimeout:    time.Second * 5,
    RetryPolicy: &temporal.RetryPolicy{
        MaximumAttempts: 3,
        NonRetryableErrorTypes: []string{
            "TemporalTimeout:StartToClose",
        },
    },
})

var re bucketupload.Result
err := workflow.ExecuteActivity(
    opts,
    bucketupload.Name,
    &bucketupload.Params{
        Path:       "/path/to/file.zip",
        Key:        "file.zip",
        BufferSize: 100_000_000,
    },
).Get(opts, &re)
```

`err` may contain any system error. `re.Key` contains the object key used in
the upload.

[gocloud.dev/blob]: https://pkg.go.dev/gocloud.dev/blob
[Go CDK guide]: https://gocloud.dev/howto/blob
[go.artefactual.dev/tools/bucket]: https://pkg.go.dev/go.artefactual.dev/tools/bucket
