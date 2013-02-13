package main

import (
//   "go/build"
   "flag"
   "fmt"
   "os"
   "os/exec"
   "log"
   "io"
   "path/filepath"
//   "github.com/laher/mkdo"
)

const VERSION="0.0.2"
const AMD64="amd64"
const X86="386"
const ARM="arm"
const DARWIN="darwin"
const LINUX="linux"
const FREEBSD="freebsd"
const WINDOWS="windows"

var PLATFORMS = [][]string {
   []string{ DARWIN, X86 },
   []string{ DARWIN, AMD64 },
   []string{ FREEBSD, X86 },
   []string{ FREEBSD, AMD64 },
   []string{ LINUX, X86 },
   []string{ LINUX, AMD64 },
   []string{ LINUX, ARM },
   []string{ WINDOWS, X86 },
   []string{ WINDOWS, AMD64 },
      }

var (
   flagSet = flag.NewFlagSet("goxc", flag.ExitOnError)
   verbose bool
   is_help bool
   is_version bool
   is_buildtoolchain bool
   aos string
   aarch string
   artifact_version string
   artifacts_destination string
)

//this function copied from 'mkdo'
func redirectIO(cmd *exec.Cmd) (*os.File, error) {
   stdout, err := cmd.StdoutPipe()
   if err != nil {
      log.Println(err)
   }
   stderr, err := cmd.StderrPipe()
   if err != nil {
      log.Println(err)
   }
   if verbose { log.Printf("Redirecting output") }
   go io.Copy(os.Stdout, stdout)
   go io.Copy(os.Stderr, stderr)
   //direct. Masked passwords work OK!
   cmd.Stdin= os.Stdin
   return nil, err
}

func BuildToolchain(goos string, arch string) {
   goroot:= os.Getenv("GOROOT")
   gohostos:= os.Getenv("GOHOSTOS")
   cmd := exec.Command(goroot+"/src/make.bash")
   cmd.Dir= goroot+"/src/"
   cmd.Args= append(cmd.Args,"--no-clean")
   cgo_enabled := func()string{if goos == gohostos {return "1"}
   return "0"}()

   cmd.Env= os.Environ()
   cmd.Env= append(cmd.Env,"GOOS="+goos)
   cmd.Env= append(cmd.Env,"CGO_ENABLED="+cgo_enabled)
   cmd.Env= append(cmd.Env,"GOARCH="+arch)
   log.Printf("'make.bash' env: GOOS=%s, CGO_ENABLED=%s, GOARCH=%s, GOROOT=%s", goos, cgo_enabled, arch, goroot)
   log.Printf("'make.bash' args: %s",cmd.Args)
   log.Printf("'make.bash' working directory: %s",cmd.Dir)

      f, err:= redirectIO(cmd)
      if err != nil {
         log.Printf("Error redirecting IO: %s",err);
      }
      if f != nil {
         defer f.Close()
      }

   err = cmd.Start()
   if err != nil {
      log.Printf("Launch error: %s",err);
     // return 1, err
   } else {
      err = cmd.Wait()
      if err != nil {
         log.Printf("Wait error: %s",err);
      } else {
         log.Printf("Complete");
      }
   }
}

func XCPlat(goos string ,arch string, call []string, is_first bool) string {
   log.Printf("building for platform %s_%s.", goos, arch)

   gopath:= os.Getenv("GOPATH")
   gohostos:= os.Getenv("GOHOSTOS")
   app_dirname,err:= filepath.Abs(call[0])
   if err != nil {
      log.Printf("Error: %v",err)
   }
   app_name:= filepath.Base(app_dirname)

   relative_dir:= artifact_version+"/"+goos+"_"+arch

   var out_destination_root string
   if artifacts_destination != "" {
      out_destination_root = artifacts_destination
   } else {
      gobin:= os.Getenv("GOBIN")
      if gobin == "" {
         gobin= gopath+"/bin"
      }
      out_destination_root = gobin + "/" + app_name + "-xc"
   }

   out_dir:= out_destination_root+"/"+relative_dir
   os.MkdirAll(out_dir, 0755)

   cmd := exec.Command("go")
   cmd.Args= append(cmd.Args,"build")
   cmd.Args= append(cmd.Args,"-o")
   var ending = ""
   if goos == WINDOWS {
      ending= ".exe"
   }
   relative_bin_for_markdown:= goos+"_"+arch+"/"+app_name+ending
   relative_bin:= relative_dir+"/"+app_name+ending
   cmd.Args= append(cmd.Args,out_destination_root+"/"+relative_bin)
   cmd.Args= append(cmd.Args,call[0])

   cmd.Env= os.Environ()

   var cgo_enabled string
   if goos == gohostos {
      cgo_enabled= "1"
   } else {
      cgo_enabled= "0"
   }

   //host OS/arch
   cmd.Env= append(cmd.Env,"GOOS="+goos)
   cmd.Env= append(cmd.Env,"CGO_ENABLED="+cgo_enabled)
   cmd.Env= append(cmd.Env,"GOARCH="+arch)
   log.Printf("'go' env: GOOS=%s, CGO_ENABLED=%s, GOARCH=%s", goos, cgo_enabled, arch)
   log.Printf("'go' args: %s",cmd.Args)
   err = cmd.Start()
   if err != nil {
      log.Printf("Launch error: %s",err);
   } else {
      err = cmd.Wait()
      if err != nil {
         log.Printf("Compiler error: %s",err);
      } else {
         log.Printf("Artifact generated OK");
         report_filename:= out_destination_root+"/"+artifact_version+"/downloads.md"
         var flags int
         if is_first {
            log.Printf("Creating %s", report_filename)
            flags= os.O_WRONLY|os.O_TRUNC|os.O_CREATE
         } else {
            flags= os.O_APPEND|os.O_WRONLY

         }
         f, err := os.OpenFile(report_filename, flags, 0600)
         if err == nil {
            defer f.Close()
            if is_first {
               _, err = fmt.Fprintf(f,"%s downloads (%s)\n------------\n\n", app_name, artifact_version)
            }
            _, err = fmt.Fprintf(f," * [%s %s](%s)\n", goos, arch, relative_bin_for_markdown)
         }
         if err != nil {
            log.Printf("Could not report to '%s': %s", report_filename, err )
         }
      }
   }
   return relative_bin
}

func help_text() {
   fmt.Fprint(os.Stderr,"`goxc` [options] <directory_name>\n")
   fmt.Fprintf(os.Stderr," Version %s. Options:\n", VERSION)
   flagSet.PrintDefaults()
}

func version_text() {
   fmt.Fprintf(os.Stderr," goxc version %s\n", VERSION)
}

func GOXC(call []string) (int,error) {
   e := flagSet.Parse(call[1:])
   if e != nil {
      return 1,e
   }
   remainder := flagSet.Args()
   if is_help  {
      help_text()
      return 0,nil
   } else if is_version {
      version_text()
      return 0,nil
   } else if is_buildtoolchain {
      //no need for remaining args
   } else if len(remainder) < 1 {
      help_text()
      return 1,nil
   }

   is_first:=true
   for _,v := range PLATFORMS {
      if aos == "" || v[0] == aos {
         if aarch == "" || v[1] == aarch {
            if is_buildtoolchain {
               BuildToolchain(v[0],v[1])
            } else {
               XCPlat(v[0],v[1], remainder, is_first)
            }
            is_first=false
         }
      }
   }
   return 0,nil
}


//main
func main() {
   log.SetPrefix("[goxc] ")
   flagSet.StringVar(&aos, "os", "", "Specify OS (linux/darwin/windows). Compiles all by default")
   flagSet.StringVar(&aarch, "arch", "", "Specify Arch (386/x64/arm). Compiles all by default")
   flagSet.StringVar(&artifact_version, "av", "latest", "Artifact version (default='latest')")
   flagSet.StringVar(&artifacts_destination, "d", "", "Destination root directory (default=$GOBIN)")
   flagSet.BoolVar(&is_buildtoolchain, "t", false, "Build cross-compiler toolchain(s)")
   flagSet.BoolVar(&is_help, "h", false, "Show this help")
   flagSet.BoolVar(&is_version, "version", false, "version info")
   flagSet.BoolVar(&verbose, "v", false, "verbose")
   GOXC(os.Args)
}

