#!/bin/env bash

set -e

pushd $1

patch --ignore-whitespace << 'EOF'
--- a/manifest.yaml
+++ b/manifest.yaml
@@ -32,6 +32,9 @@ data:
           "ipam": {
               "type": "calico-ipam"
           },
+          "container_settings": {
+              "allow_ip_forwarding": true
+          },
           "policy": {
               "type": "k8s"
           },
EOF

popd
