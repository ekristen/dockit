# Init Container Support

Dockit comes with an `init-container` subcommand that's designed to be used with Kubernetes init containers. It can even be used with the latest version of docker compose.

The `init-container` subcommand is designed to make an http call against the dockit api server to retrieve all non-expired certificates in it's database and write those to a file. The docker distribution registry can then read the file from disk on start and used to validate JWT tokens.

## Kubernetes

With kubernetes, you can use the standard `initContainers` syntax to run the subcommand prior to the docker distribution registry starting.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: registry
spec:
  initContainers:
  - name: dockit-certs
    image: ghcr.io/ekristen/dockit:latest
    command:
      - init-container
      - /dockit/pki/bundle.pem
  containers:
  - name: server
    image: registry:2
    env:
      - name: REGISTRY_AUTH_TOKEN_ISSUER
        value: dockit
      - name: REGISTRY_AUTH_TOKEN_REALM
        value: http://host.docker.internal:4315/v2/token
      - name: REGISTRY_AUTH_TOKEN_ROOTCERTBUNDLE
        value: /dockit/pki/bundle.pem
      - name: REGISTRY_AUTH_TOKEN_SERVICE
        value: registry
```

## Docker Compose

With docker compose, you can use version 3 and the latest docker compose binary you can use the newer depends_on syntax to make containers conditionally wait on each other.

```yaml
version: '3'
services:
  chown:
    image: ubuntu
    command: chown -R 999:999 /dockit
    volumes:
      - "pki:/dockit/pki"

  init:
    image: ghcr.io/ekristen/dockit:v0.2.4
    command:
      - init-container
      - /dockit/pki/bundle.pem
    restart: "no"
    environment:
      DOCKIT_BASE_URL: http://host.docker.internal:4315/v2
    volumes:
      - "pki:/dockit/pki"
    depends_on:
      chown:
        condition: service_completed_successfully

  registry:
    image: registry:2
    environment:
      REGISTRY_AUTH_TOKEN_ISSUER: dockit
      REGISTRY_AUTH_TOKEN_REALM: http://host.docker.internal:4315/v2/token
      REGISTRY_AUTH_TOKEN_ROOTCERTBUNDLE: /dockit/pki/bundle.pem
      REGISTRY_AUTH_TOKEN_SERVICE: registry
    volumes:
      - "pki:/dockit/pki"
    ports:
      - 5000:5000
    depends_on:
      init:
        condition: service_completed_successfully

volumes:
  pki:
```
