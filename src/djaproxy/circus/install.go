package circus

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/craigmj/aptlastupdate"
	"github.com/craigmj/commander"
)

var DefaultBackend = "waitress"

// Install installs circus and chausette.
// Install needs to be run as root.
func Install(destDir, app, backend string) error {
	var err error
	if "" == backend {
		backend = DefaultBackend
	}
	if "" == destDir {
		destDir, err = os.Getwd()
		if nil != err {
			return err
		}
	}
	if updated, _ := aptlastupdate.Within(time.Hour * 6); !updated {
		err = runCmd(exec.Command("apt-get", "update"))
		if nil != err {
			return err
		}
	}
	cmd := exec.Command("apt-get",
		"install",
		"-y",
		"python", "python-pip",
		"libzmq-dev",
		"libevent-dev",
		"python-dev",
		"python-virtualenv",
		"libmysqlclient-dev")
	if err := runCmd(cmd); nil != err {
		return err
	}
	cmd = exec.Command("virtualenv", destDir)
	if err := runCmd(cmd); nil != err {
		return err
	}
	cmd = exec.Command("bin/pip", "install",
		"circus",
		"circus-web",
		"chaussette",
		"gevent",
		"django",
		"mysql-python",
		"django-epiceditor",
		"django-grappelli",
		"django-request",
		DefaultBackend)
	cmd.Dir = destDir
	if err := runCmd(cmd); nil != err {
		return err
	}

	circusIniFilename := filepath.Join(destDir, "circus.ini")
	circusIni, err := os.Create(circusIniFilename)
	if nil != err {
		return err
	}
	defer circusIni.Close()
	if err = circusFile(circusIni, destDir, app, backend); nil != err {
		return err
	}
	return nil
}

// InstallCommand returns the Commander command for
// install.
func InstallCommand() *commander.Command {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	app := fs.String("app", "", "Name of the wsgi app we're using")
	backend := fs.String("backend", "waitress", "Backend to use for chaussette")
	destDir := fs.String("dest", "", "Install location")
	return commander.NewCommand(
		"install",
		"Installs circus, chaussette and all other requirements",
		fs,
		func(args []string) error {
			if "" == *app {
				return errors.New("You must provide an app name (-app)")
			}
			return Install(*destDir, *app, *backend)
		})
}

// runCmd runs the command sending stdout to stdout and
// stderr to stderr.
func runCmd(cmd *exec.Cmd) error {
	fmt.Println("About to execute: ", strings.Join(cmd.Args, " "))

	rStdout, err := cmd.StdoutPipe()
	if nil != err {
		return err
	}
	rStderr, err := cmd.StderrPipe()
	if nil != err {
		return err
	}
	go io.Copy(os.Stdout, rStdout)
	go io.Copy(os.Stderr, rStderr)
	return cmd.Run()
}

// circusFile writes the circus file to the given io.Writer
func circusFile(out io.Writer, dir, app, backend string) error {
	if "" == backend {
		backend = DefaultBackend
	}
	t := template.Must(template.New("").Parse(circusIniTemplate))
	return t.Execute(out, map[string]interface{}{
		"App":     app,
		"Backend": backend,
		"Dir":     dir,
	})

}

var circusIniTemplate = `
[circus]
endpoint = tcp://127.0.0.1:5555
pubsub_endpoint = tcp://127.0.0.1:5556
stats_endpoint = tcp://127.0.0.1:5557

[watcher:web]
cmd = {{.Dir}}/bin/chaussette --fd $(circus.sockets.web) --backend {{.Backend}} {{.App}}
use_sockets = True
numprocesses = 5
copy_env = True
virtualenv = {{.Dir}}

[socket:web]
host = 0.0.0.0
port = 8000
`
