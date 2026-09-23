package jsonvalidate_test

import (
	"path/filepath"
	"sort"
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/temporal-activities/jsonvalidate"
)

func TestActivity(t *testing.T) {
	t.Parallel()

	schemaPath := filepath.Join("testdata", "person.schema.json")
	validPath := filepath.Join("testdata", "person_valid.json")

	type test struct {
		name    string
		params  jsonvalidate.Params
		want    jsonvalidate.Result
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Valid document",
			params: jsonvalidate.Params{
				JSONPath:   validPath,
				SchemaPath: schemaPath,
			},
			want: jsonvalidate.Result{},
		},
		{
			name: "Invalid boolean",
			params: jsonvalidate.Params{
				JSONPath:   filepath.Join("testdata", "person_invalid_active.json"),
				SchemaPath: schemaPath,
			},
			want: jsonvalidate.Result{Failures: []string{"/active: got string, want boolean"}},
		},
		{
			name: "Invalid float",
			params: jsonvalidate.Params{
				JSONPath:   filepath.Join("testdata", "person_invalid_height.json"),
				SchemaPath: schemaPath,
			},
			want: jsonvalidate.Result{Failures: []string{"/height: got string, want number"}},
		},
		{
			name: "Missing field",
			params: jsonvalidate.Params{
				JSONPath:   filepath.Join("testdata", "person_missing_email.json"),
				SchemaPath: schemaPath,
			},
			want: jsonvalidate.Result{Failures: []string{"missing property 'email'"}},
		},
		{
			name: "Multiple validation errors",
			params: jsonvalidate.Params{
				JSONPath:   filepath.Join("testdata", "person_invalid_many.json"),
				SchemaPath: schemaPath,
			},
			want: jsonvalidate.Result{Failures: []string{
				"missing property 'email'",
				"/name: got number, want string",
				"/age: minimum: got -1, want 0",
				"/active: got string, want boolean",
				"/height: got string, want number",
				"additional properties 'extra' not allowed",
			}},
		},
		{
			name:    "Empty JSON path",
			params:  jsonvalidate.Params{SchemaPath: schemaPath},
			wantErr: "JSON path cannot be empty",
		},
		{
			name:    "Empty JSON schema path",
			params:  jsonvalidate.Params{JSONPath: validPath},
			wantErr: "JSON schema path cannot be empty",
		},
		{
			name: "Missing JSON document",
			params: jsonvalidate.Params{
				JSONPath:   filepath.Join("testdata", "missing.json"),
				SchemaPath: schemaPath,
			},
			wantErr: "missing.json is missing",
		},
		{
			name: "Missing JSON schema",
			params: jsonvalidate.Params{
				JSONPath:   validPath,
				SchemaPath: filepath.Join("testdata", "missing.schema.json"),
			},
			wantErr: "missing.schema.json is missing",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				jsonvalidate.New(jsonvalidate.NewValidator()).Execute,
				temporalsdk_activity.RegisterOptions{Name: jsonvalidate.Name},
			)

			future, err := env.ExecuteActivity(jsonvalidate.Name, tt.params)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			var res jsonvalidate.Result
			assert.NilError(t, future.Get(&res))
			sort.Strings(res.Failures)
			sort.Strings(tt.want.Failures)
			assert.DeepEqual(t, res, tt.want)
		})
	}
}
