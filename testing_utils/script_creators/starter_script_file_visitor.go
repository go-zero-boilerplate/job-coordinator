package script_creators

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const INSTALL_DUMMYCLI_HELP = "dummycli can be installed with `go get -u github.com/golang-devops/dummycli`"

func NewCreateStarterScriptFileVisitor(dummyCliFlagInterval time.Duration, dummyCliFlagNumber int, failureMode bool) *CreateStarterScriptFileVisitor {
	return &CreateStarterScriptFileVisitor{
		dummyCliFlagInterval: dummyCliFlagInterval,
		dummyCliFlagNumber:   dummyCliFlagNumber,
		failureMode:          failureMode,
	}
}

type CreateStarterScriptFileVisitor struct {
	dummyCliFlagInterval time.Duration
	dummyCliFlagNumber   int
	failureMode          bool

	FileNameOnly     string
	DummyCliBinPath  string //See INSTALL_DUMMYCLI_HELP above for help text how to install dummycli
	ExpectedLogLines []string

	echoText string
	Content  string
}

func (c *CreateStarterScriptFileVisitor) setCommonFields() {
	c.echoText = "Hallo there"

	if c.dummyCliFlagNumber <= 0 {
		panic("Unexpected value for 'dummyCliFlagNumber'")
	}
	c.ExpectedLogLines = []string{}
	for i := 0; i < c.dummyCliFlagNumber; i++ {
		c.ExpectedLogLines = append(c.ExpectedLogLines, fmt.Sprintf("Dummy %d/%d (every %s)", i+1, c.dummyCliFlagNumber, c.dummyCliFlagInterval.String()))
	}
	c.ExpectedLogLines = append(c.ExpectedLogLines, "Command exited with code 0")
	c.ExpectedLogLines = append(c.ExpectedLogLines, c.echoText)
}

func (c *CreateStarterScriptFileVisitor) getDummyCliFlagsCombined() string {
	args := []string{
		"-interval",
		c.dummyCliFlagInterval.String(),
		"-number",
		fmt.Sprintf("%d", c.dummyCliFlagNumber),
	}
	if c.failureMode {
		args = append(args, "-non-existing-mock-failure-flag")
	}
	return strings.Join(args, " ")
}

func (c *CreateStarterScriptFileVisitor) VisitWindows() {
	c.setCommonFields()
	c.FileNameOnly = "script.bat"
	c.DummyCliBinPath = os.ExpandEnv(`$GOPATH/bin/dummycli.exe`)
	c.Content = strings.Join([]string{
		"@echo off",
		"set errorlevel=",
		fmt.Sprintf(`"%s" %s`, c.DummyCliBinPath, c.getDummyCliFlagsCombined()), //fmt.Sprintf(`"%s" %s & if errorlevel 1 EXIT /b %%errorlevel%%`, c.DummyCliBinPath, c.getDummyCliFlagsCombined()),
		fmt.Sprintf(`if errorlevel 1 exit /b %%errorlevel%%`),
		fmt.Sprintf(`echo %s`, c.echoText),
	}, "\r\n")
}

func (c *CreateStarterScriptFileVisitor) VisitLinux() {
	//TODO: Run the starter_test on linux/darwin
	c.setCommonFields()
	c.FileNameOnly = "script.sh"
	c.DummyCliBinPath = os.ExpandEnv(`$GOPATH/bin/dummycli`)
	c.Content = strings.Join([]string{
		fmt.Sprintf(`"%s" %s`, c.DummyCliBinPath, c.getDummyCliFlagsCombined()),
		fmt.Sprintf(`rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi`),
		fmt.Sprintf(`echo %s`, c.echoText),
	}, "\n")
}

func (c *CreateStarterScriptFileVisitor) VisitDarwin() {
	c.VisitLinux()
}
