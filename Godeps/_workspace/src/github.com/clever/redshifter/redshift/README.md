# redshift
--
    import "github.com/Clever/redshifter/redshift"


## Usage

#### type Redshift

```go
type Redshift struct {
}
```

Redshift wraps a dbExecCloser and can be used to perform operations on a
redshift database.

#### func  NewRedshift

```go
func NewRedshift() (*Redshift, error)
```
NewRedshift returns a pointer to a new redshift object using configuration
values set in the flags.

#### func (*Redshift) CopyGzipCsvDataFromS3

```go
func (r *Redshift) CopyGzipCsvDataFromS3(table, file, awsRegion string, delimiter rune) error
```
CopyGzipCsvDataFromS3 copies gzipped CSV data from an S3 file into a redshift
table.

#### func (*Redshift) CopyJSONDataFromS3

```go
func (r *Redshift) CopyJSONDataFromS3(table, file, jsonpathsFile, awsRegion string) error
```
CopyJSONDataFromS3 copies JSON data present in an S3 file into a redshift table.

#### func (*Redshift) RefreshTable

```go
func (r *Redshift) RefreshTable(name, prefix, file, awsRegion string, ts postgres.TableSchema, delim rune) error
```
RefreshTable refreshes a single table by first copying gzipped CSV data into a
temporary table and later renaming the temporary table to the original one.

#### func (*Redshift) RefreshTables

```go
func (r *Redshift) RefreshTables(
	tables map[string]postgres.TableSchema, s3prefix, awsRegion string, delim rune) error
```
RefreshTables refreshes multiple tables in parallel and returns an error if any
of the copies fail.

#### func (*Redshift) VacuumAnalyze

```go
func (r *Redshift) VacuumAnalyze() error
```
VacuumAnalyze performs VACUUM FULL; ANALYZE on the redshift database. This is
useful for recreating the indices after a database has been modified and
updating the query planner.
