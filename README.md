goxc
====

Cross compiler for Go, written in Go but still using *os.exec()* to call 'go build'.

This is inspired by Dave Cheney's Bash script [golang-crosscompile](https://github.com/davecheney/golang-crosscompile).
BUT, goxc crosscompiles to all platforms at once. The artifacts are saved into a folder structure.


Pre-requisites
--------------

At this stage you still need to:

 * Download the go source code and set up the GOROOT accordingly.
 * Use golang-crosscompile (above) to build the toolchains.

Basic Usage
-----------
Simply run:
 go gocx.go .
Or using the binary:
 gocx .

Outcome
-------
Artifacts stored in:
 $GOBIN/appname/<os>/<arch>/latest/appname(.exe?)

Todo
----

 * 'build toolchains' option.
 * 'generate Downloads page' option.
 * 'download golang source' option?

See also
--------

See [my golang-crosscompile fork](https://github.com/laher/golang-crosscompile) for an added 'build-all' task similar to goxc.
