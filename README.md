goxc
====

[goxc](http://www.laher.net.nz/goxc) cross-compiles Go programs to (up to) 11 target platforms at once.

goxc is written in Go but uses *os.exec* to call 'go build' with the appropriate flags & env variables for each supported platform.

goxc was inspired by Dave Cheney's Bash script [golang-crosscompile](https://github.com/davecheney/golang-crosscompile).
BUT, goxc crosscompiles to all platforms at once. The artifacts are saved into a folder structure along with a markdown file of relative links.

Thanks to [dchest](https://github.com/dchest) for the tidy-up and adding the zip feature, and [matrixik](https://bitbucket.org/matrixik) for his improvements and input.

Installation
--------------
goxc requires the go source and the go toolchain.

 1. [Install go from source](http://golang.org/doc/install/source). (Requires gcc (or MinGW) and 'hg')

 2. Install goxc:

            go get github.com/laher/goxc

Basic Usage
-----------

### Run once

To build the toolchains for all 11 platforms:

       goxc -t

### Now build your artifacts

To build zipped binaries for your app:

	goxc path/to/app/folder

OR

	cd path/to/app/folder
	goxc


More options
------------

Use `goxc -h options` to list all options.

 * e.g. To restrict by OS and Architecture:

	goxc -os=linux -arch="amd64 arm"

 * e.g. To set a destination root folder and artifact version number:

	goxc -d=my/jekyll/site/downloads -pv=0.1.1

 * 'Package version' can be compiled into your app if you define a VERSION variable in your main package.

"Tasks"
-------

goxc performs a number of operations, defined as tasks. You can specify tasks with the '-tasks' option.

 * `goxt -t` performs one task called 'toolchain'. It's the equivalent of `gox -tasks=toolchain`
 * goxc performs several tasks, which can be summarised as follows:
    * validate (several tasks) -> cross-compile (one task called 'xc') -> package (several tasks)
 * For a list of tasks, run `goxc -h tasks`

Outcome
-------

By default, artifacts are generated and then immediately zipped into (outputfolder).

e.g. /my/outputfolder/0.1.1/linux_arm/myapp_0.1.1_linux_arm.zip

If you specified the version number -pv=123 then the filename would be myapp_0.1.1_linux_arm_123.zip.

By default, the output folder is ($GOBIN)/(appname)-xc, and the version is 'unknown', but you can specify these.

e.g.

      goxc -pv=0.1.1 -d=/home/me/myapp/ghpages/downloads/


If non-zipped, artifacts generated into a folder structure as follows:

 (outputfolder)/(version)/(OS)_(ARCH)/(appname)(.exe?)

Configuration file
-----------------

For repeatable builds (and some extra options), it is recomended to use goxc can use a configuration file to save and re-run compilations.

To create a config file, just use the -wc (write config) option.

	goxc -d=../site/downloads -os="linux windows" -wc

The configuration file is documented in much more detail in [the wiki](https://github.com/laher/goxc/wiki/config)

Settings file format
--------------------

TODO!!!

The settings file exposes plenty more options which are not available via command line options...

For now, just specify lots of options including -wc to see the output. Use -c=test to produce a test.json file.

Limitations
-----------

 * Tested on Linux, Windows (and Mac during an early version). Please test on Mac and *BSD
 * Currently goxc is only designed to build standalone Go apps without linked libraries. You can try but YMMV
 * Bug: issue with config overriding. Empty strings do not currently override non-empty strings. e.g. -pi="" doesnt override the json setting PackageInfo
 * The *API* is not considered stable yet, so please don't start embedding goxc method calls in your code yet - unless you 'Contact us' first! Then I can freeze some API details as required.

License
-------

   Copyright 2013 Am Laher

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

See also
--------
 * [TODOs](https://github.com/laher/goxc/wiki/todo)
 * [Changelog](https://github.com/laher/goxc/wiki/changelog)
 * See also [my golang-crosscompile fork](https://github.com/laher/golang-crosscompile) for an added 'go-build-all' task similar to goxc.
