<img src="https://github.com/m3dev/broom/assets/60843722/8860aaa0-e0b8-45cf-9b9e-59197dfc57a3" width="50%">

**broom** is a Kubernetes Custom Controller designed to gracefully handle Out of Memory (OOM) events in CronJobs by dynamically increasing memory limits.

It offers the following features:

* **Dynamic Memory Limit Adjustment**: Automatically increases the memory limit of containers that experience OOM events within CronJob specs.

* **Flexible Memory Adjustment Methods**: Users can choose between addition and multiplication methods for increasing memory limits.

* **Automatic Job Restart**: Optionally, broom can restart the failed jobs with updated spec for recovery.

* **Flexible Targeting**: Allows specifying target CronJobs using combinations of Name, Namespace, and Labels.

* **Slack Notification Integration**: Provides webhook notifications to Slack, informing users about the results of updates and restarts.

## Description

### How it works
When a `phase: Failed` Pod is detected, the reconciliation loop iterates as follows:

<img src="https://github.com/m3dev/broom/assets/60843722/df5219f3-d30f-47e9-84ef-a8c045cf1c1f" width="50%">

Upon identifying a Pod terminated due to `reason: OOMKilled`, the controller traces back through the `ownerReferences` from Pod to Job to CronJob.
Once the controller identifies the relevant CronJob, it prepares a modified specification with relaxed memory limits (`spec.jobTemplate.spec.containers[].resources.limits.memory`), such as doubling the original limit.
Subsequently, the controller updates the specification of the CronJob with the modified memory limits.
Optionally, the controller can restart the failed Job once using the updated memory limits. Finally, the controller sends a notification to Slack with the following content:


### Configuration

The controller can be configured using a `Broom` custom resource, which allows users to specify the following parameters:

```yaml
apiVersion: ai.m3.com/v1alpha1
kind: Broom
metadata:
  name: broom-sample
spec:
  target:
    name: oom-sample
    labels:
      m3.com/use-broom: "true"
    namespace: broom
  adjustment:
    type: Mul
    value: "2"
  restartPolicy: "OnOOM"
  slackWebhook:
    secret:
      namespace: default
      name: broom
      key: SLACK_WEBHOOK_URL
    channel: "#alert"
```

* **target** (optional): Specifies the target CronJobs to monitor for OOM events. Users can specify the target CronJobs using a combination of `name`, `namespace`, and `labels`. If not specified, the controller will monitor all CronJobs in the cluster.

* **adjustment** (required): Specifies the method and value for adjusting memory limits. Users can choose between `Add` and `Mul` methods for increasing memory limits, along with the corresponding value.

* **restartPolicy** (required): Specifies the policy for restarting failed Jobs. Users can choose between `Never` and `OnOOM` policies.

* **slackWebhook** (required): Specifies the Slack webhook integration for sending notifications. Users can provide the webhook URL using a Kubernetes Secret and specify the target channel (optional) for notifications.


## Getting Started

### Prerequisites
- go version v1.21.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/broom:tag
```

**NOTE:** This image ought to be published in the personal registry you specified. 
And it is required to have access to pull the image from the working environment. 
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/broom:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin 
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/broom:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/broom/<tag or branch>/dist/install.yaml
# kubectl apply -f https://raw.githubusercontent.com/m3dev/broom/main/dist/install.yaml
```


**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## Contributing

We welcome contributions from the community! If you'd like to contribute to this project, please follow these guidelines:

### Issues

If you encounter a bug, have a feature request, or would like to suggest an improvement, please open an issue on the GitHub repository. Make sure to provide detailed information about the problem or suggestion.

### Pull Requests

We gladly accept pull requests! Before submitting a pull request, please ensure the following:

1. Fork the repository and create your branch from `main`.
2. Ensure your code adheres to the project's coding standards.
3. Test your changes thoroughly.
4. Make sure your commits are descriptive and well-documented.
5. Update the README and any relevant documentation if necessary.

### Code of Conduct

Please note that this project is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report any unacceptable behavior.

Thank you for contributing to our project!

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
