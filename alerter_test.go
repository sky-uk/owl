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

func TestCheckErrorsReturnsNothingWhenNoIncludeSpecified(t *testing.T) {
	var services = map[string]*Service{
		"kube-apiserver": {},
	}
	config := Config{Global{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").Return("error1\nerror2")

	errors := CheckErrors(config, journalCtrl)

	assert.Empty(t, errors)
}

func TestCheckErrorsIncludeFilter(t *testing.T) {
	var services = map[string]*Service{
		"kube-apiserver": {Include: []string{": E.*"}},
	}
	config := Config{Global{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").Return(": E error1\nerror2")

	errors := CheckErrors(config, journalCtrl)

	assert.Equal(t, []string{": E error1"}, errors)
}

func TestCheckErrorsIncludeRegexFilter(t *testing.T) {
	var services = map[string]*Service{
		"kube-apiserver": {Include: []string{": E.*blah"}},
	}
	config := Config{Global{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").Return(": E error1 blah\n: E error2 nah")

	errors := CheckErrors(config, journalCtrl)

	assert.Equal(t, []string{": E error1 blah"}, errors)
}

func TestCheckErrorsHandlesMultipleIncludes(t *testing.T) {
	var services = map[string]*Service{
		"kube-apiserver": {Include: []string{": E.*blah", ": E.*yah"}},
	}
	config := Config{Global{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").Return(": E error1 blah\n: E error2 nah\n: E error3 yah")

	errors := CheckErrors(config, journalCtrl)

	assert.Equal(t, []string{": E error1 blah", ": E error3 yah"}, errors)
}

func TestCheckErrorsExcludeFilter(t *testing.T) {
	var services = map[string]*Service{
		"kube-apiserver": {Include: []string{": E.*"}, Exclude: []string{"stupid error"}},
	}
	config := Config{Global{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").Return(": E error1\nerror2\n: E stupid error")

	errors := CheckErrors(config, journalCtrl)

	assert.Equal(t, []string{": E error1"}, errors)
}

func TestCheckErrorsExcludeRegexFilter(t *testing.T) {
	var services = map[string]*Service{
		"kube-apiserver": {Include: []string{": E.*"}, Exclude: []string{"stu.*err"}},
	}
	config := Config{Global{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").Return(": E error1\nerror2\n: E stupid error")

	errors := CheckErrors(config, journalCtrl)

	assert.Equal(t, []string{": E error1"}, errors)
}

func TestCheckErrorsHandlesMultipleExcludes(t *testing.T) {
	var services = map[string]*Service{
		"kube-apiserver": {Include: []string{": E.*"}, Exclude: []string{"stu.*err", "dum.*err"}},
	}
	config := Config{Global{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").
		Return(": E error1 fun error\nerror2\n: E stupid error\nerror3: E dumb error")

	errors := CheckErrors(config, journalCtrl)

	assert.Equal(t, []string{": E error1 fun error"}, errors)
}

func TestCheckErrorsFindsIncludeAcrossManyLines(t *testing.T) {
	var services = map[string]*Service{
		"kube-apiserver": {Include: []string{": E.*starts\n.*ends here"}},
	}
	config := Config{Global{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").Return(": E error starts\n and ends here but can continue too")

	errors := CheckErrors(config, journalCtrl)

	assert.Equal(t, []string{": E error starts\n and ends here"}, errors)
}

func TestCheckErrorsFindsMultipleIncludeAcrossManyLines(t *testing.T) {
	var services = map[string]*Service{
		"kube-apiserver": {Include: []string{": E.*starts\n.*ends here"}},
	}

	const logLines = `: E error starts
continues and ends here
: E error starts
and ends here
something else comes up`

	config := Config{Global{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").Return(logLines)

	errors := CheckErrors(config, journalCtrl)

	assert.Equal(t, []string{": E error starts\ncontinues and ends here", ": E error starts\nand ends here"}, errors)
}

func TestCheckErrorsExcludePatternAcrossManyLines(t *testing.T) {
	var services = map[string]*Service{
		"kube-apiserver": {Include: []string{"SEVERE.*[\n]?.*"}, Exclude: []string{"SEVERE: Exception.*\nRejectedExecutionException:.*"}},
	}

	const logLines = `something bad happened
SEVERE: Exception that is excluded ...
RejectedExecutionException: ... rejected
SEVERE: Error that will be caught...`

	config := Config{Global{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").Return(logLines)

	errors := CheckErrors(config, journalCtrl)

	assert.Equal(t, []string{"SEVERE: Error that will be caught..."}, errors)
}

func TestCheckErrorsExcludePatternOccurringMultipleTimes(t *testing.T) {
	var services = map[string]*Service{
		"kube-apiserver": {Include: []string{"SEVERE.*", "ERROR.*"}, Exclude: []string{"SEVERE: Remove.*", "ERROR: Remove.*"}},
	}

	const logLines = `SEVERE: Keep
SEVERE: Remove
ERROR: Remove
ERROR: Remove
ERROR: Keep
ERROR: Remove
SEVERE: Remove`

	config := Config{Global{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").Return(logLines)

	errors := CheckErrors(config, journalCtrl)
	assert.Equal(t, []string{"SEVERE: Keep", "ERROR: Keep"}, errors)

}

func TestReportErrorsLimitsErrors(t *testing.T) {
	config := Config{Global{Time: 5, ErrorsToReport: 1}, map[string]*Service{
		"kube-apiserver": {Include: []string{": E"}, Exclude: []string{"stupid error"}},
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
