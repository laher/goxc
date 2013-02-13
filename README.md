goxc
====

goxc is a cross-compilation utility for Go programs, written in Go but using *os.exec()* to call 'go build'.

goxc inspired by Dave Cheney's Bash script [golang-crosscompile](https://github.com/davecheney/golang-crosscompile).
BUT, goxc crosscompiles to all platforms at once. The artifacts are saved into a folder structure.

Pre-requisites
--------------
At this stage you still need to:

 1. Use Linux (I haven't tested on other platforms)
 2. Download the go source code and set up the GOROOT accordingly.

Basic Usage
-----------

 1. To build the toolchains for all platforms:

      goxc -t

 2. To build binaries for your app:

      cd path/to/app/folder
      goxc .

 Or use 'goxc -h' for more options.

 Note that if running from source, replace 'goxc' with 'go run goxc.go'

Outcome
-------

Artifacts generated into:

 (outputfolder)/(version)/(OS)/(ARCH)/appname(.exe)

By default, the output folder is ($GOBIN)/appname, and the version is 'latest', but you can specify these.

e.g.

      goxc -av=0.1 -d=/home/me/myapp/ghpages/downloads/ .

Limitations
-----------

 * Only tested on Linux. No idea if it would work on Win/Mac/FreeBSD
 * Currently there's a bug meaning you can only run goxc on the current folder.
 * Currently goxc is only designed to build standalone Go apps without linked libraries. You can try but YMMV

Todo
----

 * Fix bug to allow building of a different directory other than current working directory.
 * 'specify artifact folder' option
 * 'specify version name' option
 * 'generate Downloads page' option
 * 'download golang source' option?
 * building .so/.dll shared libraries?

See also
--------

See [my golang-crosscompile fork](https://github.com/laher/golang-crosscompile) for an added 'build-all' task similar to goxc.
