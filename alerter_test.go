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
	config := Config{Default{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").Return("error1\nerror2")

	errors := CheckErrors(config, journalCtrl)

	assert.Equal(t, []string{"error1", "error2"}, errors)
}

func TestCheckErrorsIncludeFilter(t *testing.T) {
	var services = map[string]*Service{
		"kube-apiserver": &Service{Include: []string{": E"}},
	}
	config := Config{Default{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").Return(": E error1\nerror2")

	errors := CheckErrors(config, journalCtrl)

	assert.Equal(t, []string{": E error1"}, errors)
}

func TestCheckErrorsExcludeFilter(t *testing.T) {
	var services = map[string]*Service{
		"kube-apiserver": &Service{Include: []string{": E"}, Exclude: []string{"stupid error"}},
	}
	config := Config{Default{Time: 5}, services}
	journalCtrl := new(MockJournalCtl)
	journalCtrl.On("Logs", "kube-apiserver").Return(": E error1\nerror2\n: E stupid error")

	errors := CheckErrors(config, journalCtrl)

	assert.Equal(t, []string{": E error1"}, errors)
}
