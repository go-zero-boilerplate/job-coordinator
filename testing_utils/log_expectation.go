package testing_utils

import (
	"fmt"
	"strings"
)

type LogExpectation struct {
	LineCount int
	Lines     []ExpectedLine
}

func (l *LogExpectation) MeetsExpectation(testingLogger *TestingLogger, includeDebugLines bool) error {
	lines := testingLogger.Handler.Lines
	if !includeDebugLines {
		filteredLines := []*testingLogLine{}
		for _, line := range lines {
			if !line.IsDebug {
				filteredLines = append(filteredLines, line)
			}
		}
		lines = filteredLines
	}

	if len(lines) != l.LineCount {
		return fmt.Errorf(
			"TestingLogger does not have expected line count of %d (count was %d), lines (including debug lines = %t) were: %s",
			l.LineCount,
			len(lines),
			includeDebugLines,
			strings.Join(testingLogger.FullStrings(includeDebugLines), "\n"))
	}

	for _, el := range l.Lines {
		line := lines[el.Index]

		if el.Debug && !line.IsDebug {
			return fmt.Errorf("TestingLogger line Index=%d failed expectation to be Debug level (D=%t,I=%t,W=%t,E=%t)", el.Index, line.IsDebug, line.IsInfo, line.IsWarning, line.IsError)
		}
		if el.Info && !line.IsInfo {
			return fmt.Errorf("TestingLogger line Index=%d failed expectation to be Info level (D=%t,I=%t,W=%t,E=%t)", el.Index, line.IsDebug, line.IsInfo, line.IsWarning, line.IsError)
		}
		if el.Warning && !line.IsWarning {
			return fmt.Errorf("TestingLogger line Index=%d failed expectation to be Warning level (D=%t,I=%t,W=%t,E=%t)", el.Index, line.IsDebug, line.IsInfo, line.IsWarning, line.IsError)
		}
		if el.Error && !line.IsError {
			return fmt.Errorf("TestingLogger line Index=%d failed expectation to be Error level (D=%t,I=%t,W=%t,E=%t)", el.Index, line.IsDebug, line.IsInfo, line.IsWarning, line.IsError)
		}

		for _, subStr := range el.RequiredSubstrings {
			lineText := line.String()
			if !strings.Contains(lineText, subStr) {
				return fmt.Errorf("TestingLogger line Index=%d did not contain sub string '%s'. The actual line content was: '%s'", el.Index, subStr, lineText)
			}
		}
	}
	return nil
}

type ExpectedLine struct {
	Index              int
	RequiredSubstrings []string
	Error              bool
	Warning            bool
	Info               bool
	Debug              bool
}
