package storage

import (
	"net/http"
	"testing"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
)

type storageError struct{}

func (s *storageError) ServiceCode() azblob.ServiceCodeType {
	return "boom"
}

func (s *storageError) Error() string {
	return "boom"
}

func (s *storageError) Timeout() bool {
	return false
}

func (s *storageError) Temporary() bool {
	return false
}

func (s *storageError) Response() *http.Response {
	return &http.Response{
		StatusCode: http.StatusNotFound,
	}
}

var _ azblob.StorageError = &storageError{}

func TestIsNotFound(t *testing.T) {

	cases := []struct {
		err      error
		expected bool
	}{
		{nil, false},
		{errors.New("boom"), false},
		{&storageError{}, true},
	}

	for _, tt := range cases {
		got := IsNotFoundError(tt.err)
		if diff := cmp.Diff(got, tt.expected); diff != "" {
			t.Errorf("IsNotFoundError() = %v, expected %v\n%s", got, tt.expected, diff)
		}
	}
}
