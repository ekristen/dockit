# Dockit

Dead simple, yet powerful, authentication for the docker distribution registry.

If you are looking for a way to authenticate your docker registry without requiring a super computer, Dockit is for you.

## Registries

### Docker Distribution

Authentication with distribution is based on tokens and PKI, the x509 version of the public key has to be made available to the distribution instance upon start. Dockit comes with an init-container capability that will fetch all known keys and return them so they can be written to disk for the registry to use.

**Note:** unfortunately the distribution registry only reads the cert bundle on start and requires a restart to include any new ones.

#### Configuration

It's fairly straight forward to configure the docker distribution registry to delegate authentication to Dockit.

The easiest and recommended method is via environment variables.

The following environment values should be set:

- `REGISTRY_AUTH_TOKEN_SERVICE` this is the name of your service
- `REGISTRY_AUTH_TOKEN_ISSUER` with value `dockit`
- `REGISTRY_AUTH_TOKEN_REALM` this should be the https URL of where dockit is listening (example: <https://dockit.private.io/v2/token>)
- `REGISTRY_AUTH_TOKEN_ROOTCERTBUNDLE` should be a pem that has all valid signing certs, if using dockit init-conatiner use `/dockit/certs.pem`

## SQL

Built and designed with SQLite3 with the intention to use [litestream](https://litestream.io). MySQL should work.

## CLI

```help
NAME:
   dockit - dead simple, yet powerful, docker registry authentication

USAGE:
   dockit [global options] command [command options] [arguments...]

VERSION:
   0.1.0-dev-dirty

AUTHOR:
   Erik Kristensen <erik@erikkristensen.com>

COMMANDS:
   api-server      dockit api server
   version         print version
   init-container  Provides init container capability to fetch and store PKI from Dockit before starting registry server
   pki-generate    generates an ecdsa private key and certificate
   rbac            provides the ability to perform various RBAC related actions
   help, h         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

### API Server

```help
NAME:
   main api-server - dockit api server

USAGE:
   main api-server [command options] [arguments...]

OPTIONS:
   --node-id value              Unique ID of the Node (this should be increased for each replica) 0-1023 (1024 will select a random number between 0-1023) (default: 1024) [$DOCKIT_NODE_ID, $NODE_ID]
   --pki-generate               whether or not to generate PKI if false, you must specify --pki-file (default: true) [$DOCKIT_PKI_GENERATE, $PKI_GENERATE]
   --pki-file value             file to read PKI data from [$DOCKIT_PKI_FILE, $PKI_FILE]
   --pki-key-type value         Algorithm to use for PKI for Registry to Dockit authentication (default: "ec") [$DOCKIT_PKI_KEY_TYPE, $PKI_KEY_TYPE]
   --pki-ec-key-size value      Elliptic Curve Key Size (default: 256) [$DOCKIT_PKI_EC_KEY_SIZE, $PKI_EC_KEY_SIZE]
   --pki-rsa-key-size value     RSA Key Size (default: 4096) [$DOCKIT_PKI_RSA_KEY_SIZE, $PKI_RSA_KEY_SIZE]
   --pki-cert-years value       The number of years that internal PKI certs are good for. (default: 2) [$DOCKIT_PKI_CERT_YEARS, $PKI_CERT_YEARS]
   --port value                 Port for the HTTP Server Port (default: 4315) [$DOCKIT_PORT, $PORT]
   --metrics-port value         Port for the metrics and debug http server to listen on (default: 4316) [$METRICS_PORT, $DOCKIT_METRICS_PORT]
   --sql-dialect value          The type of sql to use, sqlite or mysql (default: "sqlite") [$DOCKIT_SQL_DIALECT, $SQL_DIALECT]
   --sql-dsn value              The DSN to use to connect to (default: "file:dockit.sqlite") [$DOCKIT_SQL_DSN, $SQL_DSN]
   --root-user value            Root Username [$DOCKIT_ROOT_USER, $ROOT_USER]
   --root-password value        Root Password [$DOCKIT_ROOT_PASSWORD, $ROOT_PASSWORD]
   --first-user-admin           Indicates if the first user to login should be made an admin (default: true) [$DOCKIT_FIRST_USER_ADMIN, $FIRST_USER_ADMIN]
   --log-level value, -l value  Log Level (default: "info") [$LOGLEVEL]
   --log-caller                 log the caller (aka line number and file) (default: false)
   --log-disable-color          disable log coloring (default: false)
   --log-full-timestamp         force log output to always show full timestamp (default: false)
   --help, -h                   show help (default: false)
```

## API

Besides the token endpoint the registry uses to request tokens for authentication purposes, dockit comes with an admin API that can be used to manage users, groups, permissions and PKI data.

### Authentication

Authentication to the Admin API is done via basic authentication using usernam/password. By default it will attempt to use docker credentials stored against the registry on your system, but the user has to have the admin flag set to true. If for some reason credentials cannot be obtained from the docker configuration, you can specify them on the command line.

## PKI

A private key and it's corresponding X509 certificate are used to sign and verify the tokens generated by dockit. Dockit uses the private key, Docker Distribution uses the X509 certificate to verify the token.

By default dockit will generate an EC private key and corresponding x509 certificate that store it in it's database, it will then serve the certificate up on an API endpoint that can be used by the `init-container` subcommand that can be placed infront of the docker distribution registry to ensure the current certificates used to verify tokens are available.

### Bring Your Own

Dockit supports bringing your own PKI via the `--pki-generate=false` and `--pki-file=<file>` command. This file must contain a private key (EC or RSA) with a corresponding X509 certificate both in PEM format.

With Bring Your Own, you can also use tools like cert-manager to manage certificates for you, see our guide on [using cert-manager](docs/guides/cert-manager.md).

## Development

Developing on Dockit is pretty straight forward you just need a golang development environment. Unless you are willing to take the time to setup trusted certificates, you'll need to modify the docker daemon configuration you are interacting with to add your test registry URI to the insecure registries list.

First you will want to start the api-server, then you can use docker-compose to fire up the registry and leverage the init-container subcommand from dockit to ensure the certificate is in place. If you rotate or clear your sqlite database, you'll want to re-deploy the docker-compose stack to ensure you get the latest certificates loaded.

1. `go run main.go api-server` -- start the api-server
2. `docker-compose up -d` -- start the docker registry

Develop as necessary.
