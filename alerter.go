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
		PrintlnDebug("Include:\n  %v", strings.Join(service.Include, "\n  "))
		PrintlnDebug("Exclude:\n  %v", strings.Join(service.Exclude, "\n  "))
	}
}

func CheckErrors(config Config, logLoader LogLoader) []string {
	var errors []string
	for serviceName, filters := range config.Service {
		logs, err := logLoader.Logs(serviceName)
		if err != nil {
			fmt.Printf("Unable to get logs for service %v, error %v\n", serviceName, err)
			continue
		}

		var patternMatcher = NewPatternMatcher(filters.Include, filters.Exclude)
		matches := patternMatcher.FindAllMatch(logs)
		for _, match := range matches {
			errors = append(errors, match)
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

type PatternMatcher struct {
	includes []*regexp.Regexp
	excludes []*regexp.Regexp
}

func NewPatternMatcher(includes []string, excludes []string) *PatternMatcher {
	toRegexp := func(patterns []string) []*regexp.Regexp {
		allRegexp := make([]*regexp.Regexp, 0, len(patterns))
		for _, pattern := range patterns {
			allRegexp = append(allRegexp, regexp.MustCompile(pattern))
		}
		return allRegexp
	}
	return &PatternMatcher{includes: toRegexp(includes), excludes: toRegexp(excludes)}
}

func (matcher *PatternMatcher) FindAllMatch(s string) []string {
	matches := make([]string, 0, 0)
	for _, include := range matcher.includes {
		for _, match := range include.FindAllString(s, -1) {
			matches = append(matches, match)
		}
	}

	for _, exclude := range matcher.excludes {
		excludedCount := 0
		for i := range matches {
			j := i - excludedCount
			match := matches[j] // ensure we get the correct element taking deletes into account

			if len(exclude.FindAllString(match, -1)) != 0 {
				matches = append(matches[:j], matches[j+1:]...)
				excludedCount++
			}
		}
	}
	return matches
}

type LogLoader interface {
	Logs(unit string) (string, error)
}

type JournalCtrl struct {
	duration int
}

func generateJournalMatchConfig(unit string) []sdjournal.Match {
	var matches []sdjournal.Match
	if unit != "*" {
		matches = []sdjournal.Match{
			{
				Field: sdjournal.SD_JOURNAL_FIELD_SYSTEMD_UNIT,
				Value: unit,
			},
		}
	}
	return matches
}

func FormatJournalEntry(entry *sdjournal.JournalEntry) (string, error) {
	unitName := entry.Fields["UNIT"]
	if unitName == "" {
		unitName = entry.Fields["_SYSTEMD_UNIT"]
	}
	timestamp := time.Unix(int64(entry.RealtimeTimestamp/1000000), 0)
	return fmt.Sprintf("%s %s: %s\n", timestamp.Format(time.RFC3339),
		unitName, entry.Fields["MESSAGE"]), nil
}

func (this JournalCtrl) Logs(unit string) (string, error) {
	timeout := time.Duration(1) * time.Second
	searchPeriod := time.Duration(-this.duration) * time.Second

	r, err := sdjournal.NewJournalReader(sdjournal.JournalReaderConfig{
		Formatter: FormatJournalEntry,
		Since:     searchPeriod,
		Matches:   generateJournalMatchConfig(unit),
	})

	if err != nil {
		fmt.Println("Failed to read journal logs ", err)
		return "", err
	}

	buf := new(bytes.Buffer)
	PrintlnDebug("Following logs since: %s", time.Now().Add(searchPeriod))
	if err = r.Follow(time.After(timeout), buf); err != sdjournal.ErrExpired {
		fmt.Println("Failed to follow logs from journalctl", err)
		return "", err
	}
	return buf.String(), nil
}

func PrintlnDebug(message string, args ...interface{}) {
	if verbose {
		fmt.Printf(message, args...)
		fmt.Println()
	}
}
