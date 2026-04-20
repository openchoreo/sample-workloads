# WSO2 Micro Integrator — OpenChoreo Component Type

This directory contains the OpenChoreo platform resources required to build and deploy WSO2 Micro Integrator projects from source on a k3d cluster.

## Files

| File | Kind | Purpose |
|---|---|---|
| `micro-integrator.yaml` | `ClusterComponentType` | Defines the `deployment/micro-integrator` component type |
| `micro-integrator-builder.yaml` | `ClusterWorkflow` | Build pipeline: checkout → build → publish → generate workload |
| `micro-integrator-build.yaml` | `ClusterWorkflowTemplate` | Argo template that compiles the Maven project and packages it into the official `wso2/wso2mi` image |

## How it works

1. A `WorkflowRun` triggers the `micro-integrator-builder` workflow.
2. The **checkout** step clones the source repository.
3. The **build-image** step:
   - Detects the MI version from `project.runtime.version` in `pom.xml` (defaults to `4.4.0`).
   - Runs `mvn clean install` inside a Maven container to produce a `.car` file.
   - Builds a multi-stage image: copies the `.car` into `wso2/wso2mi:<version>` from Docker Hub.
4. The **publish-image** step pushes the image to the in-cluster registry.
5. The **generate-workload** step creates a `Workload` CR with the built image.

## Supported MI versions

The following tags from [`wso2/wso2mi`](https://hub.docker.com/r/wso2/wso2mi) are supported:

`1.1.0` · `1.2.0` · `4.0.0` · `4.1.0` · `4.2.0` · `4.3.0` · `4.4.0` · `4.5.0`

Set `project.runtime.version` in your root `pom.xml` to pin a version. If the property is absent, `4.4.0` is used.

## Prerequisites

A running OpenChoreo cluster with control plane, data plane, and workflow plane installed. Apply these resources once per cluster:

```bash
kubectl apply \
  -f micro-integrator-build.yaml \
  -f micro-integrator-builder.yaml \
  -f micro-integrator.yaml
```

## Deploying a component

Create a `Component` and `WorkflowRun` pointing at your Maven-based MI repository:

```yaml
apiVersion: openchoreo.dev/v1alpha1
kind: Component
metadata:
  name: my-mi-service
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
        url: "https://github.com/<org>/<repo>"
        revision:
          branch: "main"
        appPath: "<path-to-maven-project>"

---
apiVersion: openchoreo.dev/v1alpha1
kind: WorkflowRun
metadata:
  name: my-mi-service-build-01
  labels:
    openchoreo.dev/project: "default"
    openchoreo.dev/component: "my-mi-service"
spec:
  workflow:
    kind: ClusterWorkflow
    name: micro-integrator-builder
    parameters:
      repository:
        url: "https://github.com/<org>/<repo>"
        revision:
          branch: "main"
        appPath: "<path-to-maven-project>"
```

> **Note:** After the build completes, patch the generated `Workload` to declare the HTTP endpoint on port `8290`:
> ```bash
> kubectl patch workload my-mi-service-workload -n default --type=merge -p '{
>   "spec": {
>     "endpoints": {
>       "http": { "type": "HTTP", "port": 8290, "visibility": ["external"] }
>     }
>   }
> }'
> ```

### Private repositories

Create a `SecretReference` with your Git credentials and pass it via `repository.secretRef`:

```yaml
    parameters:
      repository:
        url: "https://github.com/<org>/<private-repo>"
        secretRef: "my-git-credentials"
        appPath: "."
```

## Environment configuration

The `deployment/micro-integrator` component type exposes the following per-environment settings:

| Parameter | Default | Description |
|---|---|---|
| `replicas` | `1` | Number of pod replicas |
| `resources.requests.cpu` | `500m` | CPU request |
| `resources.requests.memory` | `1Gi` | Memory request |
| `imagePullPolicy` | `IfNotPresent` | Kubernetes image pull policy |
| `mode` | `DEFAULT` | `DEFAULT` runs as a long-lived server; `TASK` runs a single CAR and exits |
| `carFile` | `""` | CAR file path passed to `--car` when `mode=TASK` |

## Source repository requirements

- Must contain a `pom.xml` at the `appPath` root.
- Maven build must produce at least one `.car` file.
- Optionally include `libs/`, `dropins/`, or `deployment.toml` alongside `pom.xml` for customisation.
