package python

import (
	`fmt`
	`bytes`
	// `net/http`
	// `io`
	`os`
	`os/exec`
	`strings`
	`path/filepath`
	`text/template`
)
// Python wraps a local installation of Python that works like venv
type Python struct {
	Dir string
	Version string
}

func New(dir, ver string) (*Python, error) {
	var err error
	if ``==ver {
		ver = PYTHON_VERSION
	}
	dir, err = filepath.Abs(dir)
	if nil!=err {
		return nil, err
	}
	p := &Python{
		Dir: dir,
		Version: ver,
	}
	return p, nil
}

func (p *Python) Install() error {
	// going to cheat and write this as a shell script
	bash := exec.Command(`/bin/bash`)
	var script bytes.Buffer
	if err := _pythonInstallScript.Execute(&script, map[string]interface{}{
		`Dir`: p.Dir,
		`ParentDir`: filepath.Dir(p.Dir),
		`Version`: p.Version,
	}); nil!=err {
		return fmt.Errorf(`Failed parsing script: %w`, err)
	}
	bash.Stdin = bytes.NewReader(script.Bytes())
	bash.Stdout, bash.Stderr = os.Stdout, os.Stderr
	if err := bash.Run(); nil!=err {
		return fmt.Errorf(`ERR on bash script: %w`, err)
	}

	// // Fetch the source .tar.gz
	// getUrl := fmt.Sprintf(`https://www.python.org/ftp/python/%s/Python-%s.tgz`, ver, ver)
	// get, err := http.Get(getUrl)
	// if nil!=err {
	// 	return fmt.Errorf(`GET request '%s' failed: %w`, getUrl, err)
	// }
	// if 200!=get.StatusCode {
	// 	return fmt.Errorf(`GET request response code %d: %s`, get.StatusCode, get.Status)
	// }
	// defer get.Body.Close()

	// targz := filepath.Join(p.Dir, fmt.Sprintf("python-%s.tgz", ver))
	// os.MkdirAll(filepath.Dir(targz), 0755)
	// out, err := os.Create(targz)
	// if nil!=err {
	// 	return fmt.Errorf(`Failed creating output file %s: %w`, targz, err)
	// }
	// defer out.Close()
	// _, err := io.Copy(out, get.Body)
	// if nil!=err {
	// 	return fmt.Errorf(`Failed writing %s: %w`, targz, err)
	// }
	// out.Close()

	// // Unzip tar.gz

	return nil
}

func (p *Python) Env(pairs []string) []string {
	pyenv := []string {
		`PYTHONHOME=` + p.Dir,
		`PATH=` + fmt.Sprintf(`%s:%s`, filepath.Join(p.Dir, `bin`), os.Getenv(`PATH`)),
	}
	return MergeEnv(pyenv, pairs)
}

// MergeEnv merges two environments, preferring values from 
// the first environment and preventing duplicates
func MergeEnv(major, minor []string) []string {
	out := make([]string, 0, len(major)+len(minor))
	keys := map[string]bool{}
	getKey := func(set string) string {
		return strings.ToUpper(strings.TrimSpace(strings.Split(set,`=`)[0]))
	}
	for _, e := range(major) {
		out = append(out, e)
		keys[getKey(e)] = true
	}
	for _, e := range minor {
		key := getKey(e)
		if !keys[key] {
			out = append(out,e)
			keys[key]=true
		}
	}
	return out
}

func (p *Python) Command(env []string, cmd ...string) *exec.Cmd {
	if nil==env {
		env = []string{}
	}

	py := exec.Command(filepath.Join(p.Dir, `bin`,`python3`), cmd...)
	py.Env = p.Env(env)
	py.Stdout, py.Stderr =  os.Stdout, os.Stderr
	py.Stdin = os.Stdin
	return py
}

var _pythonInstallScript = template.Must(template.New(``).Parse(`#!/bin/bash
if [[ ! -f {{.Dir}}/bin/python3 ]]; then
	set -xe
	mkdir -p {{.ParentDir}}/python-src
	cd {{.ParentDir}}/python-src
	if [[ ! -f Python-{{.Version}}.tgz ]]; then
		curl -LO https://www.python.org/ftp/python/{{.Version}}/Python-{{.Version}}.tgz
		tar -xzf Python-{{.Version}}.tgz
	fi
	cd Python-{{.Version}}

	./configure --enable-loadable-sqlite-extensions --prefix {{.Dir}}
	make -j $(nproc) && make install
fi
`))