#!/usr/bin/env zsh

# Updated rendered manpage content in README.md.
#
# Requires GNU sed.

setopt errexit

here=$0:A:h
root=$here:h

tmpfile=$(mktemp)
echo '\n```' >>$tmpfile
man $root/ets.1 | sed -r 's/.\x08//g' >>$tmpfile
echo '```\n' >>$tmpfile

sed -i "/<!-- begin manpage -->/,/<!-- end manpage -->/ {
  /<!-- begin manpage -->/ {
    r $tmpfile
    n
  }
  /<!-- end manpage -->/!d
}" $root/README.md
