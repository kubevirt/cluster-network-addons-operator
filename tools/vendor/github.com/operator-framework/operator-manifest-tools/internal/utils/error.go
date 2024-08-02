package utils

import (
	"errors"
	"fmt"
	"os"
)

var (
	// ErrNotClusterServiceVersion is the error returned with a source isn't a CSV.
	ErrNotClusterServiceVersion = errors.New("Not a ClusterServiceVersion")

	// ErrNotFound is the error returned when a file is not found
	ErrNotFound = errors.New("path not found")

	// ErrPathExpectedDifferentType is the error returned when the path expected a different type.
	ErrPathExpectedDifferentType = errors.New("path expected different type")

	ErrNoOperatorManifests = errors.New("Missing ClusterServiceVersion in operator manifests")

	ErrTooManyCSVs = errors.New("Operator bundle may contain only 1 CSV file, but contains more")

	ErrImageIsARequiredProperty = errors.New("'image' is a required property")
)

type errBase struct {
	cause error
	err   error
}

func NewError(cause error, format string, args ...interface{}) error {
	return errBase{
		err:   fmt.Errorf(format, args...),
		cause: cause,
	}
}

func (e errBase) Error() string {
	return e.err.Error()
}

func (e errBase) Unwrap() error {
	return e.cause
}

func NewErrIsNotDirectoryOrDoesNotExist(path string) error {
	return errors.New(path + " is not a directory or does not exist")
}

func CheckIfDirectoryExists(path string) error {
	stat, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			return NewErrIsNotDirectoryOrDoesNotExist(path)
		}

		return err
	}

	if !stat.IsDir() {
		return NewErrIsNotDirectoryOrDoesNotExist(path)
	}

	return nil
}

func NewErrImageDoesNotExist(imageName string, err error) error {
	return errors.New(fmt.Sprintf("Failed to inspect %s. Make sure it exists and is accessible. Cause: %s", imageName, err.Error()))
}
