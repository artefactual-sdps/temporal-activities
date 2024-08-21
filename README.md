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
- [bucketdelete](#bucketdelete)
- [bucketdownload](#bucketdownload)
- [bucketupload](#bucketupload)
- [removefiles](#removefiles)
- [removepaths](#removepaths)
- [xmlvalidate](#xmlvalidate)

### archiveextract

Extracts the contents of an archive to a given directory.

[Read more](./archiveextract/README.md)

### archivezip

Creates a zip archive from a directory.

[Read more](./archivezip/README.md)

### bagcreate

Creates a BagIt package from a given directory.

[Read more](./bagcreate/README.md)

### bagvalidate

Validates a BagIt package.

[Read more](./bagvalidate/README.md)

### bucketdelete

Deletes a file/blob from a configured bucket.

[Read more](./bucketdelete/README.md)

### bucketdownload

Downloads a file/blob from a configured bucket.

[Read more](./bucketdownload/README.md)

### bucketupload

Uploads a file/blob to a configured bucket.

[Read more](./bucketupload/README.md)

### removefiles

Removes files within a directory matching a set of names and/or patterns.

[Read more](./removefiles/README.md)

### removepaths

Removes specific paths from the filesystem.

[Read more](./removepaths/README.md)

### xmlvalidate

Validates an XML file against an XSD schema using `xmlint`.

[Read more](./xmlvalidate/README.md)
