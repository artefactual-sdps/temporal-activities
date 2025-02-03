# xmlvalidate

Validates an XML file against an XSD schema. It also provides a validator based
on `xmllint`.

## Requirements

The `XMLLintValidator` requires the `xmllint` command to be installed on the
system where the activity worker is running. This can be installed, for example
in Ubuntu, by entering the following command:

```bash
apt-get install libxml2-utils
```

## Registration

The `Name` constant is used as example, use any name to register and execute
the activity that meets your needs. An example registration using the`xmllint`
validator:

```go
import (
    "go.temporal.io/sdk/activity"
    "go.temporal.io/sdk/worker"

    "github.com/artefactual-sdps/temporal-activities/xmlvalidate"
)

tw := worker.New(...)

tw.RegisterActivityWithOptions(
    xmlvalidate.New(xmlvalidate.NewXMLLintValidator()).Execute,
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

opts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    ScheduleToCloseTimeout: 5 * time.Minute,
    RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 1},
})

var re xmlvalidate.Result
err := workflow.ExecuteActivity(
    opts,
    xmlvalidate.Name,
    &xmlvalidate.Params{
        XMLPath: "/path/to/file.xml",
        XSDPath: "/path/to/file.xsd",
    },
).Get(opts, &re)
```

`err` may contain any non validation error. `re.Failures` contains the
`xmllint` validation output as `[]string`.
