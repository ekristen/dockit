# Managing PKI with Cert Manager

If you are using cert manager and would like to manage PKI for Dockit using it, the good news is you can. Furthermore this works with any backend that the cert-manager support so long as you can mount the certificate and corresponding private key into the Dockit container.

No matter how you request a certificate the result end's up in an secret that looks like the following.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: dockit-tls
type: kubernetes.io/tls
data:
  ca.crt: <PEM CA certificate>
  tls.key: <PEM private key>
  tls.crt: <PEM signed certificate chain>
  tls-combined.pem: <PEM private key + "\n" + PEM signed certificate chain>
  key.der: <DER binary format of private key>
```

For the dockit container, you need to pass the appropriate volume information to the deployment.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: dockit
spec:
  containers:
    - name: api-server
      image: ghcr.io/ekristen/dockit:latest
      command:
        - api-server
        - --pki-generate=false
        - --pki-file=/dockit/pki/combined.pem
      volumeMounts:
        - name: pki
          mountPath: /dockit/pki
  volumes:
    - name: pki
      secret:
        name: dockit-tls
        items:
          - key: tls-combined.pem
            path: combined.pem
```

## Bonus: Reloader by Stakator

You can handle automatic updates and restarts of Dockit by leveraging [Reloader](https://github.com/stakater/Reloader)

If you install reloader and then add the appropriate annotations to the dockit deployment, then when the PKI certificate rotates, reloader will automatically trigger a redeploy of dockit.

```yaml
metadata:
  annotations:
    reloader.stakater.com/auto: "true"
```
