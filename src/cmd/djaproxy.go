package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/craigmj/commander"

	"djaproxy/django"
	"djaproxy/upstart"
)

func main() {
	if err := commander.Execute(nil,
		WebCommand,
		upstart.UpstartScriptCommand); nil != err {
		log.Fatal(err)
	}
}

func WebCommand() *commander.Command {
	fs := flag.NewFlagSet("web", flag.ExitOnError)
	port := fs.String("http", ":8001", "Bind address and port for webserver")
	dest := fs.String("dest", "http://localhost:8000", "Destination for reverse proxy - where django is running")
	app := fs.String("app", "", "Name of the web app")
	dir := fs.String("dir", "", "Root directory for web app (where manage.py is)")
	collect := fs.Bool("collect", true, "Run collectstatic before startup")

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
			m, err := django.UrlMap(*dir, *app)
			if nil != err {
				return err
			}
			for k, v := range m {
				log.Println(k, "=", v)
			}
			django.HttpMapStatics(m)
			if *collect {
				if err = django.CollectStatic(*dir); nil != err {
					return err
				}
			}

			destUrl, _ := url.Parse(*dest)
			http.Handle("/", httputil.NewSingleHostReverseProxy(destUrl))
			return http.ListenAndServe(*port, nil)
		})
}
