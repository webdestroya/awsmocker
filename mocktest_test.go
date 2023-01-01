package awsmocker_test

import (
	"fmt"

	"github.com/webdestroya/awsmocker"
)

// A fake testing.T object used for testing the test
type TestingMock struct {
	og awsmocker.TestingT

	failed bool

	errored      bool
	logged       bool
	helperCalled bool

	errorMessages []string
	logMessages   []string

	envvars map[string]string
}

func NewTestingMock(og awsmocker.TestingT) *TestingMock {
	return &TestingMock{
		og:            og,
		envvars:       make(map[string]string),
		logMessages:   make([]string, 0),
		errorMessages: make([]string, 0),
	}
}

func (tm *TestingMock) Cleanup(f func()) {
	tm.og.Cleanup(f)
}

func (tm *TestingMock) Errorf(f string, args ...any) {
	tm.errored = true
	tm.errorMessages = append(tm.errorMessages, fmt.Sprintf(f, args...))
}

func (tm *TestingMock) Logf(f string, args ...any) {
	tm.logged = true
	tm.logMessages = append(tm.logMessages, fmt.Sprintf(f, args...))
}

func (tm *TestingMock) Fail() {
	tm.failed = true
}

func (tm *TestingMock) Setenv(k, v string) {
	tm.envvars[k] = v
	tm.og.Setenv(k, v)
}

func (tm *TestingMock) TempDir() string {
	return tm.og.TempDir()
}

func (tm *TestingMock) Helper() {
	tm.helperCalled = true
	if h, ok := tm.og.(awsmocker.THelper); ok {
		h.Helper()
	}
}

// interface adherence
var _ = (awsmocker.TestingT)(&TestingMock{})
var _ = (awsmocker.THelper)(&TestingMock{})
