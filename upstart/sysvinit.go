package upstart

import (
	"errors"
	"flag"
	"io"
	"os"
	"text/template"

	"github.com/craigmj/commander"
)

func SysVInit(service, app, dir string, out io.Writer) error {
	var err error
	if "" == dir {
		dir, err = os.Getwd()
		if nil != err {
			return err
		}
	}
	t := template.Must(template.New("").Parse(_sysvinit))
	initScript, err := findFirstFile(
		"/lib/lsb/init-functions",
		"/etc/init.d/functions",
	)
	if nil != err {
		return err
	}
	functions := map[string]string{
		"success": "success",
		"initlog": "initlog -c",
	}
	switch initScript {
	case "/lib/lsb/init-functions":
		functions["success"] = "log_success_msg"
		functions["initlog"] = "log_progress_msg"
	case "/etc/init.d/functions":
	}
	return t.Execute(out, map[string]interface{}{
		"App":                app,
		"Dir":                dir,
		"Service":            service,
		"InitFunctionScript": initScript,
		"InitLog":            functions["initlog"],
		"Success":            functions["success"],
	})
}

func SysVInitWrite(service, app, dir string) error {
	out, err := os.Create("/etc/init.d/" + service + ".sh")
	if nil != err {
		return err
	}
	defer out.Close()
	return SysVInit(service, app, dir, out)
}

func SysVCommand() *commander.Command {
	fs := flag.NewFlagSet("sysv", flag.ExitOnError)
	app := fs.String("app", "", "Name of the app to start")
	dir := fs.String("dir", "", "Directory where manage.py is")
	service := fs.String("service", "", "Service name for the init script")
	return commander.NewCommand("sysv",
		"Write sysvinit script to /etc/init.d",
		fs,
		func(args []string) error {
			if "" == *app {
				return errors.New("You must provide an app name (-app)")
			}
			if "" == *service {
				return errors.New("You need to provide a service name (-service) for the init script")
			}
			return SysVInitWrite(*service, *app, *dir)
		})
}

var _sysvinit = `#!/bin/bash

### BEGIN INIT INFO
# Provides:          {{.Service}}
# Required-Start:    $remote_fs $syslog
# Required-Stop:     $remote_fs $syslog
# Should-Start:      
# Should-Stop:       
# X-Start-Before:    
# X-Stop-After:      
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# X-Interactive:     true
# Short-Description: {{.App}} Django wsgi and proxy
# Description:       Starts the {{.App}} Django wsgi server with 
#					 Circus and a djaproxy web proxy to serve the 
#					 full-featured app.
### END INIT INFO

#
# chkconfig: 35 90 12
# description: Sassidb server
#
# Get function from functions library
. {{.InitFunctionScript}}
# Start the service FOO
start() {
        {{.InitLog}} "Starting sassidb server: "
        cd {{.Dir}}
        . {{.Dir}}/bin/activate
        {{.Dir}}/bin/circusd --daemon circus.ini
        djaproxy -dir {{.Dir}} -app {{.App}} &
        ### Create the lock file ###
        touch /var/lock/sassidb.lck
        {{.Success}} "sassidb server startup"
        echo
}
# Restart the service FOO
stop() {
        {{.InitLog}} "Stopping sassidb server: "
        killproc djaproxy
        cd {{.Dir}}
        . {{.Dir}}/bin/activate
        {{.Dir}}/bin/circusctl quit	
        ### Now, delete the lock file ###
        rm -f /var/lock/sassidb.lck
        echo
}
### main logic ###
case "$1" in
  start)
        start
        ;;
  stop)
        stop
        ;;
  status)
        status Not implemented
        ;;
  restart|reload|condrestart)
        stop
        start
        ;;
  *)
        echo $"Usage: $0 {start|stop|restart|reload|status}"
        exit 1
esac
exit 0
`

// findFirstFile finds the first existing file in the
// given parameters, and returns that filename.
// If no files are found, os.ErrNotExist is returned
func findFirstFile(files ...string) (string, error) {
	for _, f := range files {
		h, err := os.Stat(f)
		if nil == err && !h.IsDir() {
			return f, nil
		}
	}
	return "", os.ErrNotExist
}
