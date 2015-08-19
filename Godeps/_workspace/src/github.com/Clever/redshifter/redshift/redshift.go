package redshift

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/Clever/redshifter/postgres"
	"github.com/facebookgo/errgroup"
	"github.com/segmentio/go-env"
)

type dbExecCloser interface {
	Close() error
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// Redshift wraps a dbExecCloser and can be used to perform operations on a redshift database.
type Redshift struct {
	dbExecCloser
	accessID  string
	secretKey string
}

var (
	// TODO: include flag validation
	awsAccessKeyID     = env.MustGet("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey = env.MustGet("AWS_SECRET_ACCESS_KEY")
	host               = flag.String("redshifthost", "", "Address of the redshift host")
	port               = flag.Int("redshiftport", 0, "Address of the redshift host")
	db                 = flag.String("redshiftdatabase", "", "Redshift database to connect to")
	user               = flag.String("redshiftuser", "", "Redshift user to connect as")
	pwd                = flag.String("redshiftpassword", "", "Password for the redshift user")
	timeout            = flag.Duration("redshiftconnecttimeout", 10*time.Second,
		"Timeout while connecting to Redshift. Defaults to 10 seconds.")
)

// NewRedshift returns a pointer to a new redshift object using configuration values set in the
// flags.
func NewRedshift() (*Redshift, error) {
	flag.Parse()
	source := fmt.Sprintf("host=%s port=%d dbname=%s connect_timeout=%d", *host, *port, *db, int(timeout.Seconds()))
	log.Println("Connecting to Reshift Source: ", source)
	source += fmt.Sprintf(" user=%s password=%s", *user, *pwd)
	sqldb, err := sql.Open("postgres", source)
	if err != nil {
		return nil, err
	}
	return &Redshift{sqldb, awsAccessKeyID, awsSecretAccessKey}, nil
}

func (r *Redshift) logAndExec(cmd string, creds bool) (sql.Result, error) {
	log.Print("Executing Redshift command: ", cmd)
	if creds {
		cmd += fmt.Sprintf(" CREDENTIALS '%s=%s;%s=%s'", "aws_access_key_id", r.accessID, "aws_secret_access_key", r.secretKey)
	}
	return r.Exec(cmd)
}

// CopyJSONDataFromS3 copies JSON data present in an S3 file into a redshift table.
func (r *Redshift) CopyJSONDataFromS3(schema, table, file, jsonpathsFile, awsRegion string) error {
	copyCmd := fmt.Sprintf(
		`COPY "%s"."%s" FROM '%s' WITH json '%s' region '%s' timeformat 'epochsecs' COMPUPDATE ON`,
		schema, table, file, jsonpathsFile, awsRegion)
	_, err := r.logAndExec(copyCmd, true)
	return err
}

// CopyGzipCsvDataFromS3 copies gzipped CSV data from an S3 file into a redshift table.
func (r *Redshift) CopyGzipCsvDataFromS3(schema, table, file, awsRegion string, ts postgres.TableSchema, delimiter rune) error {
	cols := []string{}
	sort.Sort(ts)
	for _, ci := range ts {
		cols = append(cols, ci.Name)
	}
	copyCmd := fmt.Sprintf(
		`COPY "%s"."%s" (%s) FROM '%s' WITH REGION '%s' GZIP CSV DELIMITER '%c'`,
		schema, table, strings.Join(cols, ", "), file, awsRegion, delimiter)
	copyCmd += " IGNOREHEADER 0 ACCEPTINVCHARS TRUNCATECOLUMNS TRIMBLANKS BLANKSASNULL EMPTYASNULL DATEFORMAT 'auto' ACCEPTANYDATE COMPUPDATE ON"
	_, err := r.logAndExec(copyCmd, true)
	return err
}

// Creates a table in tmpschema using the structure of the existing table in schema.
func (r *Redshift) createTempTable(tmpschema, schema, name string) error {
	cmd := fmt.Sprintf(`CREATE TABLE "%s"."%s" (LIKE "%s"."%s")`, tmpschema, name, schema, name)
	_, err := r.logAndExec(cmd, false)
	return err
}

func (r *Redshift) refreshData(tmpschema, schema, name string) error {
	cmds := []string{
		"BEGIN TRANSACTION",
		fmt.Sprintf(`DELETE FROM "%s"."%s"`, schema, name),
		fmt.Sprintf(`INSERT INTO "%s"."%s" (SELECT * FROM "%s"."%s")`, schema, name, tmpschema, name),
		"END TRANSACTION",
	}
	_, err := r.logAndExec(strings.Join(cmds, "; "), false)
	return err
}

// RefreshTable refreshes a single table by first copying gzipped CSV data into a temporary table
// and later replacing the original table's data with the one from the temporary table in an
// atomic operation.
func (r *Redshift) refreshTable(schema, name, tmpschema, file, awsRegion string, ts postgres.TableSchema, delim rune) error {
	if err := r.createTempTable(tmpschema, schema, name); err != nil {
		return err
	}
	if err := r.CopyGzipCsvDataFromS3(tmpschema, name, file, awsRegion, ts, delim); err != nil {
		return err
	}
	if err := r.refreshData(tmpschema, schema, name); err != nil {
		return err
	}
	return r.VacuumAnalyzeTable(schema, name)
}

// RefreshTables refreshes multiple tables in parallel and returns an error if any of the copies
// fail.
func (r *Redshift) RefreshTables(
	tables map[string]postgres.TableSchema, schema, tmpschema, s3prefix, awsRegion string, delim rune) error {
	if _, err := r.logAndExec(fmt.Sprintf(`CREATE SCHEMA "%s"`, tmpschema), false); err != nil {
		return err
	}
	group := new(errgroup.Group)
	for name, ts := range tables {
		group.Add(1)
		go func(name string, ts postgres.TableSchema) {
			if err := r.refreshTable(schema, name, tmpschema, postgres.S3Filename(s3prefix, name), awsRegion, ts, delim); err != nil {
				group.Error(err)
			}
			group.Done()
		}(name, ts)
	}
	errs := new(errgroup.Group)
	if err := group.Wait(); err != nil {
		errs.Error(err)
	}
	if _, err := r.logAndExec(fmt.Sprintf(`DROP SCHEMA "%s" CASCADE`, tmpschema), false); err != nil {
		errs.Error(err)
	}
	// Use errs.Wait() to group the two errors into a single error object.
	return errs.Wait()
}

// VacuumAnalyze performs VACUUM FULL; ANALYZE on the redshift database. This is useful for
// recreating the indices after a database has been modified and updating the query planner.
func (r *Redshift) VacuumAnalyze() error {
	_, err := r.logAndExec("VACUUM FULL; ANALYZE", false)
	return err
}

// VacuumAnalyzeTable performs VACUUM FULL; ANALYZE on a specific table. This is useful for
// recreating the indices after a database has been modified and updating the query planner.
func (r *Redshift) VacuumAnalyzeTable(schema, table string) error {
	_, err := r.logAndExec(fmt.Sprintf(`VACUUM FULL "%s"."%s"; ANALYZE "%s"."%s"`, schema, table, schema, table), false)
	return err
}
