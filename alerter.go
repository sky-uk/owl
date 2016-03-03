package main

import (
	"bytes"
	"fmt"
	"github.com/coreos/go-systemd/sdjournal"
	"github.com/scalingdata/gcfg"
	"strings"
	"time"
	"os"
)

type Default struct {
	Time int
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
	err := gcfg.ReadFileInto(&config, "/etc/service_log_checks.gcfg")
	if err != nil {
		fmt.Printf("Unable to read configuration file: %v", err)
		return
	}

	var logLoader = JournalCtrl{duration: config.Default.Time}
	errors := CheckErrors(config, logLoader)
	fmt.Println("There have been ", len(errors), " errors in ", config.Default.Time, " seconds.")
	for _, element := range errors {
		fmt.Println(element)
	}
	if len(errors) > 0 {
		os.Exit(-1)
	}
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
		Since:   time.Duration(-this.duration) * time.Second,
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
