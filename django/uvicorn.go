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

type Uvicorn struct {
	Cmd *exec.Cmd
	url string
	reverseProxy *httputil.ReverseProxy
}

func StartUvicorn(python *python.Python, sock string, app string) (*Uvicorn, error) {	
	args := []string{`-m`,`uvicorn`,`--uds`,sock,fmt.Sprintf("%s.asgi:application", app)}
	// args := []string{`-m`,`uvicorn`,`--uds`,sock,fmt.Sprintf("%s.asgi:application", app)}
	uvicornUrl := fmt.Sprintf(`http+unix://%s:`, sock)
	destUrl, err := url.Parse(uvicornUrl)
	if nil!=err {
		return nil, fmt.Errorf(`Failed to parse URL to uvicorn '%s' : %w`, uvicornUrl, err)
	}
	d := &Uvicorn{
		Cmd: python.Command(nil, args...),
		url: uvicornUrl,
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

func (d *Uvicorn) Url() string {
	return d.url
}

func (d *Uvicorn) Close() error {
	return d.Cmd.Process.Kill()
}

func (d *Uvicorn) ReverseProxy() *httputil.ReverseProxy {
	return d.reverseProxy
}