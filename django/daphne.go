package django

import (
	`fmt`
	`os`
	`os/exec`
	`strings`
	`net/http/httputil`
	`net/url`
	`net/http`

	`github.com/peterbourgon/unixtransport`

	`djaproxy/python`
)

type Daphne struct {
	Cmd *exec.Cmd
	url string
	reverseProxy *httputil.ReverseProxy
}

func StartDaphne(python *python.Python, sock string, app string) (*Daphne, error) {	
	args := []string{`-m`,`daphne`,`-u`,sock,fmt.Sprintf("%s.asgi:application", app)}
	daphneUrl := fmt.Sprintf(`http+unix://%s:`, sock)
	destUrl, err := url.Parse(daphneUrl)
	if nil!=err {
		return nil, fmt.Errorf(`Failed to parse URL to daphne '%s' : %w`, daphneUrl, err)
	}
	d := &Daphne{
		Cmd: python.Command(nil, args...),
		url: daphneUrl,
		reverseProxy: httputil.NewSingleHostReverseProxy(destUrl),
	}
	transport := &http.Transport{}
	d.reverseProxy.Transport = transport
	unixtransport.Register(transport)

	d.Cmd.Stdout, d.Cmd.Stderr = os.Stdout, os.Stderr
	if err := d.Cmd.Start(); nil!=err {
		return nil, fmt.Errorf(`Error on '%s': %w`, strings.Join(args, ` `), err)
	}
	return d, nil
}

func (d *Daphne) Url() string {
	return d.url
}

func (d *Daphne) Close() error {
	return d.Cmd.Process.Kill()
}

func (d *Daphne) ReverseProxy() *httputil.ReverseProxy {
	return d.reverseProxy
}