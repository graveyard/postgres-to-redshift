# postgres-to-redshift

`postgres-to-redshift` copies postgres data to redshift via S3.
This repository is golang adaptation of the original script by Donors Choose at https://github.com/DonorsChoose/open-data-science/tree/master/postgres2redshift.

## Running

```bash
AWS_REGION='us-east-1' \
godep go run main.go \
-redshifthost=<host> \
-redshiftport=<port> \
-redshiftuser=<user> \
-redshiftpassword=<password> \
-redshiftdatabase=<database> \
-postgreshost=<host> \
-postgresdatabase=<database> \
-postgresuser=<user> \
-postgresport=<port> \
-postgrespassword=<password> \
-s3prefix=<prefix> \
-tables=<tables_csv>
```

In production, the binary is run on gearman using a standalone worker run as a cron job.
