# Daytona Binding

The Daytona Binding injects a [Daytona](https://github.com/cruise-automation/daytona) container spec into a Kubernetes Pod based upon a selector.

# Pre-requisites

* Install [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* Install [ko](https://github.com/google/ko)

# Install

To install the binding system and CRDs:

```
KO_DOCKER_REPO=gcr.io/cruise-gcr-dev ko apply -f config/
```

This step sets up:

* Namespace
* Service account with permissions
* Webhook configuration
* Custom Resource Definitions
* Deployment (runs webhook and reconciler)

# Example Binding

See [example-binding.yaml](./example-binding.yaml)

# Example Usage

See [example-ksvc.yaml](./example-ksvc.yaml)
