#!/usr/bin/env bash

/usr/local/bin/postgres-to-redshift \
-redshifthost=$REDSHIFT_HOST \
-redshiftport=$REDSHIFT_PORT \
-redshiftuser=$REDSHIFT_USER \
-redshiftpassword=$REDSHIFT_PASSWORD \
-redshiftdatabase=$REDSHIFT_DATABASE \
-redshiftschema=$REDSHIFT_SCHEMA \
-postgreshost=$POSTGRES_HOST \
-postgresdatabase=$POSTGRES_DATABASE \
-postgresuser=$POSTGRES_USER \
-postgresport=$POSTGRES_PORT \
-postgrespassword=$POSTGRES_PASSWORD \
-s3prefix=$S3PREFIX \
-tables=$TABLES_TO_COPY
