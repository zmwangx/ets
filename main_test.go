package main_test

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/creack/pty"
)

var rootdir string
var tempdir string
var executable string

func init() {
	_, currentFile, _, _ := runtime.Caller(0)
	rootdir = path.Dir(currentFile)
}

func compile(moduledir string, output string) {
	cmd := exec.Command("go", "build", "-o", output)
	cmd.Dir = moduledir
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to compile %s: %s", moduledir, err)
	}
}

func TestMain(m *testing.M) {
	var retcode int
	var err error

	defer func() { os.Exit(retcode) }()

	tempdir, err = ioutil.TempDir("", "*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempdir)

	executable = path.Join(tempdir, "ets")

	// Build ets and test fixtures to tempdir.
	compile(rootdir, executable)
	fixturesdir := path.Join(rootdir, "fixtures")
	content, err := ioutil.ReadDir(fixturesdir)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range content {
		if entry.IsDir() {
			name := entry.Name()
			compile(path.Join(fixturesdir, name), path.Join(tempdir, name))
		}
	}

	err = os.Chdir(tempdir)
	if err != nil {
		log.Fatal(err)
	}

	retcode = m.Run()
}

type parsedLine struct {
	raw      string
	prefix   string
	output   string
	captures map[string]string
}

func parseOutput(output []byte, prefixPattern string) []*parsedLine {
	linePattern := regexp.MustCompile(`^(?P<prefix>` + prefixPattern + `) (?P<output>.*)$`)
	lines := strings.Split(string(output), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1] // Drop final empty line.
	}
	parsed := make([]*parsedLine, 0)
	for _, line := range lines {
		// Drop final CR if there is one.
		if line != "" && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		m := linePattern.FindStringSubmatch(line)
		if m == nil {
			parsed = append(parsed, &parsedLine{
				raw:      line,
				prefix:   "",
				output:   "",
				captures: nil,
			})
		} else {
			captures := make(map[string]string)
			for i, name := range linePattern.SubexpNames() {
				if i != 0 && name != "" {
					captures[name] = m[i]
				}
			}
			parsed = append(parsed, &parsedLine{
				raw:      line,
				prefix:   captures["prefix"],
				output:   captures["output"],
				captures: captures,
			})
		}
	}
	return parsed
}

func TestBasic(t *testing.T) {
	defaultOutputs := []string{"out1", "err1", "out2", "err2", "out3", "err3"}
	tests := []struct {
		name            string
		args            []string
		prefixPattern   string
		expectedOutputs []string
	}{
		{
			"basic",
			[]string{"./basic"},
			`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]`,
			defaultOutputs,
		},
		{
			"basic-format",
			[]string{"-f", "%m/%d/%y %T:", "./basic"},
			`\d{2}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}:`,
			defaultOutputs,
		},
		{
			"basic-elapsed",
			[]string{"-s", "./basic"},
			`\[00:00:00\]`,
			defaultOutputs,
		},
		{
			"basic-elapsed-format",
			[]string{"-s", "-f", "%T.%f", "./basic"},
			`00:00:00\.\d{6}`,
			defaultOutputs,
		},
		{
			"basic-incremental",
			[]string{"-i", "./basic"},
			`\[00:00:00\]`,
			defaultOutputs,
		},
		{
			"basic-incremental-format",
			[]string{"-i", "-f", "%T.%f", "./basic"},
			`00:00:00\.\d{6}`,
			defaultOutputs,
		},
		{
			"basic-utc-format",
			[]string{"-u", "-f", "[%F %T%z]", "./basic"},
			`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\+0000\]`,
			defaultOutputs,
		},
		{
			"basic-timezone-format",
			[]string{"-z", "America/Los_Angeles", "-f", "[%F %T %Z]", "./basic"},
			`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} P[DS]T\]`,
			defaultOutputs,
		},
		{
			"basic-shell",
			[]string{"./basic 2>/dev/null | nl -w1 -s' '"},
			`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]`,
			[]string{"1 out1", "2 out2", "3 out3"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := exec.Command("./ets", test.args...)
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("command failed: %s", err)
			}
			parsed := parseOutput(output, test.prefixPattern)
			outputs := make([]string, 0)
			for _, pl := range parsed {
				if pl.prefix == "" {
					t.Errorf("unexpected line: %s", pl.raw)
				}
				outputs = append(outputs, pl.output)
			}
			if !reflect.DeepEqual(outputs, test.expectedOutputs) {
				t.Fatalf("wrong outputs: expected %#v, got %#v", test.expectedOutputs, outputs)
			}
		})
	}
}

func TestCR(t *testing.T) {
	cmd := exec.Command("./ets", "-f", "[timestamp]", "echo '1\r2'")
	expectedOutput := "[timestamp] 1\r[timestamp] 2\n"
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("command failed: %s", err)
	}
	if string(output) != expectedOutput {
		t.Fatalf("wrong output: expected %#v, got %#v", expectedOutput, string(output))
	}
}

func TestStdin(t *testing.T) {
	input := "out1\nout2\nout3\n"
	expectedOutputs := []string{"out1", "out2", "out3"}
	cmd := exec.Command("./ets")
	stdin, _ := cmd.StdinPipe()
	go func() {
		defer stdin.Close()
		_, _ = stdin.Write([]byte(input))
	}()
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("command failed: %s", err)
	}
	parsed := parseOutput(output, `\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]`)
	outputs := make([]string, 0)
	for _, pl := range parsed {
		if pl.prefix == "" {
			t.Errorf("unexpected line: %s", pl.raw)
		}
		outputs = append(outputs, pl.output)
	}
	if !reflect.DeepEqual(outputs, expectedOutputs) {
		t.Fatalf("wrong outputs: expected %#v, got %#v", expectedOutputs, outputs)
	}
}

func TestElapsedMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow test in short mode")
	}
	expectedOutput := "[1] out1\n[2] out2\n[3] out3\n"
	cmd := exec.Command("./ets", "-s", "-f", "[%s]", "./timed")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("command failed: %s", err)
	}
	if string(output) != expectedOutput {
		t.Fatalf("wrong output: expected %#v, got %#v", expectedOutput, string(output))
	}
}

func TestIncrementalMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow test in short mode")
	}
	expectedOutput := "[1] out1\n[1] out2\n[1] out3\n"
	cmd := exec.Command("./ets", "-i", "-f", "[%s]", "./timed")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("command failed: %s", err)
	}
	if string(output) != expectedOutput {
		t.Fatalf("wrong output: expected %#v, got %#v", expectedOutput, string(output))
	}
}

func TestExitCode(t *testing.T) {
	for code := 1; code < 6; code++ {
		t.Run("exitcode-"+strconv.Itoa(code), func(t *testing.T) {
			cmd := exec.Command("./ets", "./basic", "-exitcode", strconv.Itoa(code))
			err := cmd.Run()
			errExit, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatalf("expected ExitError, got %#v", err)
			}
			if errExit.ExitCode() != code {
				t.Fatalf("expected exit code %d, got %d", code, errExit.ExitCode())
			}
		})
	}
}

func TestSignals(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow test in short mode")
	}
	cmd := exec.Command("./ets", "./signals")
	go func() {
		time.Sleep(time.Second)
		_ = cmd.Process.Signal(syscall.SIGINT)
		time.Sleep(time.Second)
		_ = cmd.Process.Signal(syscall.SIGTERM)
	}()
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("command failed: %s", err)
	}
	parsed := parseOutput(output, `\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]`)
	outputs := make([]string, 0)
	for _, pl := range parsed {
		if pl.prefix == "" {
			t.Errorf("unexpected line: %s", pl.raw)
		}
		outputs = append(outputs, pl.output)
	}
	for _, expectedOutput := range []string{
		"busy waiting",
		"ignored SIGINT",
		"shutting down after receiving SIGTERM",
	} {
		found := false
		for _, output := range outputs {
			if output == expectedOutput {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected output %#v not found in outputs %#v", expectedOutput, outputs)
		}
	}
}

func TestWindowSize(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		prefixPattern  string
		rows           uint16
		cols           uint16
		expectedOutput string
	}{
		{
			"default",
			[]string{"./winsize"},
			`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]`,
			24,
			80,
			"58x24",
		},
		{
			"color",
			[]string{"-f", "\x1b[32m[%Y-%m-%d %H:%M:%S]\x1b[0m", "./winsize"},
			`\x1b\[32m\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]\x1b\[0m`,
			24,
			80,
			"58x24",
		},
		{
			"wide-chars",
			[]string{"-f", "[时间 %Y-%m-%d %H:%M:%S]", "./winsize"},
			`\[时间 \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]`,
			24,
			80,
			"53x24",
		},
		{
			"narrow-terminal",
			[]string{"./winsize"},
			`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]`,
			24,
			10,
			"0x24",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expectedOutputs := []string{test.expectedOutput}
			cmd := exec.Command("./ets", test.args...)
			ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: test.rows, Cols: test.cols, X: 0, Y: 0})
			if err != nil {
				t.Fatalf("failed to start command in pty: %s", err)
			}
			defer func() { _ = ptmx.Close() }()
			output, err := ioutil.ReadAll(ptmx)
			// TODO: figure out why we get &os.PathError{Op:"read", Path:"/dev/ptmx", Err:0x5} on Linux.
			// https://github.com/creack/pty/issues/100
			if len(output) == 0 && err != nil {
				t.Fatalf("failed to read pty output: %s", err)
			}
			parsed := parseOutput(output, test.prefixPattern)
			outputs := make([]string, 0)
			for _, pl := range parsed {
				if pl.prefix == "" {
					t.Errorf("unexpected line: %s", pl.raw)
				}
				outputs = append(outputs, pl.output)
			}
			if !reflect.DeepEqual(outputs, expectedOutputs) {
				t.Fatalf("wrong outputs: expected %#v, got %#v", expectedOutputs, outputs)
			}
		})
	}
}
