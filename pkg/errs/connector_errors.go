package errs

import "fmt"

type ConnectorError struct {
	FunctionName string
	Err          error
}

func (e *ConnectorError) Error() string {
	return fmt.Sprintf("ConnectorError: [%s] %s", e.FunctionName, e.Err)
}

func (e *ConnectorError) Unwrap() error {
	return e.Err
}

func NewConnectorError(name string, err error) error {
	return &ConnectorError{FunctionName: name, Err: err}
}

type VersionMismatchError struct {
	Key             string
	ExpectedVersion string
	CurrentVersion  string
	Err             error
}

func (e *VersionMismatchError) Error() string {
	return fmt.Sprintf("VersionMismatchError: Key[%s], ExpectedVersion[%s], CurrentVersion[%s], Err[%s]", e.Key, e.ExpectedVersion, e.CurrentVersion, e.Err)
}

func (e *VersionMismatchError) Unwrap() error {
	return e.Err
}

func NewVersionMismatchError(key, expectedVersion, currentVersion string, err error) error {
	return &VersionMismatchError{Key: key, ExpectedVersion: expectedVersion, CurrentVersion: currentVersion, Err: err}
}
