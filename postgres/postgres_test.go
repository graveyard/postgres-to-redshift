package postgres

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/pg.v2"
)

type mockPgDB struct {
	copycmds  *[]string
	querycmds *[]string
}

func (m mockPgDB) Query(f pg.Factory, q string, args ...interface{}) (*pg.Result, error) {
	*m.querycmds = append(*m.querycmds, q)
	row1 := f.New()
	if ci, ok := row1.(*ColInfo); ok {
		*ci = ColInfo{Ordinal: 1, Name: "colname1", ColType: "coltype1", DefaultVal: "defaultval1", NotNull: true, PrimaryKey: true}
	}
	row2 := f.New()
	if ci, ok := row2.(*ColInfo); ok {
		*ci = ColInfo{Ordinal: 2, Name: "colname2", ColType: "coltype2", DefaultVal: "", NotNull: false, PrimaryKey: false}
	}
	return nil, nil
}

func (m mockPgDB) CopyTo(wc io.WriteCloser, q string, args ...interface{}) (*pg.Result, error) {
	*m.copycmds = append(*m.copycmds, q)
	wc.Write([]byte("test copy output"))
	wc.Close()
	return nil, nil
}

func (m mockPgDB) Close() error {
	return nil
}

func TestDumpTableToS3(t *testing.T) {
	file, out := new(string), new(string)
	mockS3Writer := func(fname string, r io.Reader, len int64) error {
		*file = fname
		gzipr, err := gzip.NewReader(r)
		if err != nil {
			t.Fatalf("Error creating gzip reader. %s", err.Error())
		}
		buf, err := ioutil.ReadAll(gzipr)
		*out = string(buf)
		return err
	}
	copycmds, querycmds := []string{}, []string{}
	mockp := &DB{mockPgDB{&copycmds, &querycmds}, mockS3Writer}
	assert.NoError(t, mockp.DumpTableToS3("tablename", "s3file"))
	expcmds := []string{"COPY tablename TO STDOUT WITH (FORMAT csv, DELIMITER '|', HEADER 0)"}
	assert.Equal(t, expcmds, copycmds)
	assert.Equal(t, "s3file", *file)
	assert.Equal(t, "test copy output", *out)
}

func TestGetTableSchema(t *testing.T) {
	copycmds, querycmds := []string{}, []string{}
	mockp := &DB{mockPgDB{&copycmds, &querycmds}, nil}
	ts, err := mockp.GetTableSchema("tablename", "namespace")
	assert.NoError(t, err)
	expcmds := []string{fmt.Sprintf(schemaQueryFormat, "namespace", "tablename")}
	assert.Equal(t, expcmds, querycmds)
	expts := TableSchema{
		&ColInfo{Ordinal: 1, Name: "colname1", ColType: "coltype1", DefaultVal: "defaultval1", NotNull: true, PrimaryKey: true},
		&ColInfo{Ordinal: 2, Name: "colname2", ColType: "coltype2", DefaultVal: "", NotNull: false, PrimaryKey: false},
	}
	assert.Equal(t, expts, ts)
}
