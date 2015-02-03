package redshift

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/clever/redshifter/postgres"
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
	tmpprefix = flag.String("tmptableprefix", "tmp_refresh_table_",
		"Prefix for temporary tables to ensure they don't collide with existing ones.")
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
func (r *Redshift) CopyJSONDataFromS3(table, file, jsonpathsFile, awsRegion string) error {
	copyCmd := fmt.Sprint(
		"COPY ", table, " FROM '", file, "' WITH",
		" json '", jsonpathsFile,
		"' region '", awsRegion,
		"' timeformat 'epochsecs'",
		" COMPUPDATE ON")
	_, err := r.logAndExec(copyCmd, true)
	return err
}

// CopyGzipCsvDataFromS3 copies gzipped CSV data from an S3 file into a redshift table.
func (r *Redshift) CopyGzipCsvDataFromS3(table, file, awsRegion string, delimiter rune) error {
	copyCmd := fmt.Sprint(
		"COPY ", table, " FROM '", file, "' WITH ",
		"REGION '", awsRegion, "'",
		" GZIP CSV DELIMITER '", string(delimiter), "'",
		" IGNOREHEADER 0 ACCEPTINVCHARS TRUNCATECOLUMNS TRIMBLANKS BLANKSASNULL EMPTYASNULL DATEFORMAT 'auto' ACCEPTANYDATE COMPUPDATE ON")
	_, err := r.logAndExec(copyCmd, true)
	return err
}

// TODO: explore adding distkey.
func columnString(c *postgres.ColInfo) string {
	attributes := []string{}
	constraints := []string{}
	if c.DefaultVal != "" {
		attributes = append(attributes, "DEFAULT "+c.DefaultVal)
	}
	if c.PrimaryKey {
		attributes = append(attributes, "SORTKEY")
		constraints = append(constraints, "PRIMARY KEY")
	}
	if c.NotNull {
		constraints = append(constraints, "NOT NULL")
	}
	return fmt.Sprintf("%s %s %s %s", c.Name, c.ColType, strings.Join(attributes, " "), strings.Join(constraints, " "))
}

func (r *Redshift) createTable(name string, ts postgres.TableSchema) error {
	colStrings := []string{}
	sort.Sort(ts)
	for _, ci := range ts {
		colStrings = append(colStrings, columnString(ci))
	}
	cmd := fmt.Sprintf("CREATE TABLE %s (%s)", name, strings.Join(colStrings, ", "))
	_, err := r.logAndExec(cmd, false)
	return err
}

// RefreshTable refreshes a single table by first copying gzipped CSV data into a temporary table
// and later renaming the temporary table to the original one.
func (r *Redshift) RefreshTable(name, prefix, file, awsRegion string, ts postgres.TableSchema, delim rune) error {
	tmptable := prefix + name
	if _, err := r.logAndExec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tmptable), false); err != nil {
		return err
	}
	if err := r.createTable(tmptable, ts); err != nil {
		return err
	}
	if err := r.CopyGzipCsvDataFromS3(tmptable, file, awsRegion, delim); err != nil {
		return err
	}
	_, err := r.logAndExec(fmt.Sprintf("DROP TABLE %s; ALTER TABLE %s RENAME TO %s;", name, tmptable, name), false)
	return err
}

// RefreshTables refreshes multiple tables in parallel and returns an error if any of the copies
// fail.
func (r *Redshift) RefreshTables(
	tables map[string]postgres.TableSchema, s3prefix, awsRegion string, delim rune) error {
	group := new(errgroup.Group)
	for name, ts := range tables {
		group.Add(1)
		go func(name string, ts postgres.TableSchema) {
			if err := r.RefreshTable(name, *tmpprefix, postgres.S3Filename(s3prefix, name), awsRegion, ts, delim); err != nil {
				group.Error(err)
			}
			group.Done()
		}(name, ts)
	}
	return group.Wait()
}

// VacuumAnalyze performs VACUUM FULL; ANALYZE on the redshift database. This is useful for
// recreating the indices after a database has been modified and updating the query planner.
func (r *Redshift) VacuumAnalyze() error {
	_, err := r.logAndExec("VACUUM FULL; ANALYZE", false)
	return err
}
