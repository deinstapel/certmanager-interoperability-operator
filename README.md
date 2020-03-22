# certmanager-interoperability-operator

This small operators manages `kubernetes.io/tls` secrets to be interoperable between CertManager and traefik.
CertManager creates certificates containing a `ca.crt` key. This key contains the CA of this certificate.
With Traefik, we need this content to be in `tls.ca`.

This operator listens to all namespaces and copies the `ca.crt` key to the `tls.ca` key, if the `tls.ca` key does not exist or contains something else, which is not the same as the `ca.crt` key.

## Installation

We provide a manifest, which is ready to deploy. This manifest will deploy this operator within the `operators` namespace.

`kubectl apply -f deploy/certmanager-interoperability-operator.yaml`