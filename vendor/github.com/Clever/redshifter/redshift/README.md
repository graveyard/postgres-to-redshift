# redshift
--
    import "github.com/Clever/redshifter/redshift"


## Usage

#### type ColInfo

```go
type ColInfo struct {
	Ordinal     int    `yaml:"ordinal"`
	Name        string `yaml:"dest"`
	Type        string `yaml:"type"`
	DefaultVal  string `yaml:"defaultval"`
	NotNull     bool   `yaml:"notnull"`
	PrimaryKey  bool   `yaml:"primarykey"`
	DistKey     bool   `yaml:"distkey"`
	SortOrdinal int    `yaml:"sortord"`
}
```

ColInfo is a struct that contains information about a column in a Redshift
database. SortOrdinal and DistKey only make sense for Redshift

#### type Meta

```go
type Meta struct {
	DataDateColumn string `yaml:"datadatecolumn"`
	Schema         string `yaml:"schema"`
}
```

Meta holds information that might be not in Redshift or annoying to access in
this case, we want to know the schema a table is part of and the column which
corresponds to the timestamp at which the data was gathered NOTE: this will be
useful for the s3-to-redshift worker, but is currently not very useful same with
the yaml info

#### type Redshift

```go
type Redshift struct {
}
```

Redshift wraps a dbExecCloser and can be used to perform operations on a
redshift database.

#### func  NewRedshift

```go
func NewRedshift(host, port, db, user, password string, timeout int, s3Info S3Info) (*Redshift, error)
```
NewRedshift returns a pointer to a new redshift object using configuration
values passed in on instantiation and the AWS env vars we assume exist Don't
need to pass s3 info unless doing a COPY operation

#### func (*Redshift) RefreshTables

```go
func (r *Redshift) RefreshTables(
	tables map[string]Table, schema, s3prefix string, delim rune) error
```
RefreshTables refreshes multiple tables in parallel and returns an error if any
of the copies fail.

#### func (*Redshift) RunCSVCopy

```go
func (r *Redshift) RunCSVCopy(tx *sql.Tx, schema, table, file string, ts Table, delimiter rune, creds, gzip bool) error
```
RunCSVCopy copies gzipped CSV data from an S3 file into a redshift table this is
meant to be run in a transaction, so the first arg must be a sql.Tx

#### func (*Redshift) RunJSONCopy

```go
func (r *Redshift) RunJSONCopy(tx *sql.Tx, schema, table, filename, jsonPaths string, creds, gzip bool) error
```
RunJSONCopy copies JSON data present in an S3 file into a redshift table. this
is meant to be run in a transaction, so the first arg must be a sql.Tx if not
using jsonPaths, set to "auto"

#### func (*Redshift) RunTruncate

```go
func (r *Redshift) RunTruncate(tx *sql.Tx, schema, table string) error
```
RunTruncate deletes all items from a table, given a transaction, a schema string
and a table name you shuold run vacuum and analyze soon after doing this for
performance reasons

#### func (*Redshift) VacuumAnalyze

```go
func (r *Redshift) VacuumAnalyze() error
```
VacuumAnalyze performs VACUUM FULL; ANALYZE on the redshift database. This is
useful for recreating the indices after a database has been modified and
updating the query planner.

#### func (*Redshift) VacuumAnalyzeTable

```go
func (r *Redshift) VacuumAnalyzeTable(schema, table string) error
```
VacuumAnalyzeTable performs VACUUM FULL; ANALYZE on a specific table. This is
useful for recreating the indices after a database has been modified and
updating the query planner.

#### type S3Info

```go
type S3Info struct {
	Region    string
	AccessID  string
	SecretKey string
}
```

S3Info holds the information necessary to copy data from s3 buckets

#### type Table

```go
type Table struct {
	Name    string    `yaml:"dest"`
	Columns []ColInfo `yaml:"columns"`
	Meta    Meta      `yaml:"meta"`
}
```

Table is our representation of a Redshift table the main difference is an added
metadata section and YAML unmarshalling guidance
