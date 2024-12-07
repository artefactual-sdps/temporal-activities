package bagvalidate_test

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/bagvalidate"
)

const (
	textFileTxtCorrect = `This is a Test file
`

	bagInfoTxt = `Bag-Software-Agent: bagit.py v1.8.1 <https://github.com/LibraryOfCongress/bagit-python>
Bagging-Date: 2021-10-11
Payload-Oxum: 20.1
`

	bagitTxt = `BagIt-Version: 0.97
Tag-File-Character-Encoding: UTF-8
`

	manifestSha256Txt = `20cd2eb771177035f483363951203be7cd85f176aaa7d124a56eb4c83562a861  data/test-file.txt`

	tagManifestSha256Text = `e91f941be5973ff71f1dccbdd1a32d598881893a7f21be516aca743da38b1689 bagit.txt
c4600f10b98eb9f179781387e7ce80ff89b4a29793be74ccd037b44b0bf27c00 bag-info.txt
4698e56fb06c495df8f928fd3158d274ca070cc066a770ecb5cc364a9ff12edc manifest-sha256.txt`
)

func TestValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		bagPath string
		wantErr string
	}{
		{
			name: "Validate non-bag",
			bagPath: fs.NewDir(t, "",
				fs.WithDir("data",
					fs.WithFile("test-file.txt", textFileTxtCorrect),
				),
			).Path(),
			wantErr: "not a bag: bagit.txt not found",
		},
		{
			name: "Validate valid bag",
			bagPath: fs.NewDir(t, "",
				fs.WithDir("data",
					fs.WithFile("test-file.txt", textFileTxtCorrect),
				),
				fs.WithFile("bag-info.txt", bagInfoTxt),
				fs.WithFile("bagit.txt", bagitTxt),
				fs.WithFile("manifest-sha256.txt", manifestSha256Txt),
				fs.WithFile("tagmanifest-sha256.txt", tagManifestSha256Text),
			).Path(),
		},
		{
			name: "Validate invalid bag",
			bagPath: fs.NewDir(t, "",
				fs.WithDir("data",
					fs.WithFile("test-file.txt", "This is wrong"),
				),
				fs.WithFile("bag-info.txt", bagInfoTxt),
				fs.WithFile("bagit.txt", bagitTxt),
				fs.WithFile("manifest-sha256.txt", manifestSha256Txt),
				fs.WithFile("tagmanifest-sha256.txt", tagManifestSha256Text),
			).Path(),
			wantErr: "payload-oxum validation failed. expected 1 files and 20 bytes but found 1 files and 13 bytes",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := bagvalidate.NewValidator()
			err := v.Validate(tt.bagPath)

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("error is nil, expecting: %q", tt.wantErr)
				}

				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func TestNoopValidator(t *testing.T) {
	v := bagvalidate.NewNoopValidator()
	assert.NilError(t, v.Validate(""))
}
