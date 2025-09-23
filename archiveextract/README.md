# archiveextract

Extracts the contents of an given archive to a directory. It supports the
formats recognized by [github.com/mholt/archives] and allows configuring the
path and permissions of the extracted directories and files.

If the permissions configuration is not provided, it will use `0o700` for
directories and `0o600` for files. A destination path for the extracted folder
can be set as part of the activity parameters. If a destination path is not
provided the contents of the archive will be extracted to a subdirectory in the
same directory as the original archive file.

## Registration

The `Name` constant is used as example, use any name to register and execute
the activity that meets your needs. An example registration:

```go
import (
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"

	"github.com/artefactual-sdps/temporal-activities/archiveextract"
)

tw := worker.New(...)

cfg := archiveextract.Config{
    DirMode: 0o755,
    FileMode: 0o644,
}

tw.RegisterActivityWithOptions(
    archiveextract.New(cfg).Execute,
    activity.RegisterOptions{Name: archiveextract.Name},
)
```

## Execution

An example execution:

```go
import (
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "github.com/artefactual-sdps/temporal-activities/archiveextract"
)

opts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    ScheduleToCloseTimeout: 15 * time.Minute,
    RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 1},
})

var re archiveextract.Result
err := workflow.ExecuteActivity(
    opts,
    archiveextract.Name,
    &archiveextract.Params{
        SourcePath: "/path/to/example.zip",
		DestPath:   "/path/to/destination",
    },
).Get(opts, &re)
```

`err` may contain any system error. `re.ExtractPath` will be the final path to
the extracted archive contents.

[github.com/mholt/archives]: https://pkg.go.dev/github.com/mholt/archives
