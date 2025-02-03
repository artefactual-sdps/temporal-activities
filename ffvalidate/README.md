# ffvalidate

Identifies the file format of the files at the given path, recursively walking
any sub-directories, and validates that the formats are in the list of allowed
formats.

## Requirements

This activity requires reading an allowed file formats CSV file which path is
set on the activity configuration. This file can contain multiple columns, the
only requirement is to include a column with the `PRONOM PUID` heading (case
insensitive) and the allowed values. A simplified example:

```csv
Format name,PRONOM PUID
text,x-fmt/16
PDF/A,fmt/95
CSV,x-fmt/18
SIARD,fmt/161
TIFF,fmt/353
JPEG 2000,x-fmt/392
WAVE,fmt/141
FFV1,fmt/569
MPEG-4,fmt/199
XML/XSD,x-fmt/280
```

## Registration

The `Name` constant is used as example, use any name to register and execute
the activity that meets your needs. An example registration:

```go
import (
    "go.temporal.io/sdk/activity"
    "go.temporal.io/sdk/worker"

    "github.com/artefactual-sdps/temporal-activities/ffvalidate"
)

tw := worker.New(...)
cfg := ffvalidate.Config{AllowlistPath: "/path/to/allowed_file_formats.csv"}

tw.RegisterActivityWithOptions(
    ffvalidate.New(cfg).Execute,
    activity.RegisterOptions{Name: ffvalidate.Name},
)
```

## Execution

An example execution:

```go
import (
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "github.com/artefactual-sdps/temporal-activities/ffvalidate"
)

opts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    ScheduleToCloseTimeout: 5 * time.Minute,
    RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 1},
})

var re ffvalidate.Result
err := workflow.ExecuteActivity(
    opts,
    ffvalidate.Name,
    &ffvalidate.Params{Path: "/path/to/dir"},
).Get(opts, &re)
```

`err` may contain any non validation error. `re.Failures` contains a list of
files that are not an allowed format.
