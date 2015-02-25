package redshift

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/clever/redshifter/postgres"
	"github.com/stretchr/testify/assert"
)

type mockSQLDB []string

func (m *mockSQLDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	*m = mockSQLDB(append([]string(*m), fmt.Sprintf(query)))
	return nil, nil
}

func (m *mockSQLDB) Close() error {
	return nil
}

func TestCopyJSONDataFromS3(t *testing.T) {
	schema, table, file, jsonpathsFile, awsRegion := "testschema", "tablename", "s3://path", "s3://jsonpathsfile", "testregion"
	exp := fmt.Sprintf("COPY \"%s\".\"%s\" FROM '%s' WITH json '%s' region '%s' timeformat 'epochsecs' COMPUPDATE ON",
		schema, table, file, jsonpathsFile, awsRegion)
	exp += " CREDENTIALS 'aws_access_key_id=accesskey;aws_secret_access_key=secretkey'"
	cmds := mockSQLDB([]string{})
	mockrs := Redshift{&cmds, "accesskey", "secretkey"}
	err := mockrs.CopyJSONDataFromS3(schema, table, file, jsonpathsFile, awsRegion)
	assert.NoError(t, err)
	assert.Equal(t, mockSQLDB{exp}, cmds)
}

func TestCopyGzipCsvDataFromS3(t *testing.T) {
	schema, table, file, awsRegion, delimiter := "testschema", "tablename", "s3://path", "testregion", '|'
	exp := fmt.Sprintf("COPY \"%s\".\"%s\" FROM '%s' WITH REGION '%s' GZIP CSV DELIMITER '%c' IGNOREHEADER 0",
		schema, table, file, awsRegion, delimiter)
	exp += " ACCEPTINVCHARS TRUNCATECOLUMNS TRIMBLANKS BLANKSASNULL EMPTYASNULL DATEFORMAT 'auto' ACCEPTANYDATE COMPUPDATE ON"
	exp += " CREDENTIALS 'aws_access_key_id=accesskey;aws_secret_access_key=secretkey'"
	cmds := mockSQLDB([]string{})
	mockrs := Redshift{&cmds, "accesskey", "secretkey"}
	err := mockrs.CopyGzipCsvDataFromS3(schema, table, file, awsRegion, delimiter)
	assert.NoError(t, err)
	assert.Equal(t, mockSQLDB{exp}, cmds)
}

func TestCreateTable(t *testing.T) {
	ts := postgres.TableSchema{
		{3, "field3", "type3", "defaultval3", false, false},
		{1, "field1", "type1", "", true, false},
		{2, "field2", "type2", "", false, true},
	}
	exp := "CREATE TABLE \"testschema\".\"tablename\" (field1 type1  NOT NULL, field2 type2 SORTKEY PRIMARY KEY, field3 type3 DEFAULT defaultval3 )"
	cmds := mockSQLDB([]string{})
	mockrs := Redshift{&cmds, "accesskey", "secretkey"}
	err := mockrs.createTable("testschema", "tablename", ts)
	assert.NoError(t, err)
	assert.Equal(t, mockSQLDB{exp}, cmds)
}

func TestRefreshTable(t *testing.T) {
	schema, name, prefix, file, awsRegion, delim := "testschema", "tablename", "test_prefix_", "s3://path", "testRegion", '|'
	ts := postgres.TableSchema{
		{3, "field3", "type3", "defaultval3", false, false},
		{1, "field1", "type1", "", true, false},
		{2, "field2", "type2", "", false, true},
	}
	copycmd := fmt.Sprintf("COPY \"%s\".\"%s\" FROM '%s' WITH REGION '%s' GZIP CSV DELIMITER '%c' IGNOREHEADER 0",
		schema, prefix+name, file, awsRegion, delim)
	copycmd += " ACCEPTINVCHARS TRUNCATECOLUMNS TRIMBLANKS BLANKSASNULL EMPTYASNULL DATEFORMAT 'auto' ACCEPTANYDATE COMPUPDATE ON"
	copycmd += " CREDENTIALS 'aws_access_key_id=accesskey;aws_secret_access_key=secretkey'"
	expcmds := mockSQLDB{
		"DROP TABLE IF EXISTS \"testschema\".\"test_prefix_tablename\"",
		"CREATE TABLE \"testschema\".\"test_prefix_tablename\" (field1 type1  NOT NULL, field2 type2 SORTKEY PRIMARY KEY, field3 type3 DEFAULT defaultval3 )",
		copycmd,
		"DROP TABLE IF EXISTS \"testschema\".\"tablename\"; ALTER TABLE \"testschema\".\"test_prefix_tablename\" RENAME TO \"testschema\".\"tablename\";",
	}
	cmds := mockSQLDB([]string{})
	mockrs := Redshift{&cmds, "accesskey", "secretkey"}
	err := mockrs.RefreshTable(schema, name, prefix, file, awsRegion, ts, delim)
	assert.NoError(t, err)
	assert.Equal(t, expcmds, cmds)
}
