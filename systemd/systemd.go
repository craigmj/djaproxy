package systemd

import (
	`fmt`
	`os`
	`os/exec`
	`path/filepath`
	`text/template`
)

func getName(name string) string {
	if ``==name {
		return filepath.Base(os.Args[0])
	}
	return name
}

type config struct {
	Name string
}

func NewConfig(name string) *config {
	return &config{ 
		Name: getName(name),
	}
}

func (c *config) ServicePath() string {
	return `/lib/systemd/system/` + c.Name + `.service`
}

func SystemdServiceRestartEdit(serviceName, waitFor string) error {
	return fmt.Errorf(`SystemdServiceRestartEdit not implemented: done with /lib/systemd/system/X.service.d/custom.conf file instead`)
}

func SystemdInstall(name, user, group, workingDir string, args []string) error {
	var err error
	service := NewConfig(name)
	if ``==workingDir || !filepath.IsAbs(workingDir) {
		wd, err := os.Getwd()
		if nil!=err {
			return fmt.Errorf(`Failed to getwd: %w`, err)
		}
		if ``==workingDir {
			workingDir = wd
		} else {
			workingDir = filepath.Clean(filepath.Join(wd, workingDir))
		}
	}
	executable, err := os.Executable()
	if nil!=err {
		return fmt.Errorf(`Finding executable failed: %w`, err)
	}
	out, err := os.Create(service.ServicePath())
	if nil!=err {
		return fmt.Errorf(`Creating ServicePath() failed: %w`, err)
	}
	defer out.Close()
	if err := _systemdService.Execute(out, map[string]interface{} {
		"WorkingDirectory":workingDir,
		"Executable": executable,
		"Args":args,
		"Name":service.Name,
		"User":user,
		"Group":group,
	}); nil!=err {
		return err
	}
	if err := exec.Command(`systemctl`,`daemon-reload`).Run(); nil!=err {
		return err
	}
	if err := exec.Command(`systemctl`, `enable`, service.Name).Run(); nil!=err {
		return err
	}
	if err := exec.Command(`systemctl`,`stop`,service.Name).Run(); nil!=err {
		return err
	}
	if err := exec.Command(`systemctl`, `start`, service.Name).Run(); nil!=err {
		return err
	}
	return nil
}

func SystemdUninstall(name string) error {
	service := NewConfig(name)
	if err := exec.Command(`systemctl`, `stop`, service.Name).Run(); nil!=err {
		log.Error(err)
	}
	if err := exec.Command(`systemctl`,`disable`, service.Name).Run(); nil!=err {
		log.Error(err)
	}
	if err := os.Remove(service.ServicePath()); nil!=err {
		log.Error(err)
	}
	if err := exec.Command(`systemctl`, `daemon-reload`).Run(); nil!=err {
		return err
	}
	return nil
}

var _systemdService = template.Must(template.New(``).Parse(`
[Unit]
Description={{.Name}}
;After=mysql.service
;Requires=mysql.service

[Install]
WantedBy=multi-user.target

[Service]
Type=simple
WorkingDirectory={{.WorkingDirectory}}
{{with .User}}User={{.}}{{end}}
{{with .Group}}Group={{.}}{{end}}
ExecStart={{.Executable}} {{range .Args}}"{{.}}" {{end}}
Restart=on-failure
RestartSec=5s
`))