# Hello World MI

A minimal REST API that responds with `{"Hello": "World"}` — a good starting point for deploying WSO2 Micro Integrator projects on OpenChoreo.

**Source:** [wso2/choreo-samples — hello-world-mi](https://github.com/wso2/choreo-samples/tree/main/hello-world-mi)

## Endpoint

```
GET /HelloWorld  →  {"Hello": "World"}
```

## Prerequisites

- OpenChoreo cluster running with control plane, data plane, and workflow plane installed.
- The `deployment/micro-integrator` component type and `micro-integrator-builder` workflow applied from the parent directory:

```bash
kubectl apply \
  -f ../micro-integrator-build.yaml \
  -f ../micro-integrator-builder.yaml \
  -f ../micro-integrator.yaml
```

## Deploy

### 1. Create the Component and trigger a build

```bash
kubectl apply -f - <<EOF
apiVersion: openchoreo.dev/v1alpha1
kind: Component
metadata:
  name: hello-world-mi
  namespace: default
spec:
  owner:
    projectName: default
  componentType:
    kind: ClusterComponentType
    name: deployment/micro-integrator
  autoDeploy: true
  workflow:
    kind: ClusterWorkflow
    name: micro-integrator-builder
    parameters:
      repository:
        url: "https://github.com/wso2/choreo-samples"
        revision:
          branch: "main"
        appPath: "hello-world-mi"
---
apiVersion: openchoreo.dev/v1alpha1
kind: WorkflowRun
metadata:
  name: hello-world-mi-build-01
  labels:
    openchoreo.dev/project: "default"
    openchoreo.dev/component: "hello-world-mi"
spec:
  workflow:
    kind: ClusterWorkflow
    name: micro-integrator-builder
    parameters:
      repository:
        url: "https://github.com/wso2/choreo-samples"
        revision:
          branch: "main"
        appPath: "hello-world-mi"
EOF
```

### 2. Watch the build

```bash
kubectl get workflow hello-world-mi-build-01 -n workflows-default --watch
```

The build runs four steps: `checkout-source` → `build-image` → `publish-image` → `generate-workload-cr`. The `build-image` step downloads Maven dependencies and pulls `wso2/wso2mi:4.4.0` — expect 3–5 minutes on first run.

### 3. Declare the HTTP endpoint

The build generates a `Workload` without endpoint metadata. Patch it to expose port 8290:

```bash
kubectl patch workload hello-world-mi-workload -n default --type=merge -p '{
  "spec": {
    "endpoints": {
      "http": { "type": "HTTP", "port": 8290, "visibility": ["external"] }
    }
  }
}'
```

### 4. Verify the deployment

```bash
kubectl get deployment -A -l openchoreo.dev/component=hello-world-mi
```

### 5. Get the URL and invoke

```bash
HOSTNAME=$(kubectl get releasebinding -n default \
  -l openchoreo.dev/component=hello-world-mi \
  -o jsonpath='{.items[0].status.endpoints[0].externalURLs.http.host}')

PATH_PREFIX=$(kubectl get releasebinding -n default \
  -l openchoreo.dev/component=hello-world-mi \
  -o jsonpath='{.items[0].status.endpoints[0].externalURLs.http.path}')

curl "http://${HOSTNAME}:19080${PATH_PREFIX}/HelloWorld"
```

Expected response:

```json
{"Hello": "World"}
```

## Clean up

```bash
kubectl delete component hello-world-mi -n default
kubectl delete workload hello-world-mi-workload -n default
```
