goxc
====

Cross compiler for Go, written in Go but still using *os.exec()* to call 'go build'.

This is inspired by Dave Cheney's Bash script [golang-crosscompile](https://github.com/davecheney/golang-crosscompile).

BUT, goxc crosscompiles to all platforms at once. See [my golang-crosscompile fork](https://github.com/laher/golang-crosscompile) for an added 'build-all' task similar to this one.

At this stage you still need to:

 * Download the go source code and set up the GOROOT accordingly.
 * Use golang-crosscompile (above) to build the toolchains.

Todo:

 * 'build toolchains' option.
 * 'download golang source' option?
