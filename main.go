package main

import (
	"flag"
	"log"
	"strings"

	"github.com/clever/go-utils/flagutil"
	"github.com/clever/redshifter/postgres"
	"github.com/clever/redshifter/redshift"
	"github.com/segmentio/go-env"
)

var (
	awsRegion      = env.MustGet("AWS_REGION")
	s3prefix       = flagutil.RequiredStringFlag("s3prefix", "s3 path to be used as a prefix for temporary storage of postgres data", nil)
	tablesCSV      = flagutil.RequiredStringFlag("tables", "Tables to copy as CSV", nil)
	dumppg         = flag.Bool("dumppostgres", true, "Whether to dump postgres")
	updateRS       = flag.Bool("updateredshift", true, "Whether to replace redshift")
	redshiftSchema = flag.String("redshiftschema", "public", "Schema name to store the tables.")
)

func main() {
	flag.Parse()
	if err := flagutil.ValidateFlags(nil); err != nil {
		log.Fatal(err.Error())
	}
	tables := strings.Split(*tablesCSV, ",")

	pgdb := postgres.NewDB(postgres.Config{PoolSize: len(tables)})
	defer pgdb.Close()
	tsmap, err := pgdb.GetTableSchemas(tables, "")
	if err != nil {
		log.Fatal(err)
	}
	if *dumppg {
		if err := pgdb.DumpTablesToS3(tables, *s3prefix); err != nil {
			log.Fatal(err)
		}
		log.Println("POSTGRES DUMPED TO S3")
	}
	if *updateRS {
		r, err := redshift.NewRedshift()
		defer r.Close()
		if err != nil {
			log.Fatal(err)
		}
		if err := r.RefreshTables(tsmap, *redshiftSchema, *s3prefix, awsRegion, '|'); err != nil {
			log.Fatal(err)
		}
		if err := r.VacuumAnalyze(); err != nil {
			log.Fatal(err)
		}
	}
}
