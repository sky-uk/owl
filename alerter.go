package main

import (
	"bytes"
	"fmt"
	"github.com/coreos/go-systemd/sdjournal"
	"github.com/scalingdata/gcfg"
	"os"
	"strings"
	"time"
)

type Default struct {
	Time int
	ErrorsToReport int
	FatalNumberOfErrors int
}

type Service struct {
	Include []string
	Exclude []string
}

type Config struct {
	Default Default
	Service map[string]*Service
}

func main() {
	var config Config
	err := gcfg.ReadFileInto(&config, "/etc/owl.gcfg")
	if err != nil {
		fmt.Printf("Unable to read configuration file: %v", err)
		return
	}

	var logLoader = JournalCtrl{duration: config.Default.Time}
	errors := CheckErrors(config, logLoader)
	report := ReportErrors(config, errors)
	for _, element := range report {
		fmt.Println(element)
	}
	if len(errors) > config.Default.FatalNumberOfErrors {
		os.Exit(-1)
	}
}

func ReportErrors(config Config, errors []string) []string {
	size := config.Default.ErrorsToReport + 1
	report := make([]string, 0, size)
	serviceNames := make([]string, 0, len(config.Service))
	for serviceName, _ := range config.Service {
		serviceNames = append(serviceNames, serviceName)
	}
	report = append(report, fmt.Sprintf("There have been %v errors in the last %v seconds for services: %v", len(errors), config.Default.Time, serviceNames))
	errorsToReport := config.Default.ErrorsToReport
	if len(errors) < errorsToReport {
		errorsToReport = len(errors)
	}
	for _, element := range errors[0:errorsToReport] {
		report = append(report, element)
	}
	return report
}

func CheckErrors(config Config, logLoader LogLoader) []string {
	errors := make([]string, 0, 0)
	for k, v := range config.Service {
		logs, _ := logLoader.Logs(k)
		splitLogs := strings.Split(logs, "\n")
		for _, logMessage := range splitLogs {
			if included(logMessage, v.Include) && !excluded(logMessage, v.Exclude) {
				errors = append(errors, logMessage)
			}
		}
	}
	return errors
}

func included(logMessage string, includes []string) bool {
	for _, include := range includes {
		if !strings.Contains(logMessage, include) {
			return false
		}
	}
	return true
}

func excluded(logMessage string, excludes []string) bool {
	for _, exclude := range excludes {
		if strings.Contains(logMessage, exclude) {
			return true
		}
	}
	return false
}

type LogLoader interface {
	Logs(unit string) (string, error)
}

type JournalCtrl struct {
	duration int
}

func (this JournalCtrl) Logs(unit string) (string, error) {
	timeout := time.Duration(1) * time.Second
	r, err := sdjournal.NewJournalReader(sdjournal.JournalReaderConfig{
		Since: time.Duration(-this.duration) * time.Second,
		Matches: []sdjournal.Match{
			{
				Field: sdjournal.SD_JOURNAL_FIELD_SYSTEMD_UNIT,
				Value: unit,
			},
		},
	})
	if err != nil {
		fmt.Println("Failed to read journal logs ", err)
		return "", err
	}
	buf := new(bytes.Buffer)
	if err = r.Follow(time.After(timeout), buf); err != sdjournal.ErrExpired {
		fmt.Println("Failed to follow logs from journalctl", err)
		return "", err
	}
	return buf.String(), nil
}
