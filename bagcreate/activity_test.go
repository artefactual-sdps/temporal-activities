package bagcreate_test

import (
	"io/fs"
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/bagcreate"
)

const (
	dirMode  fs.FileMode = 0o700
	fileMode fs.FileMode = 0o600

	sha256manifest string = `5896fb5c3f2944f57c993fa06c130ff2c4182e4fea61c2597c52b0f9d437040e  data/another.txt
4450c8a88130a3b397bfc659245c4f0f87a8c79d017a60bdb1bd32f4b51c8133  data/small.txt
`
	sha512manifest string = `946af3bfd3b0b84ea0d99136085dcd66ee7e769371dbcd097ed35fd377116087e25d004afd68dc48e4eb0bcb6a434b04078577b531a7da1452296d1ae98d20b3  data/another.txt
8cbdd4ed5452f7c066509c066d5ea87fc03f30b0c67153624a1bce4d6e14b6709b5e78caf723cdf419d0efad4db96ba1cad3196783c26a7743029459bdd148b0  data/small.txt
`
)

func testBagManifest(t *testing.T) tfs.Manifest {
	return tfs.Expected(t,
		tfs.WithFile("bag-info.txt", "", tfs.MatchAnyFileContent, tfs.WithMode(fileMode)),
		tfs.WithFile("bagit.txt", "", tfs.MatchAnyFileContent, tfs.WithMode(fileMode)),
		tfs.WithFile("manifest-sha512.txt", sha512manifest, tfs.WithMode(fileMode)),
		tfs.WithFile("tagmanifest-sha512.txt", "", tfs.MatchAnyFileContent, tfs.WithMode(fileMode)),
		tfs.WithDir("data", tfs.WithMode(dirMode),
			tfs.WithFile("small.txt", "I am a small file.\n", tfs.WithMode(fileMode)),
			tfs.WithFile("another.txt", "I am another file.\n", tfs.WithMode(fileMode)),
		),
	)
}

func sourcePath(t *testing.T) string {
	t.Helper()

	td := tfs.NewDir(t, "sdps_bagit_create_test",
		tfs.WithFile("small.txt", "I am a small file.\n"),
		tfs.WithFile("another.txt", "I am another file.\n"),
	)

	return td.Path()
}

func TestActivity(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		cfg     bagcreate.Config
		params  bagcreate.Params
		want    tfs.Manifest
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Creates a bag in place",
			params: bagcreate.Params{
				SourcePath: sourcePath(t),
			},
			want: testBagManifest(t),
		},
		{
			name: "Creates a bag in a new dir",
			params: bagcreate.Params{
				SourcePath: sourcePath(t),
				BagPath:    tfs.NewDir(t, "sdps_bagit_create_test").Path(),
			},
			want: testBagManifest(t),
		},
		{
			name: "Creates a bag with SHA-256 checksums",
			cfg:  bagcreate.Config{ChecksumAlgorithm: "sha256"},
			params: bagcreate.Params{
				SourcePath: sourcePath(t),
				BagPath:    tfs.NewDir(t, "sdps_bagit_create_test").Path(),
			},
			want: tfs.Expected(t,
				tfs.WithFile("bag-info.txt", "", tfs.MatchAnyFileContent, tfs.WithMode(fileMode)),
				tfs.WithFile("bagit.txt", "", tfs.MatchAnyFileContent, tfs.WithMode(fileMode)),
				tfs.WithFile("manifest-sha256.txt", sha256manifest, tfs.WithMode(fileMode)),
				tfs.WithFile("tagmanifest-sha256.txt", "", tfs.MatchAnyFileContent, tfs.WithMode(fileMode)),
				tfs.WithDir("data", tfs.WithMode(dirMode),
					tfs.WithFile("small.txt", "I am a small file.\n", tfs.WithMode(fileMode)),
					tfs.WithFile("another.txt", "I am another file.\n", tfs.WithMode(fileMode)),
				),
			),
		},
		{
			name: "Errors if source dir is empty",
			params: bagcreate.Params{
				SourcePath: tfs.NewDir(t, "sdps_bagit_create_test").Path(),
			},
			wantErr: "activity error (type: bag-create, scheduledEventID: 0, startedEventID: 0, identity: ): bagcreate: create bag: could not create a bag, no files present in",
		},
		{
			name: "Errors if bag path isn't writable",
			params: bagcreate.Params{
				SourcePath: sourcePath(t),
				BagPath:    tfs.NewDir(t, "sdps_bagit_create_test", tfs.WithMode(0o600)).Path(),
			},
			wantErr: "activity error (type: bag-create, scheduledEventID: 0, startedEventID: 0, identity: ): bagcreate: copy source dir to bag path: open",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				bagcreate.New(tt.cfg).Execute,
				temporalsdk_activity.RegisterOptions{Name: bagcreate.Name},
			)

			enc, err := env.ExecuteActivity(bagcreate.Name, tt.params)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			var result bagcreate.Result
			_ = enc.Get(&result)

			// The bag metadata (bag-info.txt, bagit.txt) are non-deterministic,
			// so just assert that the expected metadata files are present, the
			// manifest is correct, and data directory contains the right files.
			assert.Assert(t, tfs.Equal(result.BagPath, tt.want))
		})
	}
}
