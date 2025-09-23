[![Go Reference](https://pkg.go.dev/badge/github.com/artefactual-sdps/temporal-activities.svg)](https://pkg.go.dev/github.com/artefactual-sdps/temporal-activities)
[![Go Report](https://goreportcard.com/badge/github.com/artefactual-sdps/temporal-activities)](https://goreportcard.com/report/github.com/artefactual-sdps/temporal-activities)
[![Tests](https://github.com/artefactual-sdps/temporal-activities/actions/workflows/test.yml/badge.svg)](https://github.com/artefactual-sdps/temporal-activities/actions/workflows/test.yml)
[![Coverage](https://img.shields.io/codecov/c/github/artefactual-sdps/temporal-activities)](https://app.codecov.io/gh/artefactual-sdps/temporal-activities)

# Temporal Activities

The `temporal-activities` repository contains several predefined Temporal
activities implemented in Go that are ready to be integrated into your Temporal
workflows. Each activity is self-contained within its own package, with a
dedicated README file providing details on its functionality, usage, and
configuration.

## List of Temporal Activities

- [archiveextract](#archiveextract)
- [archivezip](#archivezip)
- [bagcreate](#bagcreate)
- [bagvalidate](#bagvalidate)
- [bucketcopy](#bucketcopy)
- [bucketdelete](#bucketdelete)
- [bucketdownload](#bucketdownload)
- [bucketupload](#bucketupload)
- [ffvalidate](#ffvalidate)
- [removefiles](#removefiles)
- [removepaths](#removepaths)
- [xmlvalidate](#xmlvalidate)

### archiveextract

Extracts the contents of an given archive to a directory. It supports the
formats recognized by [github.com/mholt/archives] and allows configuring the
path and permissions of the extracted directories and files.

[Read more](./archiveextract/README.md)

### archivezip

Creates a Zip archive from a given directory. Allows setting the destination
path, if not set then the source directory path + ".zip" will be used.

[Read more](./archivezip/README.md)

### bagcreate

Creates a BagIt Bag from a given directory path. Allows setting the path where
the Bag should be created and the checksum algorithm used to generate file
checksums.

[Read more](./bagcreate/README.md)

### bagvalidate

Checks if the given directory is a valid BagIt Bag.

[Read more](./bagvalidate/README.md)

### bucketcopy

Copies a blob within a configured [gocloud.dev/blob] bucket. The activity
accepts source and destination keys and performs an in-bucket copy operation.

### bucketdelete

Deletes a file/blob from a configured [gocloud.dev/blob] bucket.

[Read more](./bucketdelete/README.md)

### bucketdownload

Downloads a file/blob from a configured [gocloud.dev/blob] bucket. Allows
setting the directory, filename and permissions of the downloaded file.

[Read more](./bucketdownload/README.md)

### bucketupload

Uploads a local file to a configured [gocloud.dev/blob] bucket. Allows setting
the object key and the buffer size.

[Read more](./bucketupload/README.md)

### ffvalidate

Identifies the file format of the files at the given path, recursively walking
any sub-directories, and validates that the formats are in the list of allowed
formats.

[Read more](./ffvalidate/README.md)

### removefiles

Deletes any file or directory (and children) within a given path whose name
matches one of the values passed as names or patterns, and returns a count of
deleted items.

[Read more](./removefiles/README.md)

### removepaths

Removes all the given directory/file paths and any children they may contain
and returns any errors encountered.

[Read more](./removepaths/README.md)

### xmlvalidate

Validates an XML file against an XSD schema. It also provides a validator based
on `xmllint`.

[Read more](./xmlvalidate/README.md)

[github.com/mholt/archives]: https://pkg.go.dev/github.com/mholt/archives
[gocloud.dev/blob]: https://pkg.go.dev/gocloud.dev/blob
