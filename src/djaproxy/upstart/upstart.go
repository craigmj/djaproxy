package upstart

import (
	"errors"
	"flag"
	"io"
	"os"
	"text/template"

	"github.com/craigmj/commander"
)

// UpstartScript writes the upstart script to the output writer.
// The cwd is the tjomma directory.
func UpstartScript(script, dir, app string, out io.Writer) error {
	t := template.Must(template.New("").Parse(script))
	data := struct {
		App string
		Dir string
	}{App: app, Dir: dir}
	return t.Execute(out, &data)
}

func UpstartCommand() *commander.Command {
	fs := flag.NewFlagSet("upstart", flag.ExitOnError)
	dir := fs.String("dir", "", "Home directory")
	app := fs.String("app", "", "App name")
	return commander.NewCommand(
		"upstart",
		"Output the upstart scripts to start the app and its circus server (sudo this)",
		fs,
		func(args []string) error {
			if "" == *dir {
				return errors.New("You need to specify the home directory (-dir)")
			}
			if "" == *app {
				return errors.New("You need to specify the app name (-app)")
			}

			err := writeUpstart(_upstart_circus,
				*app+"-circus",
				*dir, *app)
			if nil != err {
				return err
			}
			err = writeUpstart(_upstart_proxy,
				*app+"-proxy",
				*dir, *app)
			if nil != err {
				return err
			}
			return nil
		})
}

func writeUpstart(script, dest, dir, app string) error {
	out, err := os.Create("/etc/init/" + dest + ".conf")
	if nil != err {
		return err
	}
	defer out.Close()
	return UpstartScript(script, dir, app, out)
}

var _upstart_circus = `
# {{.App}} circus
description "{{.App}}-circus"

start on runlevel [2345]
stop on runlevel [!2345]

respawn
expect fork

script
	cd '{{.Dir}}'
	. {{.Dir}}/bin/activate
    {{.Dir}}/bin/circusd --daemon circus.ini &
end script

emits {{.App}}_circus_starting
`

var _upstart_proxy = `
# {{.App}} djaproxy for the app
#
description	"{{.App}}-proxy"

start on runlevel [2345]
stop on runlevel [!2345]

respawn
expect fork

script 
  cd '{{.Dir}}'
  djaproxy web -app '{{.App}}' -dir '{{.Dir}}' &
end script

emits {{.App}}_proxy_starting
`
