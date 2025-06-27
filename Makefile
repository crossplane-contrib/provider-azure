# ====================================================================================
# Setup Project

PROJECT_NAME := provider-azure
PROJECT_REPO := github.com/crossplane-contrib/$(PROJECT_NAME)

PLATFORMS ?= linux_amd64 linux_arm64

# kind-related versions
KIND_VERSION ?= v0.12.0
KIND_NODE_IMAGE_TAG ?= v1.23.4

# -include will silently skip missing files, which allows us
# to load those files with a target in the Makefile. If only
# "include" was used, the make command would fail and refuse
# to run a target until the include commands succeeded.
-include build/makelib/common.mk

# ====================================================================================
# Setup Output

-include build/makelib/output.mk

# ====================================================================================
# Setup Go

# Set a sane default so that the nprocs calculation below is less noisy on the initial
# loading of this file
NPROCS ?= 1

# each of our test suites starts a kube-apiserver and running many test suites in
# parallel can lead to high CPU utilization. by default we reduce the parallelism
# to half the number of CPU cores.
GO_TEST_PARALLEL := $(shell echo $$(( $(NPROCS) / 2 )))

GO_STATIC_PACKAGES = $(GO_PROJECT)/cmd/provider
GO_LDFLAGS += -X $(GO_PROJECT)/pkg/version.Version=$(VERSION)
GO_SUBDIRS += cmd pkg apis
GO111MODULE = on
-include build/makelib/golang.mk

# ====================================================================================
# Setup Kubernetes tools

UP_VERSION = v0.13.0
UP_CHANNEL = stable
-include build/makelib/k8s_tools.mk

# ====================================================================================
# Setup Images

IMAGES = provider-azure 
-include build/makelib/imagelight.mk

# ====================================================================================
# Setup XPKG

XPKG_REG_ORGS ?= xpkg.upbound.io/crossplane-contrib index.docker.io/crossplanecontrib
# NOTE(hasheddan): skip promoting on xpkg.upbound.io as channel tags are
# inferred.
XPKG_REG_ORGS_NO_PROMOTE ?= xpkg.upbound.io/crossplane-contrib
XPKGS = provider-azure
-include build/makelib/xpkg.mk

# NOTE(hasheddan): we force image building to happen prior to xpkg build so that
# we ensure image is present in daemon.
xpkg.build.provider-azure: do.build.images

# ====================================================================================
# Targets

# run `make help` to see the targets and options

# We want submodules to be set up the first time `make` is run.
# We manage the build/ folder and its Makefiles as a submodule.
# The first time `make` is run, the includes of build/*.mk files will
# all fail, and this target will be run. The next time, the default as defined
# by the includes will be run instead.
fallthrough: submodules
	@echo Initial setup complete. Running make again . . .
	@make

# NOTE(hasheddan): the build submodule currently overrides XDG_CACHE_HOME in
# order to force the Helm 3 to use the .work/helm directory. This causes Go on
# Linux machines to use that directory as the build cache as well. We should
# adjust this behavior in the build submodule because it is also causing Linux
# users to duplicate their build cache, but for now we just make it easier to
# identify its location in CI so that we cache between builds.
go.cachedir:
	@go env GOCACHE

# NOTE(hasheddan): we must ensure up is installed in tool cache prior to build
# as including the k8s_tools machinery prior to the xpkg machinery sets UP to
# point to tool cache.
build.init: $(UP)

# Generate a coverage report for cobertura applying exclusions on
# - generated file
cobertura:
	@cat $(GO_TEST_OUTPUT)/coverage.txt | \
		grep -v zz_generated.deepcopy | \
		$(GOCOVER_COBERTURA) > $(GO_TEST_OUTPUT)/cobertura-coverage.xml

# Ensure a PR is ready for review.
reviewable: generate lint
	@go mod tidy

# Ensure branch is clean.
check-diff: reviewable
	@$(INFO) checking that branch is clean
	@test -z "$$(git status --porcelain)" || $(FAIL)
	@$(OK) branch is clean

# integration tests
e2e.run: test-integration

# Run integration tests.
test-integration: $(KIND) $(KUBECTL) $(HELM3)
	@$(INFO) running integration tests using kind $(KIND_VERSION)
	@KIND_NODE_IMAGE_TAG=${KIND_NODE_IMAGE_TAG} KIND_VERSION=${KIND_VERSION} $(ROOT_DIR)/cluster/local/integration_tests.sh || $(FAIL)
	@$(OK) integration tests passed

# Update the submodules, such as the common build scripts.
submodules:
	@git submodule sync
	@git submodule update --init --recursive

# This is for running out-of-cluster locally, and is for convenience. Running
# this make target will print out the command which was used. For more control,
# try running the binary directly with different arguments.
run: go.build
	@$(INFO) Running Crossplane locally out-of-cluster . . .
	@# To see other arguments that can be provided, run the command with --help instead
	$(GO_OUT_DIR)/provider --debug

dev: $(KIND) $(KUBECTL)
	@$(INFO) Creating kind cluster
	@$(KIND) create cluster --name=$(PROJECT_NAME)-dev
	@$(KUBECTL) cluster-info --context kind-$(PROJECT_NAME)-dev
	@$(INFO) Installing Crossplane CRDs
	@$(KUBECTL) apply -k https://github.com/crossplane/crossplane//cluster?ref=master
	@$(INFO) Installing Provider Openshift CRDs
	@$(KUBECTL) apply -R -f package/crds
	@$(INFO) Starting Provider Openshift controllers
	@$(GO) run cmd/provider/main.go --debug

dev-clean: $(KIND) $(KUBECTL)
	@$(INFO) Deleting kind cluster
	@$(KIND) delete cluster --name=$(PROJECT_NAME)-dev

manifests:
	@$(WARN) Deprecated. Please run make generate instead.

# TODO(negz): This is only used for the AKS tests. Remove it once we test AKS
# using unit tests.
KUBEBUILDER_VERSION ?= 1.0.8
KUBEBUILDER := $(TOOLS_HOST_DIR)/kubebuilder-$(KUBEBUILDER_VERSION)
KUBEBUILDER_OS ?= $(GOHOSTOS)
KUBEBUILDER_ARCH ?= $(GOHOSTARCH)
TEST_ASSET_KUBE_APISERVER := $(KUBEBUILDER)/kube-apiserver
TEST_ASSET_ETCD := $(KUBEBUILDER)/etcd
export TEST_ASSET_KUBE_APISERVER TEST_ASSET_ETCD

test.init: $(KUBEBUILDER)

.PHONY: cobertura reviewable submodules fallthrough test-integration run manifests crds.clean

# ====================================================================================
# Special Targets
# Install gomplate
GOMPLATE_VERSION := 3.10.0
GOMPLATE := $(TOOLS_HOST_DIR)/gomplate-$(GOMPLATE_VERSION)

$(GOMPLATE):
	@$(INFO) installing gomplate $(SAFEHOSTPLATFORM)
	@mkdir -p $(TOOLS_HOST_DIR)
	@curl -fsSLo $(GOMPLATE) https://github.com/hairyhenderson/gomplate/releases/download/v$(GOMPLATE_VERSION)/gomplate_$(SAFEHOSTPLATFORM) || $(FAIL)
	@chmod +x $(GOMPLATE)
	@$(OK) installing gomplate $(SAFEHOSTPLATFORM)

export GOMPLATE

# This target prepares repo for your provider by replacing all "openshift"
# occurrences with your provider name.
# This target can only be run once, if you want to rerun for some reason,
# consider stashing/resetting your git state.
# Arguments:
#   provider: Camel case name of your provider, e.g. GitHub, PlanetScale
provider.prepare:
	@[ "${provider}" ] || ( echo "argument \"provider\" is not set"; exit 1 )
	@PROVIDER=$(provider) ./hack/helpers/prepare.sh

# This target adds a new api type and its controller.
# You would still need to register new api in "apis/<provider>.go" and
# controller in "internal/controller/<provider>.go".
# Arguments:
#   provider: Camel case name of your provider, e.g. GitHub, PlanetScale
#   group: API group for the type you want to add.
#   kind: Kind of the type you want to add
#	apiversion: API version of the type you want to add. Optional and defaults to "v1alpha1"
provider.addtype: $(GOMPLATE)
	@[ "${provider}" ] || ( echo "argument \"provider\" is not set"; exit 1 )
	@[ "${group}" ] || ( echo "argument \"group\" is not set"; exit 1 )
	@[ "${kind}" ] || ( echo "argument \"kind\" is not set"; exit 1 )
	@PROVIDER=$(provider) GROUP=$(group) KIND=$(kind) APIVERSION=$(apiversion) ./hack/helpers/addtype.sh


define CROSSPLANE_MAKE_HELP
Crossplane Targets:
    cobertura             Generate a coverage report for cobertura applying exclusions on generated files.
    reviewable            Ensure a PR is ready for review.
    submodules            Update the submodules, such as the common build scripts.
    run                   Run crossplane locally, out-of-cluster. Useful for development.

endef
# The reason CROSSPLANE_MAKE_HELP is used instead of CROSSPLANE_HELP is because the crossplane
# binary will try to use CROSSPLANE_HELP if it is set, and this is for something different.
export CROSSPLANE_MAKE_HELP

crossplane.help:
	@echo "$$CROSSPLANE_MAKE_HELP"

help-special: crossplane.help

.PHONY: crossplane.help help-special

# TODO(negz): This is only used for the AKS tests. Remove it once we test AKS
# using unit tests.
$(KUBEBUILDER):
	@$(INFO) installing kubebuilder $(KUBEBUILDER_VERSION)
	@mkdir -p $(TOOLS_HOST_DIR)/tmp || $(FAIL)
	@curl -fsSL https://github.com/kubernetes-sigs/kubebuilder/releases/download/v$(KUBEBUILDER_VERSION)/kubebuilder_$(KUBEBUILDER_VERSION)_$(KUBEBUILDER_OS)_$(KUBEBUILDER_ARCH).tar.gz | tar -xz -C $(TOOLS_HOST_DIR)/tmp  || $(FAIL)
	@mv $(TOOLS_HOST_DIR)/tmp/kubebuilder_$(KUBEBUILDER_VERSION)_$(KUBEBUILDER_OS)_$(KUBEBUILDER_ARCH)/bin $(KUBEBUILDER) || $(FAIL)
	@rm -fr $(TOOLS_HOST_DIR)/tmp
	@$(OK) installing kubebuilder $(KUBEBUILDER_VERSION)
