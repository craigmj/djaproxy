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
		Cmd: python.Command(os.Environ(), args...),
		url: uvicornUrl,
		reverseProxy: httputil.NewSingleHostReverseProxy(destUrl),
	}
	origDirector := d.reverseProxy.Director
	d.reverseProxy.Director = func(r *http.Request) {
		origDirector(r)
		ilog.Printf("Request for %s - remote-addr = %s, x-forwarded-for=%s", r.URL.String(), r.Header.Get(`Remote-Addr`), r.Header.Get(`X-Forwarded-For`))
		// if r.Header.Get(`Remote-Addr`)==`` {
		// 	r.Header.Add(`Remote-Addr`, strings.Split(r.Header.Get(`X-Forwarded-For`),  `, `)[0])
		// }
		// r.Header.Add(`x-djaproxy`, `v1`)
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