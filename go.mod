module github.com/crossplane/provider-azure

go 1.16

require (
	github.com/Azure/azure-pipeline-go v0.2.2 // indirect
	github.com/Azure/azure-sdk-for-go v61.4.0+incompatible
	github.com/Azure/azure-sdk-for-go/sdk/azcore v0.22.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v0.13.2
	github.com/Azure/azure-storage-blob-go v0.7.0
	// azure-sdk-for-go repository does not use go.mod so we need to maintain this dependency manually.
	github.com/Azure/go-autorest/autorest v0.11.18
	github.com/Azure/go-autorest/autorest/adal v0.9.13
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.0
	github.com/Azure/go-autorest/autorest/date v0.3.0
	github.com/Azure/go-autorest/autorest/to v0.3.0
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/crossplane/crossplane-runtime v0.14.0
	github.com/crossplane/crossplane-tools v0.0.0-20210320162312-1baca298c527
	github.com/gofrs/uuid v4.2.0+incompatible // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/google/go-cmp v0.5.2
	github.com/google/uuid v1.3.0
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-ieproxy v0.0.0-20190805055040-f9202b1cfdeb // indirect
	github.com/mitchellh/copystructure v1.2.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/gomega v1.10.2
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pkg/errors v0.9.1
	golang.org/x/crypto v0.0.0-20220321153916-2c7772ba3064 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8 // indirect
	golang.org/x/tools v0.0.0-20200916195026-c9a70fc28ce3 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.1
	sigs.k8s.io/controller-runtime v0.8.0
	sigs.k8s.io/controller-tools v0.4.0
)
