package django

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

// UrlMap uses python and the settings.py to work out the mappings
// of urls to directories
func UrlMap(dir string, app string) (map[string]string, error) {
	code := template.Must(template.New("").Parse(test_py))
	py := exec.Command("bin/python")
	py.Dir = dir
	in, err := py.StdinPipe()
	if nil != err {
		return nil, err
	}
	out, err := py.StdoutPipe()
	if nil != err {
		return nil, err
	}
	if err = py.Start(); nil != err {
		return nil, fmt.Errorf("Error running UrlMap: %s", err.Error())
	}
	go func() {
		code.Execute(in, map[string]string{"App": app})
		in.Close()
	}()

	var buf bytes.Buffer
	io.Copy(&buf, out)
	io.Copy(os.Stdout, bytes.NewReader(buf.Bytes()))

	js := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	m := make(map[string]string, 0)
	if err = js.Decode(&m); nil != err {
		return nil, fmt.Errorf("Error on JSON Decode: %s", err.Error())
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
func CollectStatic(dir string) error {
	cmd := exec.Command("bin/python", "manage.py", "collectstatic", "--noinput")
	cmd.Dir = dir
	out, _ := cmd.StdoutPipe()
	go io.Copy(os.Stdout, out)
	errpipe, _ := cmd.StderrPipe()
	go io.Copy(os.Stderr, errpipe)

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
print json.dumps(d)
`
