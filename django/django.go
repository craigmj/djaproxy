package django

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	`djaproxy/python`
)

// UrlMap uses python and the settings.py to work out the mappings
// of urls to directories
func UrlMap(python *python.Python, dir string, app string) (map[string]string, error) {
	var err error
	code := template.Must(template.New("").Parse(test_py))
	var pyCode bytes.Buffer
	if err := code.Execute(&pyCode, map[string]string{"App": app}); nil!=err {
		return nil, fmt.Errorf(`Failed parsing python template: %w`, err)
	}

	py := python.Command(nil)
	py.Dir = dir
	log.Printf(`Running urlmap command in %s`, py.Dir)
	for _, e := range py.Env {
		log.Printf("ENV: %s", e)
	}
	py.Stdin = nil
	pyin, err := py.StdinPipe()
	if nil!=err {
		return nil, err
	}
	go func() {
		io.Copy(pyin, bytes.NewReader(pyCode.Bytes()))
		pyin.Close()
	}()
	py.Stdout = nil
	pyout, err := py.StdoutPipe()
	if err = py.Start(); nil != err {
		return nil, fmt.Errorf("Error running UrlMap: %s", err.Error())
	}
	// py.Stdin = bytes.NewReader(pyCode.Bytes())
	var buf bytes.Buffer
	io.Copy(io.MultiWriter(&buf, os.Stdout), pyout)

	js := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	m := make(map[string]string, 0)
	if err = js.Decode(&m); nil != err {
		return nil, fmt.Errorf("Error on JSON Decode executing '%s' : %s", pyCode.String(), err.Error())
	}
	for u, p := range m {
		if !filepath.IsAbs(p) {
			m[u] = filepath.Clean(filepath.Join(dir, p))
		}
	}
	return m, nil
}

// HttpMapStatics maps all keys as urls to directories as strings
func HttpMapStatics(m map[string]string) {
	for u, root := range m {
		prefix := u
		l := len(prefix)
		if prefix[l-1] == '/' {
			prefix = prefix[0 : l-1]
		}
		log.Printf("Handling %s (%s) to %s", u, prefix, root)
		http.Handle(u, http.StripPrefix(prefix, http.FileServer(http.Dir(root))))
	}
}

// CollectStatics runs the python 'collectstatic' manage.py command
// in the directory given
func CollectStatic(python *python.Python, dir string) error {
	cmd := python.Command(nil, "manage.py", "collectstatic", "--noinput")
	cmd.Dir = dir
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	err := cmd.Run()
	if nil != err {
		fmt.Println("ERROR: on collectstatic : %s", err.Error())
		return err
	}
	return nil
}

const test_py = `
from {{.App}} import settings
import json
d = {}
if hasattr(settings, 'STATIC_ROOT') and hasattr(settings,'STATIC_URL'):
	d[settings.STATIC_URL] = settings.STATIC_ROOT
if hasattr(settings, 'MEDIA_ROOT') and hasattr(settings,'MEDIA_URL'):
	d[settings.MEDIA_URL] = settings.MEDIA_ROOT
print (json.dumps(d))
`
