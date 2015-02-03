# pathio
--
    import "github.com/Clever/pathio"

Package pathio is a package that allows writing to and reading from different
types of paths transparently. It supports two types of paths:

    1. Local file paths
    2. S3 File Paths (s3://bucket/key)

Note that using s3 paths requires setting three environment variables

    1. AWS_SECRET_ACCESS_KEY
    2. AWS_ACCESS_KEY_ID
    3. AWS_REGION

## Usage

#### func  Reader

```go
func Reader(path string) (io.Reader, error)
```
Reader returns an io.Reader for the specified path. The path can either be a
local file path or an S3 path.

#### func  Write

```go
func Write(path string, input []byte) error
```
Write writes a byte array to the specified path. The path can be either a local
file path of an S3 path.

#### func  WriteReader

```go
func WriteReader(path string, input io.Reader, length int64) error
```
WriteReader writes all the data read from the specified io.Reader to the output
path. The path can either a local file path or an S3 path.
