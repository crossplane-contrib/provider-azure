# Building and Installing the Crossplane Azure Provider

`provider-azure` is composed of a golang project and can be built directly with standard `golang` tools. We currently support two different platforms for building:

* Linux: most modern distros should work although most testing has been done on Ubuntu
* Mac: macOS 10.6+ is supported

## Build Requirements

An Intel-based machine (recommend 2+ cores, 2+ GB of memory and 128GB of SSD). Inside your build environment (Docker for Mac or a VM), 6+ GB memory is also recommended.

The following tools are need on the host:

* curl
* docker (1.12+) or Docker for Mac (17+)
* git
* make
* golang
* rsync (if you're using the build container on mac)
* helm (v2.8.2+)
* kubebuilder (v1.0.4+)

## Build

You can build the Crossplane Azure Provider for the host platform by simply running the command below.
Building in parallel with the `-j` option is recommended.

```console
make -j4
```

The first time `make` is run, the build submodule will be synced and
updated. After initial setup, it can be updated by running `make submodules`.

Run `make help` for more options.

### Go support

To build with as-yet unvalidated go versions set GO_SUPPORTED_VERSIONS:

```console
GO_SUPPORTED_VERSIONS=1.15 make
```

## Building inside the cross container

Official Crossplane builds are done inside a build container. This ensures that we get a consistent build, test and release environment. To run the build inside the cross container run:

```console
> build/run make -j4
```

The first run of `build/run` will build the container itself and could take a few minutes to complete, but subsequent builds should go much faster.

## Run

To run the provider from source outside a k8s cluster it is necessary to register the CRDs and add a ProviderConfig before.

The necessary CRDs can be found in `package/crds/`. To register them in k8s run

```console
kubectl apply -f package/crds/
```

After that apply your custom ProviderConfig

```console
kubectl apply -f myProviderConfig.yaml
```

Then the provider can be run with

```console
make run
```

## Install

Installation instructions for local development builds can be found in the [Crossplane contributing guide](https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md#establishing-a-development-environment).
