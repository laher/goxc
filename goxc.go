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

const VERSION="0.0.1"
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
   osy string
   arch string
)
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
   log.Printf("'make.bash' working folder: %s",cmd.Dir)

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

func XCPlat(goos string ,arch string, call []string) {
   log.Printf("running for platform %s/%s.", goos, arch)

   gopath:= os.Getenv("GOPATH")
   gobin:= os.Getenv("GOBIN")
   gohostos:= os.Getenv("GOHOSTOS")
   if gobin == "" {
      //log.Print("No $GOBIN. Using $GOPATH/bin")
      gobin= gopath+"/bin"
   }
   myfoldername,err:= filepath.Abs(call[0])
   if err != nil {
     log.Printf("Error: %v",err)
   }
   myfoldername= filepath.Base(myfoldername)
   mybin:= gobin+"/"+myfoldername+"/"+goos+"/"+arch+"/latest"
   os.MkdirAll(mybin, 0755)

   cmd := exec.Command("go")
   cmd.Args= append(cmd.Args,"build")
   cmd.Args= append(cmd.Args,"-o")
   var ending = ""
   if goos == WINDOWS {
      ending= ".exe"
   }
   cmd.Args= append(cmd.Args,mybin+"/"+myfoldername+ending)
   cmd.Args= append(cmd.Args,call[0])

   cmd.Env= os.Environ()

   cgo_enabled := func()string{if goos == gohostos {return "1"}
   return "0"}()

   //host OS/arch
   cmd.Env= append(cmd.Env,"GOOS="+goos)
   cmd.Env= append(cmd.Env,"CGO_ENABLED="+cgo_enabled)
   cmd.Env= append(cmd.Env,"GOARCH="+arch)
   cmd.Env= append(cmd.Env,"GOBIN="+gobin)
   log.Printf("'go' env: GOOS=%s, CGO_ENABLED=%s, GOARCH=%s, GOBIN=%s", goos, cgo_enabled, arch, gobin)
   log.Printf("'go' args: %s",cmd.Args)
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
func help_text() {
   fmt.Fprint(os.Stderr,"`goxc` [options] <foldername>\n")
   fmt.Fprintf(os.Stderr," Version %s. Options:\n", VERSION)
   flagSet.PrintDefaults()
   // NO longer needed!
   //fmt.Fprint(os.Stderr,"Tip 2: mkdo doesn't mask password prompts. Beware!\n")
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
   for _,v := range PLATFORMS {
      if osy == "" || v[0] == osy {
         if arch == "" || v[1] == arch {
            if is_buildtoolchain {
               BuildToolchain(v[0],v[1])
            } else {
               XCPlat(v[0],v[1], remainder)
            }
         }
      }
   }
   return 0,nil
}


//main

func main() {
   log.SetPrefix("[goxc] ")
   flagSet.StringVar(&osy, "os", "", "Specify OS")
   flagSet.StringVar(&arch, "arch", "", "specify Arch (386/x64/..)")
   flagSet.BoolVar(&is_buildtoolchain, "t", false, "Build cross-compiler toolchain(s)")
   flagSet.BoolVar(&is_help, "h", false, "Show this help")
   flagSet.BoolVar(&is_version, "version", false, "version info")
   flagSet.BoolVar(&verbose, "v", false, "verbose")
   GOXC(os.Args)
}

