goxc
====

[goxc](http://www.laher.net.nz/goxc) is a build tool for Go, with a focus on cross-compiling and packaging.

By default, goxc [g]zips (& .debs for Linux) the programs, and generates a 'downloads page' in markdown (with a Jekyll header). Goxc also provides integration with [bintray.com](https://bintray.com) for simple uploads.

goxc is written in Go but uses *os.exec* to call 'go build' with the appropriate flags & env variables for each supported platform.

goxc was inspired by Dave Cheney's Bash script [golang-crosscompile](https://github.com/davecheney/golang-crosscompile).
BUT, goxc crosscompiles to all platforms at once. The artifacts are saved into a directory structure along with a markdown file of relative links.

Thanks to [dchest](https://github.com/dchest) for the tidy-up and adding the zip feature, and [matrixik](https://bitbucket.org/matrixik) for his improvements and input.

**NOTE: Version 0.5.0 and 0.6.0 have significant changes, which may affect you in certain cases - especially if you are using config files.**
**Please read about these changes [0.5](https://github.com/laher/goxc/wiki/upgrading-0.5) and [0.6](https://github.com/laher/goxc/issues/7)**

Notable Features
----------------
 * (re-)building toolchain to all or selected platforms.
 * Cross-compilation, to all supported platforms.
 * filtering on target platform (via commandline options or config file)
 * Zip (or tar.gz) archiving of cross-compiled artifacts
 * Packaging into .debs (for Debian/Ubuntu Linux)
 * Bundling of READMEs etc into archives
 * bintray.com integration (deploys binaries to bintray.com. bintray.com registration required.)
 * 'downloads page' generation (markdown format).
 * Configuration files for repeatable builds. Includes support for multiple configurations per-project.
 * Per-task configuration options.
 * Override files for 'local' working-copy-specific (or branch-specific) configurations.
 * Config file generation & upgrade (using -wc option).
 * go-test, go-vet, go-fmt, go-install, go-clean tasks.
 * version number interpolation during build/test/... (uses go's -ldflags compiler option)

Installation
--------------
goxc requires the go source and the go toolchain.

 1. [Install go from source](http://golang.org/doc/install/source). (Requires gcc (or MinGW) and 'hg')

 2. Install goxc:

            go get github.com/laher/goxc

Basic Usage
-----------

### Run once

To build the toolchains for all platforms:

       goxc -t

### Now build your artifacts

To build [g]zipped binaries and .debs for your app:

	cd path/to/app/dir
	goxc


More options
------------

Use `goxc -h options` to list all options.

 * e.g. To restrict by OS and Architecture:

	goxc -os=linux -arch="amd64 arm"

 * e.g. To set a destination root directory and artifact version number:

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
 * For more info on a particular taks, run `goxc -h <taskname>`. This will also show you the configuration options available for that task.

Outcome
-------

By default, artifacts are generated and then immediately archived into (outputdir).

e.g.1 /my/outputdir/0.1.1/linux_arm/myapp_0.1.1_linux_arm.tar.gz
e.g.2 /my/outputdir/0.1.1/windows_386/myapp_0.1.1_windows_386.zip
e.g.3 /my/outputdir/0.1.1/linux_386/myapp_0.1.1_linux_386.deb

The version number is specified with -pv=0.1.1 .

By default, the output directory is ($GOBIN)/(appname)-xc, and the version is 'unknown', but you can specify these.

e.g.

      goxc -pv=0.1.1 -d=/home/me/myuser-github-pages/myapp/downloads/

*NOTE: it's **bad idea** to use project-level github-pages - your repo will become huge. User-level gh-pages are an option, but it's better to use the 'bintray' tasks.*:

If non-archived, artifacts generated into a directory structure as follows:

 (outputdir)/(version)/(OS)_(ARCH)/(appname)(.exe?)

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
