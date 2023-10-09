package logger

import (
	"fmt"
	"path/filepath"
	"testing"

	"os"

	logging "github.com/op/go-logging"
)

func TestGetLoggerShouldGetTheLoggerForGivenModule(t *testing.T) {
	Initialize(false, logging.INFO, API, "_testdata")

	l := LoggersMap.GetLogger("api-js")
	if l == nil {
		t.Error("Expected a logger to be initilized for api-js")
	}
}

func TestLoggerInitWithInfoLevel(t *testing.T) {
	Initialize(false, logging.INFO, API, "_testdata")

	if !LoggersMap.GetLogger(moduleID).IsEnabledFor(logging.INFO) {
		t.Error("Expected Log to be enabled for INFO")
	}
}

func TestLoggerInitWithDefaultLevel(t *testing.T) {
	Initialize(false, logging.INFO, API, "_testdata")

	if !LoggersMap.GetLogger(moduleID).IsEnabledFor(logging.INFO) {
		t.Error("Expected Log to be enabled for default log level")
	}
}

func TestLoggerInitWithDebugLevel(t *testing.T) {
	Initialize(false, logging.DEBUG, API, "_testdata")

	if !LoggersMap.GetLogger(moduleID).IsEnabledFor(logging.DEBUG) {
		t.Error("Expected Log to be enabled for DEBUG")
	}
}

func TestLoggerInitWithWarningLevel(t *testing.T) {
	Initialize(false, logging.WARNING, API, "_testdata")

	if !LoggersMap.GetLogger(moduleID).IsEnabledFor(logging.WARNING) {
		t.Error("Expected Log to be enabled for WARNING")
	}
}

func TestLoggerInitWithErrorLevel(t *testing.T) {
	Initialize(false, logging.ERROR, API, "_testdata")

	if !LoggersMap.GetLogger(moduleID).IsEnabledFor(logging.ERROR) {
		t.Error("Expected Log to be enabled for ERROR")
	}
}

func TestGetLogFileGivenRelativePathInProject(t *testing.T) {
	projectRoot, _ := filepath.Abs("_testdata")
	want := filepath.Join(projectRoot, string(API))

	got := getLogFile(API, "_testdata")
	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func TestGetLogFileWhenLogsDirNotSet(t *testing.T) {
	projectRoot, _ := filepath.Abs("_testdata")
	want := filepath.Join(projectRoot, string(API))

	got := getLogFile(API, "_testdata")
	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func TestGetLogFileInProjectWhenRelativeCustomLogsDirIsSet(t *testing.T) {
	myLogsDir := "my_logs"
	os.Setenv(logsDirectory, myLogsDir)
	defer os.Unsetenv(logsDirectory)

	projectRoot, _ := filepath.Abs("_testdata")
	want := filepath.Join(projectRoot, myLogsDir, string(API))

	got := getLogFile(API, "_testdata/"+myLogsDir)

	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func TestGetLogFileInProjectWhenAbsoluteCustomLogsDirIsSet(t *testing.T) {
	myLogsDir, err := filepath.Abs("my_logs")
	if err != nil {
		t.Errorf("Unable to convert to absolute path, %s", err.Error())
	}

	_ = os.Setenv(logsDirectory, myLogsDir)
	defer os.Unsetenv(logsDirectory)

	want := filepath.Join(myLogsDir, string(API))

	got := getLogFile(API, "my_logs")

	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func TestGetErrorText(t *testing.T) {
	expectedText := fmt.Sprintf("An Error has Occurred: %s", "some error")
	fatalErrors = append(fatalErrors, expectedText)
	got := GetFatalErrorMsg()
	want := expectedText

	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
	fatalErrors = []string{}
}

func TestToJSONWithPlainText(t *testing.T) {
	outMessage := &OutMessage{MessageType: "out", Message: "plain text"}
	want := "{\"type\":\"out\",\"message\":\"plain text\"}"

	got, _ := outMessage.ToJSON()
	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func TestToJSONWithInvalidJSONCharacters(t *testing.T) {
	outMessage := &OutMessage{MessageType: "out", Message: "\n, \t, and \\ needs to be escaped to create a valid JSON"}
	want := "{\"type\":\"out\",\"message\":\"\\n, \\t, and \\\\ needs to be escaped to create a valid JSON\"}"

	got, _ := outMessage.ToJSON()
	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}
