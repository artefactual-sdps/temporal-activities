# xmlvalidate

Validates an XML file against an XSD schema using `xmlint`.

## Requirements

This activity requires the `xmllint` command be installed on the system where
the activity worker is running. This can be installed, in Ubuntu, by entering
the following command:

```bash
apt-get install libxml2-utils
```

## Registration

The `Name` constant is used as example, use a different name to register and
execute the activity if that doesn't suit your needs. An example registration:

```go
import (
    "go.temporal.io/sdk/activity"
    "go.temporal.io/sdk/worker"

    "github.com/artefactual-sdps/temporal-activities/xmlvalidate"
)

tw := worker.New(...)

tw.RegisterActivityWithOptions(
    xmlvalidate.New().Execute,
    activity.RegisterOptions{Name: xmlvalidate.Name},
)
```

## Execution

An example execution:

```go
import (
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "github.com/artefactual-sdps/temporal-activities/xmlvalidate"
)

ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    ScheduleToCloseTimeout: 5 * time.Minute,
    RetryPolicy: &temporal.RetryPolicy{MaximumAttempts: 1},
})

var re xmlvalidate.Result
err := workflow.ExecuteActivity(
    ctx,
    xmlvalidate.Name,
    &xmlvalidate.Params{
        XMLFilePath: "/path/to/file.xml",
        XSDFilePath: "/path/to/file.xsd",
    },
).Get(ctx, &re)
```

`err` may contain any non validation error. `re.Failures` contains the
`xmllint` validation output as `[]bytes`.
