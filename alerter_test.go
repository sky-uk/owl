package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockJournalCtl struct {
	mock.Mock
}

func (m MockJournalCtl) Logs(unit string) (string, error) {
	args := m.Called(unit)
	return args.String(0), nil
}

func TestCheckErrors(t *testing.T) {
	var services = map[string]*Service{
		"kube-apiserver": &Service{},
	}
	config := Config{Global{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").Return("error1\nerror2")

	errors := CheckErrors(config, journalCtrl)

	assert.Equal(t, []string{"error1", "error2"}, errors)
}

func TestCheckErrorsIncludeFilter(t *testing.T) {
	var services = map[string]*Service{
		"kube-apiserver": &Service{Include: []string{": E"}},
	}
	config := Config{Global{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").Return(": E error1\nerror2")

	errors := CheckErrors(config, journalCtrl)

	assert.Equal(t, []string{": E error1"}, errors)
}

func TestCheckErrorsExcludeFilter(t *testing.T) {
	var services = map[string]*Service{
		"kube-apiserver": &Service{Include: []string{": E"}, Exclude: []string{"stupid error"}},
	}
	config := Config{Global{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").Return(": E error1\nerror2\n: E stupid error")

	errors := CheckErrors(config, journalCtrl)

	assert.Equal(t, []string{": E error1"}, errors)
}

func TestReportErrorsLimitsErrors(t *testing.T) {
	config := Config{Global{Time: 5, ErrorsToReport: 1}, map[string]*Service{
		"kube-apiserver": &Service{Include: []string{": E"}, Exclude: []string{"stupid error"}},
	}}

	errors := []string{"error 1", "error 2", "error 3"}
	report := ReportErrors(config, errors)

	assert.Equal(t, []string{"There have been 3 errors in the last 5 seconds for services: [kube-apiserver]", "error 1"}, report)
}

func TestReportErrorsHandlesLessErrorsThanConfigured(t *testing.T) {
	config := Config{Global{Time: 5, ErrorsToReport: 5}, map[string]*Service{}}

	errors := []string{"error 1", "error 2", "error 3"}
	report := ReportErrors(config, errors)

	assert.Equal(t, []string{"There have been 3 errors in the last 5 seconds for services: []", "error 1", "error 2", "error 3"}, report)
}
