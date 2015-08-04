# postgres
--
    import "github.com/Clever/redshifter/postgres"


## Usage

#### func  S3Filename

```go
func S3Filename(prefix string, table string) string
```
S3Filename returns the s3 filename used for storing the table data.

#### type ColInfo

```go
type ColInfo struct {
	Ordinal    int
	Name       string
	ColType    string
	DefaultVal string
	NotNull    bool
	PrimaryKey bool
}
```

ColInfo is a struct that contains information about a column in a postgreSQL
database.

#### type Config

```go
type Config struct {
	PoolSize int
}
```

Config is a struct used to specify configuration for the postgreSQL connection.

#### type DB

```go
type DB struct {
}
```

DB is a struct that is used to perform operations on a postgreSQL database.

#### func  NewDB

```go
func NewDB(cfg Config) *DB
```
NewDB returns a DB struct initialized based on flags.

#### func (*DB) DumpTableToS3

```go
func (db *DB) DumpTableToS3(table string, s3file string) error
```
DumpTableToS3 dumps a single table to S3 by executing a COPY TO query and
writing the gzipped CSV data to an S3 file.

#### func (*DB) DumpTablesToS3

```go
func (db *DB) DumpTablesToS3(tables []string, s3prefix string) error
```
DumpTablesToS3 dumps multiple tables to s3 in parallel.

#### func (*DB) GetTableSchema

```go
func (db *DB) GetTableSchema(table, namespace string) (TableSchema, error)
```
GetTableSchema returns the schema for a postgresSQL table by performing a query
on the postgreSQL internal tables. TODO: include foreign key relations

#### func (*DB) GetTableSchemas

```go
func (db *DB) GetTableSchemas(tables []string, namespace string) (map[string]TableSchema, error)
```
GetTableSchemas returns a map from a tablename to its schema. Gets schemas for
different tables in parallel.

#### type TableSchema

```go
type TableSchema []*ColInfo
```

TableSchema is a type which models the schema of a postgreSQL table.

#### func (TableSchema) Len

```go
func (ts TableSchema) Len() int
```

#### func (TableSchema) Less

```go
func (ts TableSchema) Less(i, j int) bool
```

#### func (*TableSchema) New

```go
func (ts *TableSchema) New() interface{}
```
New adds a pointer to a new ColInfo object to the TableSchema and returns it.

#### func (TableSchema) Swap

```go
func (ts TableSchema) Swap(i, j int)
```
