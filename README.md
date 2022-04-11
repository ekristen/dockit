# Dockit

Dead simple, yet powerful, authentication for any OCI registry.

## Overview

Dockit is a robust authentication system for an OCI compatible registry.

## Admin API

Dockit has an Admin API that is used to manage the permissions. Dockit supports users and groups and both can have permissions assigned to them.

### Authentication

Authentication to the Admin API is done via basic authentication using usernam/password. By default it will attempt to use docker credentials stored against the registry on your system, but the user has to have the admin flag set to true. If for some reason credentials cannot be obtained from the docker configuration, you can specify them on the command line.

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

## Development

Developing on Dockit is pretty straight forward you just need a golang development environment. Unless you are willing to take the time to setup trusted certificates, you'll need to modify the docker daemon configuration you are interacting with to add your test registry URI to the insecure registries list.

The `docker-compose.yml` in the repository is setup to deploy the docker distribution registry version 2 and point it at the `./hack/pki` folder to load the cert bundle.

1. `go run main.go api-server` -- start the api-server
2. `go run main.go init-container ./hack/pki/server/pem` -- run the init-container
3. `docker-compose up -d` -- start the docker registry

Develop as necessary.
