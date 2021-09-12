package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	// "net/http/httputil"
	// "net/url"
	`os`
	`path/filepath`

	"github.com/craigmj/commander"

	"djaproxy/ansible"
	"djaproxy/circus"
	"djaproxy/django"
	"djaproxy/upstart"
	`djaproxy/systemd`
	`djaproxy/python`
)

func main() {
	if err := commander.Execute(nil,
		WebCommand,
		ansible.AnsibleCommand,
		upstart.UpstartCommand,
		upstart.SysVCommand,
		circus.InstallCommand,
		systemd.SystemdInstallCommand,
		python.PythonCommand,
	); nil != err {
		log.Fatal(err)
	}
}

func WebCommand() *commander.Command {
	fs := flag.NewFlagSet("web", flag.ExitOnError)
	bind := fs.String("bind", ":8001", "Bind address and port for webserver")
	// dest := fs.String("dest", "http://localhost:8000", "Destination for reverse proxy - where django is running")
	app := fs.String("app", "", "Name of the web app")
	wd, err := os.Getwd()
	if nil!=err {
		panic(err)
	}
	dir := fs.String("dir", wd, "Root directory for web app (where manage.py is)")
	collect := fs.Bool("collect", true, "Run collectstatic before startup")
	pythonPath := fs.String("python", "python", "python directory (installed with djaproxy python install)")

	return commander.NewCommand(
		"web",
		"Start the web proxy server",
		fs,
		func(args []string) error {
			if "" == *app {
				return errors.New("You need to specify the name fo the web app (-app)")
			}
			if "" == *dir {
				return errors.New("You need to specify the root directory for the web app (-dir)")
			}
			python, err := python.New(*pythonPath, ``)
			if nil!=err {
				return err
			}

			m, err := django.UrlMap(python, *dir, *app)
			if nil != err {
				return err
			}
			for k, v := range m {
				log.Println(k, "=", v)
			}
			django.HttpMapStatics(m)
			if *collect {
				if err = django.CollectStatic(python, *dir); nil != err {
					return err
				}
			}

			daphneSock := filepath.Join(wd, fmt.Sprintf(`daphne-%s.sock`, *app))
			daphne, err := django.StartDaphne(python, daphneSock, *app)
			defer daphne.Close()

			http.Handle("/", daphne.ReverseProxy())
			fmt.Println("Starting server on", *bind, ", proxy on", daphne.Url())
			return http.ListenAndServe(*bind, nil)
		})
}
