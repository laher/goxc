package tasks

/*
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
*/

import (
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/archive"
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"path/filepath"
)

var rmBinTask = Task{
	"rmbin",
	"delete binary. Normally runs after 'archive' task to reduce size of output folder.",
	runTaskRmBin}

//runs automatically
func init() {
	register(rmBinTask)
}

func runTaskRmBin(tp taskParams) error {
	for _, platformArr := range tp.destPlatforms {
		destOs := platformArr[0]
		destArch := platformArr[1]
		rmBinPlat(destOs, destArch, tp.appName, tp.outDestRoot, tp.settings)
	}
	//TODO return error
	return nil
}

func rmBinPlat(goos, arch, appName, outDestRoot string, settings config.Settings) {
	relativeBin := core.GetRelativeBin(goos, arch, appName, false, settings.GetFullVersionName())
	archive.RemoveArchivedBinary(filepath.Join(outDestRoot, relativeBin))
}
