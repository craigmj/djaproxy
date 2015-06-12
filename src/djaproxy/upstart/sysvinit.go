package upstart

import (
	"errors"
	"flag"
	"io"
	"os"
	"text/template"

	"github.com/craigmj/commander"
)

func SysVInit(app, dir string, out io.Writer) error {
	var err error
	if "" == dir {
		dir, err = os.Getwd()
		if nil != err {
			return err
		}
	}
	t := template.Must(template.New("").Parse(_sysvinit))
	return t.Execute(out, map[string]interface{}{
		"App": app,
		"Dir": dir,
	})
}

func SysVInitWrite(app, dir string) error {
	out, err := os.Create("/etc/init.d/dja-" + app + ".sh")
	if nil != err {
		return err
	}
	defer out.Close()
	return SysVInit(app, dir, out)
}

func SysVCommand() *commander.Command {
	fs := flag.NewFlagSet("sysv", flag.ExitOnError)
	app := fs.String("app", "", "Name of the app to start")
	dir := fs.String("dir", "", "Directory where manage.py is")
	return commander.NewCommand("sysv",
		"Write sysvinit script to /etc/init.d",
		fs,
		func(args []string) error {
			if "" == *app {
				return errors.New("You must provide an app name (-app)")
			}
			return SysVInitWrite(*app, *dir)
		})
}

var _sysvinit = `#!/bin/bash
#!/bin/bash
#
# chkconfig: 35 90 12
# description: Sassidb server
#
# Get function from functions library
. /etc/init.d/functions
# Start the service FOO
start() {
        initlog -c "echo -n Starting sassidb server: "
        cd {{.Dir}}
        . {{.Dir}}/bin/activate
        {{.Dir}}/bin/circusd --daemon circus.ini
        djaproxy -dir {{.Dir}} -app {{.App}} &
        ### Create the lock file ###
        touch /var/lock/sassidb.lck
        success $"sassidb server startup"
        echo
}
# Restart the service FOO
stop() {
        initlog -c "echo -n Stopping sassidb server: "
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
