# Stack-Azure

## Overview

This `stack-azure` repository is the implementation of a Crossplane infrastructure
[stack](https://github.com/crossplane/crossplane/blob/master/design/design-doc-stacks.md) for
[Microsoft Azure](https://azure.microsoft.com/).
The stack that is built from the source code in this repository can be installed into a Crossplane control plane and adds the following new functionality:

* Custom Resource Definitions (CRDs) that model Azure infrastructure and services (e.g. [Azure Database for MySQL](https://azure.microsoft.com/en-us/services/mysql/), [AKS clusters](https://azure.microsoft.com/en-us/services/kubernetes-service/), etc.)
* Controllers to provision these resources in Azure based on the users desired state captured in CRDs they create
* Implementations of Crossplane's [portable resource abstractions](https://crossplane.io/docs/master/running-resources.html), enabling Azure resources to fulfill a user's general need for cloud services

## Getting Started and Documentation

For getting started guides, installation, deployment, and administration, see our [Documentation](https://crossplane.io/docs/latest).

## Contributing

Stack-Azure is a community driven project and we welcome contributions.
See the Crossplane [Contributing](https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md) guidelines to get started.

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please open an [issue](https://github.com/crossplane/stack-azure/issues).

## Contact

Please use the following to reach members of the community:

- Slack: Join our [slack channel](https://slack.crossplane.io)
- Forums: [crossplane-dev](https://groups.google.com/forum/#!forum/crossplane-dev)
- Twitter: [@crossplane_io](https://twitter.com/crossplane_io)
- Email: [info@crossplane.io](mailto:info@crossplane.io)

## Roadmap

Stack-Azure goals and milestones are currently tracked in the Crossplane repository.
More information can be found in [ROADMAP.md](https://github.com/crossplane/crossplane/blob/master/ROADMAP.md).

## Governance and Owners

Stack-Azure is run according to the same [Governance](https://github.com/crossplane/crossplane/blob/master/GOVERNANCE.md) and [Ownership](https://github.com/crossplane/crossplane/blob/master/OWNERS.md) structure as the core Crossplane project.

## Code of Conduct

Stack-Azure adheres to the same [Code of Conduct](https://github.com/crossplane/crossplane/blob/master/CODE_OF_CONDUCT.md) as the core Crossplane project.

## Licensing

Stack-Azure is under the Apache 2.0 license.

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fcrossplaneio%2Fstack-azure.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fcrossplaneio%2Fstack-azure?ref=badge_large)
