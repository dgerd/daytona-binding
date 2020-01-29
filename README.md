# Daytona Binding

The Daytona Binding injects a [Daytona](https://github.com/cruise-automation/daytona) container spec into a Kubernetes Pod based upon a selector.

# Pre-requisites

* Install [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* Install [ko](https://github.com/google/ko)

# Install

To install the binding system and CRDs:

```
ko apply -f config/
```

This step sets up:

* Namespace
* Service account with permissions
* Webhook configuration
* Custom Resource Definitions
* Deployment (runs webhook and reconciler)

# Setup Binding

See [example.yaml](./example.yaml)


