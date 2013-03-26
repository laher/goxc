goxc
====

[goxc](http://www.laher.net.nz/goxc) cross-compiles Go programs to (up to) 11 target platforms at once.

By default, goxc [g]zips up the programs and generates a 'downloads page' in markdown (with a Jekyll header).

goxc is written in Go but uses *os.exec* to call 'go build' with the appropriate flags & env variables for each supported platform.

goxc was inspired by Dave Cheney's Bash script [golang-crosscompile](https://github.com/davecheney/golang-crosscompile).
BUT, goxc crosscompiles to all platforms at once. The artifacts are saved into a folder structure along with a markdown file of relative links.

Thanks to [dchest](https://github.com/dchest) for the tidy-up and adding the zip feature, and [matrixik](https://bitbucket.org/matrixik) for his improvements and input.

**NOTE: Version 0.5.0 (2013-03-24) is a major change. If you are using config files, please [read about the changes](https://github.com/laher/goxc/wiki/upgrading-0.5) for more information!**

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

To build [g]zipped binaries for your app:

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

goxc performs a number of operations, defined as tasks. You can specify tasks with the '-tasks=' option.

 * `goxc -t` performs one task called 'toolchain'. It's the equivalent of `goxc -tasks=toolchain ~`
 * The *default* task is actually several tasks, which can be summarised as follows:
    * validate (tests the code) -> compile (cross-compiles code) -> package ([g]zips up the executables and builds a 'downloads' page)
 * You can specify one or more tasks, such as `goxc -tasks=go-fmt,xc`
 * You can skip tasks with '-tasks-='. Skip the 'package' stage with `goxc -tasks-=package`
 * For a list of tasks and 'task aliases', run `goxc -h tasks`

### Available tasks

 * toolchain       Build toolchain. Make sure to run this each time you update go source.
 * clean-destination  Delete the output folder for this version of the artifact.
 * go-clean        runs `go clean`.
 * go-fmt          runs `go fmt ./...`.
 * go-test         runs `go test ./...`. (folder is configurable).
 * go-vet          runs `go vet ./...`.
 * go-install      runs `go install`. installs a version consistent with goxc-built binaries.
 * xc              Cross compile. Builds executables for other platforms.
 * codesign        sign code for Mac. Only Mac hosts are supported for this task.
 * archive         Create a compressed archive. Currently 'zip' format is used for all platforms except Linux (tar.gz)
 * rmbin           delete binary. Normally runs after 'archive' task to reduce size of output folder.
 * downloads-page  Generate a downloads page. Currently only supports Markdown format.

### Task aliases

Task aliases are a name given to a sequence of tasks. You can specify tasks or aliases interchangeably.
Specify aliases wherever possible.

 * all             [toolchain go-fmt go-vet go-test go-install xc codesign archive rmbin downloads-page]
 * package         [archive rmbin downloads-page]
 * default         [go-vet go-test go-install xc codesign archive rmbin downloads-page]
 * validate        [go-vet go-test]
 * compile         [go-install xc codesign]
 * clean           [go-clean clean-destination]

### NEW TASK in 0.5.3

There's a new task to generate .debs (Debian/Ubuntu installers).

This is very nascent so I expect bugs to crop up, and config variables to change in the near future.

Eventually this will be included into the default workflow.

For now, to generate debs, please check [goxc's own config file](https://github.com/laher/goxc/blob/master/.goxc.json) for config parameters, then use the following tasks list:

	goxc -tasks=validate,compile,pkg-build,package

Alternatively, run your normal tasks excluding 'rmbin', then call pkg-build individually.

	goxc -tasks-=rmbin
	goxc -tasks=pkg-build

Outcome
-------

By default, artifacts are generated and then immediately archived into (outputfolder).

e.g.1 /my/outputfolder/0.1.1/linux_arm/myapp_0.1.1_linux_arm.tar.gz
e.g.2 /my/outputfolder/0.1.1/windows_386/myapp_0.1.1_windows_386.zip

If you specified the version number -pv=123 then the filename would be myapp_0.1.1_linux_arm_123.tar.gz.

By default, the output folder is ($GOBIN)/(appname)-xc, and the version is 'unknown', but you can specify these.

e.g.

      goxc -pv=0.1.1 -d=/home/me/myapp/ghpages/downloads/


If non-archived, artifacts generated into a folder structure as follows:

 (outputfolder)/(version)/(OS)_(ARCH)/(appname)(.exe?)

Configuration file
-----------------

For repeatable builds (and some extra options), it is recomended to use goxc can use a configuration file to save and re-run compilations.

To create a config file, just use the -wc (write config) option.

	goxc -d=../site/downloads -os="linux windows" -wc

The configuration file is documented in much more detail in [the wiki](https://github.com/laher/goxc/wiki/config)

Limitations
-----------

 * Tested on Linux, Windows (and Mac during an early version). Please test on Mac and *BSD
 * Currently goxc is only designed to build standalone Go apps without linked libraries. You can try but YMMV
 * The *API* is not considered stable yet, so please don't start embedding goxc method calls in your code yet - unless you 'Contact us' first! Then I can freeze some API details as required.
 * Bug: issue with config overriding. Empty strings do not currently override non-empty strings. e.g. -pi="" doesnt override the json setting PackageInfo

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
 * [Package Versioning](https://github.com/laher/goxc/wiki/versioning)
 * [Upgrading to v0.5](https://github.com/laher/goxc/wiki/upgrading-0.5)
 * See also [my golang-crosscompile fork](https://github.com/laher/golang-crosscompile) for an added 'go-build-all' task similar to goxc.
