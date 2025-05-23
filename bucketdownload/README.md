# bucketdownload

Downloads a file/blob from a configured [gocloud.dev/blob] bucket. Allows
setting the directory, filename and permissions of the downloaded file. It will
create any missing directory if needed when the the dir path is set, otherwise
it will create one in the default directory for temporary files. If the
permission parameters are not set, it will use `0o700` for directories and
`0o600` for the file. If the filename is not provided, it will use the object
key.

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

	"github.com/artefactual-sdps/temporal-activities/bucketdownload"
)

tw := worker.New(...)

b, err := bucket.NewWithConfig(ctx, &bucket.Config{URL: "<driver-url>"})
if err != nil {
    // Handle error.
}
defer b.Close()

tw.RegisterActivityWithOptions(
    bucketdownload.New(b).Execute,
    activity.RegisterOptions{Name: bucketdownload.Name},
)
```

## Execution

An example execution:

```go
import (
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "github.com/artefactual-sdps/temporal-activities/bucketdownload"
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

var re bucketdownload.Result
err := workflow.ExecuteActivity(
    opts,
    bucketdownload.Name,
    &bucketdownload.Params{
        DirPath:  "/path/to/dir",
        DirPerm:  0o755,
        FileName: "file.zip",
        FilePerm: 0o644,
        Key:      "changed.zip",
    },
).Get(opts, &re)
```

`err` may contain any system error. `re.FilePath` contains the full path to the
downloaded file.

[gocloud.dev/blob]: https://pkg.go.dev/gocloud.dev/blob
[Go CDK guide]: https://gocloud.dev/howto/blob
[go.artefactual.dev/tools/bucket]: https://pkg.go.dev/go.artefactual.dev/tools/bucket
