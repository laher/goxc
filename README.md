goxc
====

[goxc](http://www.laher.net.nz/goxc) cross-compiles Go programs to (up to) 9 target platforms at once.

goxc is written in Go but uses *os.exec* to call 'go build' with the appropriate flags & env variables for each supported platform.

goxc was inspired by Dave Cheney's Bash script [golang-crosscompile](https://github.com/davecheney/golang-crosscompile).
BUT, goxc crosscompiles to all platforms at once. The artifacts are saved into a folder structure along with a markdown file of relative links.

Thanks to [dchest](https://github.com/dchest) for the tidy-up and adding the zip feature, and [matrixik](https://bitbucket.org/matrixik) for his improvements.

Installation
--------------
goxc requires the go source and the go toolchain.

 1. [Install go from source](http://golang.org/doc/install/source). (Requires 'hg' and gcc (or MinGW))

 2. Install goxc:

            go get github.com/laher/goxc

Basic Usage
-----------

### Run once:

To build the toolchains for all 9 platforms:

       goxc -t

### Now build your artifacts

To build zipped binaries for your app:

       goxc path/to/app/folder

### Going further

See 'goxc -h' for more options.

e.g. To restrict by OS and Architecture:

      goxc -os=windows -arch=amd64 .

e.g. To set a destination root folder and artifact version number:

      goxc -d=my/jekyll/site/downloads -v=0.1.1 .

e.g. To output non-zipped binaries into folders:

      goxc -z=false .

Outcome
-------

By default, artifacts are generated and then immediately zipped into (outputfolder).

e.g. /my/outputfolder/latest/myapp_linux_arm.zip

If you specified the version number -av=123 then the filename would be myapp_linux_arm_123.zip.

By default, the output folder is ($GOBIN)/(appname)-xc, and the version is 'latest', but you can specify these.

e.g.

      goxc -av=0.1 -d=/home/me/myapp/ghpages/downloads/ .


If non-zipped, artifacts generated into a folder structure as follows:

 (outputfolder)/(version)/(OS)_(ARCH)/(appname)(.exe?)

Limitations
-----------

 * Only tested on Linux & Windows. Please test on Mac/FreeBSD
 * Currently goxc is only designed to build standalone Go apps without linked libraries. You can try but YMMV

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

Todo
----

 * Manifest file for setting up default settings. Preferably manifest.json/goxc.json, similar to Chrome extensions or npm packages.
 * "Copy resources" option for INSTALL, READMEs, LICENSE, configs etc
 * Much more error handling and potentially stop-on-error=true flag
 * Unit Tests!!
 * Configurable 'downloads' page: name, format (e.g. markdown/html/ReST) & header/footer.
 * Artifact types: tgz, maybe packaging too (.deb/.rpm/.pkg...).
 * Automatically build target toolchain if missing? (need to work out detection mechanism)
 * 'download golang source' option?
 * building .so/.dll shared libraries?
 * One day: Build plugins (for OS-specific wizardry and stuff)? Pre- and post-processing scripts?

See also
--------
See also [my golang-crosscompile fork](https://github.com/laher/golang-crosscompile) for an added 'build-all' task similar to goxc.
