package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/coreos/go-systemd/sdjournal"
	"github.com/scalingdata/gcfg"
	"os"
	"regexp"
	"strings"
	"time"
)

type Global struct {
	Time           int
	ErrorsToReport int
	AlertThreshold int
}

type Service struct {
	Include []string
	Exclude []string
}

type Config struct {
	Global  Global
	Service map[string]*Service
}

var configFile string
var verbose bool

func init() {
	const (
		defaultConfig = "/etc/owl/config"
	)
	flag.StringVar(&configFile, "config", defaultConfig, "owl's configuration file")
	flag.BoolVar(&verbose, "verbose", false, "verbose logging")
}

func main() {

	flag.Parse()

	var config Config
	err := gcfg.ReadFileInto(&config, configFile)
	if err != nil {
		fmt.Printf("Unable to read configuration file: %v\n", err)
		flag.Usage()
		os.Exit(-1)
	}

	logConfig(config)

	var logLoader = JournalCtrl{duration: config.Global.Time}
	errors := CheckErrors(config, logLoader)
	report := ReportErrors(config, errors)
	for _, element := range report {
		fmt.Println(element)
	}
	if len(errors) > config.Global.AlertThreshold {
		os.Exit(-1)
	}
}

func logConfig(config Config) {
	fmt.Printf("Starting owl with lookback=%vs, lastErrors=%v, alertThreshold=%v\n",
		config.Global.Time, config.Global.ErrorsToReport, config.Global.AlertThreshold)

	for name, service := range config.Service {
		fmt.Printf("Watching [%v]\n", name)
		if verbose {
			fmt.Printf("Include:\n  %v", strings.Join(service.Include, "\n  "))
			fmt.Println()
			fmt.Printf("Exclude:\n  %v", strings.Join(service.Exclude, "\n  "))
			fmt.Println()
		}
	}
}

func CheckErrors(config Config, logLoader LogLoader) []string {
	errors := make([]string, 0, 0)
	for serviceName, filters := range config.Service {
		logs, err := logLoader.Logs(serviceName)
		if err != nil {
			fmt.Println("Unable to get logs for service %v, error %v", serviceName, err)
			continue
		}
		splitLogs := strings.Split(logs, "\n")
		for _, logMessage := range splitLogs {
			if included(logMessage, filters.Include) && !excluded(logMessage, filters.Exclude) {
				errors = append(errors, logMessage)
			}
		}
	}
	return errors
}

func ReportErrors(config Config, errors []string) []string {
	size := config.Global.ErrorsToReport + 1
	report := make([]string, 0, size)
	serviceNames := make([]string, 0, len(config.Service))
	for serviceName := range config.Service {
		serviceNames = append(serviceNames, serviceName)
	}
	report = append(report, fmt.Sprintf("There have been %v errors in the last %v seconds for services: %v", len(errors), config.Global.Time, serviceNames))
	errorsToReport := config.Global.ErrorsToReport
	if len(errors) < errorsToReport {
		errorsToReport = len(errors)
	}
	for _, element := range errors[0:errorsToReport] {
		report = append(report, element)
	}
	return report
}

func included(logMessage string, includes []string) bool {
	if len(includes) == 0 {
		return true
	}

	for _, include := range includes {
		matched, err := regexp.MatchString(include, logMessage)
		if err != nil {
			panic(err)
		}

		if matched {
			return true
		}
	}
	return false
}

func excluded(logMessage string, excludes []string) bool {
	for _, exclude := range excludes {
		matched, err := regexp.MatchString(exclude, logMessage)
		if err != nil {
			panic(err)
		}

		if matched {
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
