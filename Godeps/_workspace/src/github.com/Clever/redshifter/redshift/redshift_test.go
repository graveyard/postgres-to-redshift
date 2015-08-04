package redshift

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/Clever/redshifter/postgres"
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
	ts := postgres.TableSchema{
		{3, "field3", "type3", "defaultval3", false, false},
		{1, "field1", "type1", "", true, false},
		{2, "field2", "type2", "", false, true},
	}
	exp := fmt.Sprintf(`COPY "%s"."%s" (%s) FROM '%s' WITH REGION '%s' GZIP CSV DELIMITER '%c'`,
		schema, table, "field1, field2, field3", file, awsRegion, delimiter)
	exp += " IGNOREHEADER 0 ACCEPTINVCHARS TRUNCATECOLUMNS TRIMBLANKS BLANKSASNULL EMPTYASNULL DATEFORMAT 'auto' ACCEPTANYDATE COMPUPDATE ON"
	exp += " CREDENTIALS 'aws_access_key_id=accesskey;aws_secret_access_key=secretkey'"
	cmds := mockSQLDB([]string{})
	mockrs := Redshift{&cmds, "accesskey", "secretkey"}
	err := mockrs.CopyGzipCsvDataFromS3(schema, table, file, awsRegion, ts, delimiter)
	assert.NoError(t, err)
	assert.Equal(t, mockSQLDB{exp}, cmds)
}

func TestCreateTable(t *testing.T) {
	tmpschema, schema, name := "testtmpschema", "testschema", "testtable"
	exp := fmt.Sprintf(`CREATE TABLE "%s"."%s" (LIKE "%s"."%s")`, tmpschema, name, schema, name)
	cmds := mockSQLDB([]string{})
	mockrs := Redshift{&cmds, "accesskey", "secretkey"}
	err := mockrs.createTempTable(tmpschema, schema, name)
	assert.NoError(t, err)
	assert.Equal(t, mockSQLDB{exp}, cmds)
}

func TestRefreshTable(t *testing.T) {
	schema, name, tmpschema, file, awsRegion, delim := "testschema", "tablename", "testtmpschema", "s3://path", "testRegion", '|'
	ts := postgres.TableSchema{
		{3, "field3", "type3", "defaultval3", false, false},
		{1, "field1", "type1", "", true, false},
		{2, "field2", "type2", "", false, true},
	}
	copycmd := fmt.Sprintf(`COPY "%s"."%s" (%s) FROM '%s' WITH REGION '%s' GZIP CSV DELIMITER '%c'`,
		tmpschema, name, "field1, field2, field3", file, awsRegion, delim)
	copycmd += " IGNOREHEADER 0 ACCEPTINVCHARS TRUNCATECOLUMNS TRIMBLANKS BLANKSASNULL EMPTYASNULL DATEFORMAT 'auto' ACCEPTANYDATE COMPUPDATE ON"
	copycmd += " CREDENTIALS 'aws_access_key_id=accesskey;aws_secret_access_key=secretkey'"
	datarefreshcmds := []string{
		"BEGIN TRANSACTION",
		fmt.Sprintf(`DELETE FROM "%s"."%s"`, schema, name),
		fmt.Sprintf(`INSERT INTO "%s"."%s" (SELECT * FROM "%s"."%s")`, schema, name, tmpschema, name),
		"END TRANSACTION",
	}
	expcmds := mockSQLDB{
		fmt.Sprintf(`CREATE TABLE "%s"."%s" (LIKE "%s"."%s")`, tmpschema, name, schema, name),
		copycmd,
		strings.Join(datarefreshcmds, "; "),
	}
	cmds := mockSQLDB([]string{})
	mockrs := Redshift{&cmds, "accesskey", "secretkey"}
	err := mockrs.refreshTable(schema, name, tmpschema, file, awsRegion, ts, delim)
	assert.NoError(t, err)
	assert.Equal(t, expcmds, cmds)
}
