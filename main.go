package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/mattn/go-runewidth"
	"github.com/riywo/loginshell"
	flag "github.com/spf13/pflag"
)

var version = "unknown"

// Regexp to strip ANSI escape sequences from string. Credit:
// https://github.com/chalk/ansi-regex/blob/2b56fb0c7a07108e5b54241e8faec160d393aedb/index.js#L4-L7
// https://github.com/acarl005/stripansi/blob/5a71ef0e047df0427e87a79f27009029921f1f9b/stripansi.go#L7
var ansiEscapes = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")

func printStreamWithTimestamper(r io.Reader, timestamper *Timestamper, delim string) {
	scanner := bufio.NewScanner(r)
	// Split on \r\n|\r|\n, and return the line as well as the line ending (\r
	// or \n is preserved, \r\n is collapsed to \n). Adaptation of
	// bufio.ScanLines.
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		lfpos := bytes.IndexByte(data, '\n')
		crpos := bytes.IndexByte(data, '\r')
		if crpos >= 0 {
			if lfpos < 0 || lfpos > crpos+1 {
				// We have a CR-terminated "line".
				return crpos + 1, data[0 : crpos+1], nil
			}
			if lfpos == crpos+1 {
				// We have a CRLF-terminated line.
				return lfpos + 1, append(data[0:crpos], '\n'), nil
			}
		}
		if lfpos >= 0 {
			// We have a LF-terminated line.
			return lfpos + 1, data[0 : lfpos+1], nil
		}
		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	})
	for scanner.Scan() {
		fmt.Print(timestamper.CurrentTimestampString(), delim, scanner.Text())
	}
}

func runCommandWithTimestamper(args []string, timestamper *Timestamper, delim string) error {
	// Calculate optimal pty size, taking into account horizontal space taken up by timestamps.
	getPtyWinsize := func() *pty.Winsize {
		winsize, err := pty.GetsizeFull(os.Stdin)
		if err != nil {
			// Most likely stdin isn't a tty, in which case we don't care.
			return winsize
		}
		totalCols := winsize.Cols
		plainTimestampString := ansiEscapes.ReplaceAllString(timestamper.CurrentTimestampString(), "")
		// Timestamp width along with one space character.
		occupiedWidth := uint16(runewidth.StringWidth(plainTimestampString)) + 1
		var effectiveCols uint16 = 0
		if occupiedWidth < totalCols {
			effectiveCols = totalCols - occupiedWidth
		}
		winsize.Cols = effectiveCols
		// Best effort estimate of the effective width in pixels.
		if totalCols > 0 {
			winsize.X = winsize.X * effectiveCols / totalCols
		}
		return winsize
	}

	command := exec.Command(args[0], args[1:]...)
	ptmx, err := pty.StartWithSize(command, getPtyWinsize())
	if err != nil {
		return err
	}
	defer func() { _ = ptmx.Close() }()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigs {
			switch sig {
			case syscall.SIGWINCH:
				if err := pty.Setsize(ptmx, getPtyWinsize()); err != nil {
					log.Println("error resizing pty:", err)
				}

			case syscall.SIGINT:
				_ = syscall.Kill(-command.Process.Pid, syscall.SIGINT)

			case syscall.SIGTERM:
				_ = syscall.Kill(-command.Process.Pid, syscall.SIGTERM)

			default:
			}
		}
	}()
	sigs <- syscall.SIGWINCH

	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()

	printStreamWithTimestamper(ptmx, timestamper, delim)

	return command.Wait()
}

func main() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	var elapsedMode = flag.BoolP("elapsed", "s", false, "show elapsed timestamps")
	var incrementalMode = flag.BoolP("incremental", "i", false, "show incremental timestamps")
	var format = flag.StringP("format", "f", "", "show timestamps in this format")
	var delim = flag.StringP("delim", "d", " ", "delimiter after timestamp (default is space)")
	var utc = flag.BoolP("utc", "u", false, "show absolute timestamps in UTC")
	var timezoneName = flag.StringP("timezone", "z", "", "show absolute timestamps in this timezone, e.g. America/New_York")
	var color = flag.BoolP("color", "c", false, "show timestamps in color")
	var printHelp = flag.BoolP("help", "h", false, "print help and exit")
	var printVersion = flag.BoolP("version", "v", false, "print version and exit")
	flag.CommandLine.SortFlags = false
	flag.SetInterspersed(false)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
ets -- command output timestamper

ets prefixes each line of a command's output with a timestamp. Lines are
delimited by CR, LF, or CRLF.

Usage:

  %s [-s | -i] [-f format] [-d delim] [-u | -z timezone] command [arg ...]
  %s [options] shell_command
  %s [options]

The three usage strings correspond to three command execution modes:

* If given a single command without whitespace(s), or a command and its
  arguments, execute the command with exec in a pty;

* If given a single command with whitespace(s), the command is treated as
  a shell command and executed as SHELL -c shell_command, where SHELL is
  the current user's login shell, or sh if login shell cannot be determined;

* If given no command, output is read from stdin, and the user is
  responsible for piping in a command's output.

There are three mutually exclusive timestamp modes:

* The default is absolute time mode, where timestamps from the wall clock
  are shown;

* -s, --elapsed turns on elapsed time mode, where every timestamp is the
  time elapsed from the start of the command (using a monotonic clock);

* -i, --incremental turns on incremental time mode, where every timestamp is
  the time elapsed since the last timestamp (using a monotonic clock).

The default format of the prefixed timestamps depends on the timestamp mode
active. Users may supply a custom format string with the -f, --format option.
The format string is basically a strftime(3) format string; see the man page
or README for details on supported formatting directives.

The timezone for absolute timestamps can be controlled via the -u, --utc
and -z, --timezone options. --timezone accepts IANA time zone names, e.g.,
America/Los_Angeles. Local time is used by default.

Options:
`, os.Args[0], os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *printHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	mode := AbsoluteTimeMode
	if *elapsedMode && *incrementalMode {
		log.Fatal("conflicting flags --elapsed and --incremental")
	}
	if *elapsedMode {
		mode = ElapsedTimeMode
	}
	if *incrementalMode {
		mode = IncrementalTimeMode
	}
	if *format == "" {
		if mode == AbsoluteTimeMode {
			*format = "[%F %T]"
		} else {
			*format = "[%T]"
		}
	}
	s, err := strconv.Unquote(`"` + *delim + `"`)
	if err != nil {
		log.Fatalf("error parsing delimiter string: %s", err)
	} else {
		*delim = s
	}
	timezone := time.Local
	if *utc && *timezoneName != "" {
		log.Fatal("conflicting flags --utc and --timezone")
	}
	if *utc {
		timezone = time.UTC
	}
	if *timezoneName != "" {
		location, err := time.LoadLocation(*timezoneName)
		if err != nil {
			log.Fatal(err)
		}
		timezone = location
	}
	if *color {
		*format = "\x1b[32m" + *format + "\x1b[0m"
	}
	args := flag.Args()

	timestamper, err := NewTimestamper(*format, mode, timezone)
	if err != nil {
		log.Fatal(err)
	}

	exitCode := 0
	if len(args) == 0 {
		printStreamWithTimestamper(os.Stdin, timestamper, *delim)
	} else {
		if len(args) == 1 {
			arg0 := args[0]
			if matched, _ := regexp.MatchString(`\s`, arg0); matched {
				shell, err := loginshell.Shell()
				if err != nil {
					shell = "sh"
				}
				args = []string{shell, "-c", arg0}
			}
		}
		if err = runCommandWithTimestamper(args, timestamper, *delim); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				log.Fatal(err)
			}
		}
	}
	os.Exit(exitCode)
}
