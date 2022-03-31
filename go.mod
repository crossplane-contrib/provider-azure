module github.com/crossplane/provider-azure

go 1.13

require (
	github.com/Azure/azure-pipeline-go v0.2.2 // indirect
	github.com/Azure/azure-sdk-for-go v61.4.0+incompatible
	github.com/Azure/azure-storage-blob-go v0.7.0
	// azure-sdk-for-go repository does not use go.mod so we need to maintain this dependency manually.
	github.com/Azure/go-autorest/autorest v0.11.18
	github.com/Azure/go-autorest/autorest/adal v0.9.13
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.0
	github.com/Azure/go-autorest/autorest/date v0.3.0
	github.com/Azure/go-autorest/autorest/to v0.3.0
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/crossplane/crossplane-runtime v0.15.1-0.20220315141414-988c9ba9c255
	github.com/crossplane/crossplane-tools v0.0.0-20220310165030-1f43fc12793e
	github.com/gofrs/uuid v4.2.0+incompatible // indirect
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.1.2
	github.com/mattn/go-ieproxy v0.0.0-20190805055040-f9202b1cfdeb // indirect
	github.com/mitchellh/copystructure v1.2.0
	github.com/onsi/gomega v1.17.0
	github.com/pkg/errors v0.9.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/client-go v0.23.0
	sigs.k8s.io/controller-runtime v0.11.0
	sigs.k8s.io/controller-tools v0.8.0
)
