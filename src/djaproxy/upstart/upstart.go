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
func UpstartScript(dir, app string, out io.Writer) error {
	t := template.Must(template.New("").Parse(_upstart))
	data := struct {
		App string
		Dir string
	}{App: app, Dir: dir}
	return t.Execute(out, &data)
}

var _upstart = `
# djaproxy for {{.App}}
#
description	"django_{{.App}} webproxy"

start on runlevel [2345]
stop on runlevel [!2345]

respawn
expect fork

script 
  cd '{{.Dir}}'
  exec djaproxy web -app '{{.App}}' -dir '{{.Dir}}' &
end script

emits django_{{.App}}_proxy_starting
`

func UpstartScriptCommand() *commander.Command {
	fs := flag.NewFlagSet("upstart-script", flag.ExitOnError)
	dir := fs.String("dir", "", "Home directory")
	app := fs.String("app", "", "App name")
	return commander.NewCommand(
		"upstart-script",
		"Output the upstart script",
		fs,
		func(args []string) error {
			if "" == *dir {
				return errors.New("You need to specify the home directory (-dir)")
			}
			if "" == *app {
				return errors.New("You need to specify the app name (-app)")
			}
			return UpstartScript(*dir, *app, os.Stdout)
		})
}
