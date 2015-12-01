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
-redshiftschema=<schema> \
-postgreshost=<host> \
-postgresdatabase=<database> \
-postgresuser=<user> \
-postgresport=<port> \
-postgrespassword=<password> \
-s3prefix=<prefix> \
-tables=<tables_csv>
```

In production, the binary is run on gearman using a standalone worker run as a cron job.

## Changing Dependencies

### New Packages

When adding a new package, you can simply use `make vendor` to update your imports.
This should bring in the new dependency that was previously undeclared.
The change should be reflected in [Godeps.json](Godeps/Godeps.json) as well as [vendor/](vendor/).

### Existing Packages

First ensure that you have your desired version of the package checked out in your `$GOPATH`.

When to change the version of an existing package, you will need to use the godep tool.
You must specify the package with the `update` command, if you use multiple subpackages of a repo you will need to specify all of them.
So if you use package github.com/Clever/foo/a and github.com/Clever/foo/b, you will need to specify both a and b, not just foo.

```
# depending on github.com/Clever/foo
godep update github.com/Clever/foo

# depending on github.com/Clever/foo/a and github.com/Clever/foo/b
godep update github.com/Clever/foo/a github.com/Clever/foo/b
```

