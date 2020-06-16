<h1 align="center"><img src="assets/logo.svg" height="50" alt="ets" /></h1>

<p align="center">
  <a href="https://github.com/zmwangx/ets/releases"><img src="https://img.shields.io/github/v/release/zmwangx/ets" alt="GitHub release" /></a>
  <a href="https://github.com/zmwangx/ets/actions"><img src="https://github.com/zmwangx/ets/workflows/test/badge.svg?branch=master" alt="Build status" /></a>
</p>

<p align="center"><img src="assets/animation.svg" alt="ets" /></p>

`ets` is a command output timestamper — it prefixes each line of a command's output with a timestamp.

The purpose of `ets` is similar to that of moreutils [`ts(1)`](https://manpages.ubuntu.com/manpages/focal/en/man1/ts.1.html), but `ets` differentiates itself from similar offerings by running commands directly within ptys, hence solving thorny issues like pipe buffering and commands disabling color and interactive features when detecting a pipe as output. (`ets` also provides a reading-from-stdin mode if you insist.) A more detailed comparison of `ets` and `ts` can be found [below](#comparison-to-moreutils-ts).

`ets` currently supports macOS, Linux, and various other *ix variants.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Examples](#examples)
- [Installation](#installation)
- [Usage](#usage)
- [Comparison to moreutils ts](#comparison-to-moreutils-ts)
- [License](#license)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Examples

Run a command with `ets`:

```console
$ ets ping localhost
[2020-06-16 17:13:03] PING localhost (127.0.0.1): 56 data bytes
[2020-06-16 17:13:03] 64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.036 ms
[2020-06-16 17:13:04] 64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.077 ms
[2020-06-16 17:13:05] 64 bytes from 127.0.0.1: icmp_seq=2 ttl=64 time=0.037 ms
...
```

Run a shell command:

```console
$ ets 'ping localhost | grep icmp'
[2020-06-16 17:13:03] 64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.036 ms
[2020-06-16 17:13:04] 64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.077 ms
[2020-06-16 17:13:05] 64 bytes from 127.0.0.1: icmp_seq=2 ttl=64 time=0.037 ms
...
```

Pipe command output into stdin:

```console
$ ping localhost | grep icmp | ets
[2020-06-16 17:13:03] 64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.036 ms
[2020-06-16 17:13:04] 64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.077 ms
[2020-06-16 17:13:05] 64 bytes from 127.0.0.1: icmp_seq=2 ttl=64 time=0.037 ms
...
```

Show elapsed time:

```console
$ ets -s ping -i5 localhost
[00:00:00] PING localhost (127.0.0.1): 56 data bytes
[00:00:00] 64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.039 ms
[00:00:05] 64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.031 ms
[00:00:10] 64 bytes from 127.0.0.1: icmp_seq=2 ttl=64 time=0.030 ms
[00:00:15] 64 bytes from 127.0.0.1: icmp_seq=3 ttl=64 time=0.045 ms
...
```

Show incremental time (since last timestamp):

```console
$ ets -i ping -i5 localhost
[00:00:00] PING localhost (127.0.0.1): 56 data bytes
[00:00:00] 64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.043 ms
[00:00:05] 64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.047 ms
[00:00:05] 64 bytes from 127.0.0.1: icmp_seq=2 ttl=64 time=0.116 ms
[00:00:05] 64 bytes from 127.0.0.1: icmp_seq=3 ttl=64 time=0.071 ms
...
```

Use a different timestamp format:

```console
$ ets -f '%b %d %T|' ping localhost
Jun 16 17:13:03| PING localhost (127.0.0.1): 56 data bytes
Jun 16 17:13:03| 64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.036 ms
Jun 16 17:13:04| 64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.077 ms
Jun 16 17:13:05| 64 bytes from 127.0.0.1: icmp_seq=2 ttl=64 time=0.037 ms
...
```

Millisecond precision (microsecond available too):

```console
$ ets -s -f '[%T.%L]' ping -i 0.1 localhost
[00:00:00.004] PING localhost (127.0.0.1): 56 data bytes
[00:00:00.004] 64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.032 ms
[00:00:00.108] 64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.038 ms
[00:00:00.209] 64 bytes from 127.0.0.1: icmp_seq=2 ttl=64 time=0.051 ms
[00:00:00.311] 64 bytes from 127.0.0.1: icmp_seq=3 ttl=64 time=0.049 ms
...
```

Use a different timezone:

```console
$ ets ping localhost  # UTC
[2020-06-16 09:13:03] PING localhost (127.0.0.1): 56 data bytes
[2020-06-16 09:13:03] 64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.036 ms
[2020-06-16 09:13:04] 64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.077 ms
[2020-06-16 09:13:05] 64 bytes from 127.0.0.1: icmp_seq=2 ttl=64 time=0.037 ms
```

```console
$ ets -z America/Los_Angeles -f '[%F %T%z]' ping localhost
[2020-06-16 02:13:03-0700] PING localhost (127.0.0.1): 56 data bytes
[2020-06-16 02:13:03-0700] 64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.036 ms
[2020-06-16 02:13:04-0700] 64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.077 ms
[2020-06-16 02:13:05-0700] 64 bytes from 127.0.0.1: icmp_seq=2 ttl=64 time=0.037 ms
```

## Installation

TBD.

## Usage

<!-- begin manpage -->

```

ETS(1)                    BSD General Commands Manual                   ETS(1)

NAME
     ets -- command output timestamper

SYNOPSIS
     ets [-s | -i] [-f format] [-u | -z timezone] command [arg ...]
     ets [options] shell_command
     ets [options]

DESCRIPTION
     ets prefixes each line of a command's output with a timestamp.

     The three forms in SYNOPSIS correspond to three command execution modes:

     o If given a single command without whitespace(s), or a command and its
       arguments, execute the command with exec in a pty;

     o If given a single command with whitespace(s), the command is treated as
       a shell command and executed as `SHELL -c shell_command', where SHELL
       is the current user's login shell, or sh if login shell cannot be
       determined;

     o If given no command, output is read from stdin, and the user is respon-
       sible for piping in a command's output.

     There are three mutually exclusive timestamp modes:

     o The default is absolute time mode, where timestamps from the wall clock
       are shown;

     o -s, --elapsed turns on elapsed time mode, where every timestamp is the
       time elapsed from the start of the command (using a monotonic clock);

     o -i, --incremental turns on incremental time mode, where every timestamp
       is the time elapsed since the last timestamp (using a monotonic clock).

     The default format of the prefixed timestamps depends on the timestamp
     mode active. Users may supply a custom format string with the -f,
     --format option.

     The timezone for absolute timestamps can be controlled via the -u, --utc
     and -z, --timezone options. Local time is used by default.

     The full list of options:

     -s, --elapsed
              Run in elapsed time mode.

     -i, --incremental
              Run in incremental time mode.

     -f, --format format
              Use custom strftime(3)-style format string format for prefixed
              timestamps.

              The default is ``[%Y-%m-%d %H:%M:%S]'' for absolute time mode
              and ``[%H:%M:%S]'' for elapsed and incremental time modes.

              See FORMATTING DIRECTIVES for details.

     -u, --utc
              Use UTC for absolute timestamps instead of local time.

              This option is mutually exclusive with --z, --timezone.

     -z, --timezone timezone
              Use timezone for absolute timestamps instead of local time.
              timezone is an IANA time zone name, e.g.
              ``America/Los_Angeles''.

              This option is mutually exclusive with -u, --utc.

FORMATTING DIRECTIVES
     Formatting directives largely match strftime(3)'s directives on FreeBSD
     and macOS, with the following differences:

     o Additional directives %f for microsecond and %L for millisecond are
       supported.

     o POSIX locale extensions %E* and %O* are not supported;

     o glibc extensions %-*, %_*, and %0* are not supported;

     o Directives %G, %g, and %+ are not supported.

     Below is the full list of supported directives:

     %A    is replaced by national representation of the full weekday name.

     %a    is replaced by national representation of the abbreviated weekday
           name.

     %B    is replaced by national representation of the full month name.

     %b    is replaced by national representation of the abbreviated month
           name.

     %C    is replaced by (year / 100) as decimal number; single digits are
           preceded by a zero.

     %c    is replaced by national representation of time and date.

     %D    is equivalent to ``%m/%d/%y''.

     %d    is replaced by the day of the month as a decimal number (01-31).

     %e    is replaced by the day of the month as a decimal number (1-31);
           single digits are preceded by a blank.

     %F    is equivalent to ``%Y-%m-%d''.

     %f    is replaced by the microsecond as a decimal number (000000-999999).

     %H    is replaced by the hour (24-hour clock) as a decimal number
           (00-23).

     %h    the same as %b.

     %I    is replaced by the hour (12-hour clock) as a decimal number
           (01-12).

     %j    is replaced by the day of the year as a decimal number (001-366).

     %k    is replaced by the hour (24-hour clock) as a decimal number (0-23);
           single digits are preceded by a blank.

     %L    is replaced by the millisecond as a decimal number (000-999).

     %l    is replaced by the hour (12-hour clock) as a decimal number (1-12);
           single digits are preceded by a blank.

     %M    is replaced by the minute as a decimal number (00-59).

     %m    is replaced by the month as a decimal number (01-12).

     %n    is replaced by a newline.

     %p    is replaced by national representation of either "ante meridiem"
           (a.m.)  or "post meridiem" (p.m.)  as appropriate.

     %R    is equivalent to ``%H:%M''.

     %r    is equivalent to ``%I:%M:%S %p''.

     %S    is replaced by the second as a decimal number (00-60).

     %s    is replaced by the number of seconds since the Epoch, UTC (see
           mktime(3)).

     %T    is equivalent to ``%H:%M:%S''.

     %t    is replaced by a tab.

     %U    is replaced by the week number of the year (Sunday as the first day
           of the week) as a decimal number (00-53).

     %u    is replaced by the weekday (Monday as the first day of the week) as
           a decimal number (1-7).

     %V    is replaced by the week number of the year (Monday as the first day
           of the week) as a decimal number (01-53).  If the week containing
           January 1 has four or more days in the new year, then it is week 1;
           otherwise it is the last week of the previous year, and the next
           week is week 1.

     %v    is equivalent to ``%e-%b-%Y''.

     %W    is replaced by the week number of the year (Monday as the first day
           of the week) as a decimal number (00-53).

     %w    is replaced by the weekday (Sunday as the first day of the week) as
           a decimal number (0-6).

     %X    is replaced by national representation of the time.

     %x    is replaced by national representation of the date.

     %Y    is replaced by the year with century as a decimal number.

     %y    is replaced by the year without century as a decimal number
           (00-99).

     %Z    is replaced by the time zone name.

     %z    is replaced by the time zone offset from UTC; a leading plus sign
           stands for east of UTC, a minus sign for west of UTC, hours and
           minutes follow with two digits each and no delimiter between them
           (common form for RFC 822 date headers).

     %%    is replaced by `%'.

SEE ALSO
     ts(1), strftime(3)

HISTORY
     The name ets comes from ``enhanced ts'', referring to moreutils ts(1).

AUTHORS
     Zhiming Wang <i@zhimingwang.org>

                                 June 16, 2020
```

<!-- end manpage -->

## Comparison to moreutils ts

Advantages:

- Runs commands in ptys, making ets mostly transparent and avoiding pipe-related issues like buffering and lost coloring and interactivity.
- Has better operating defaults (uses monotonic clock where appropriate) and better formatting defaults (subjective).
- Supports alternative time zones.
- Is written in Go, not Perl, so you install a single executable, not script plus modules.
- Has an executable name that doesn't conflict with other known packages. moreutils as a whole is a conflicting hell, and ts alone conflicts with at least task-spooler.

Disadvantages:

- Needs an additional `-f` for format string, because ets reserves positional arguments for its core competency. Hopefully offset by better default.
- Does not support the `-r` mode of ts. It's a largely unrelated mode of operation and I couldn't even get `ts -r` to work anywhere, maybe because optional dependencies aren't satisfied, or maybe I misunderstood the feature altogether. Anyway, not interested.
- Supports fewer formatting directives. Let me know if this is actually an issue, it could be fixable.

## License

Copyright © 2020 Zhiming Wang <i@zhimingwang.org>

The project is distributed under [the MIT license](https://opensource.org/licenses/MIT).

Special thanks to DinosoftLab on None Project for the [hourglass icon](https://thenounproject.com/term/hourglass/1674538/) used in the logo, and [termtosvg](https://github.com/nbedos/termtosvg) for the animated terminal recording.
