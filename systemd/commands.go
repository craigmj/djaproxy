package systemd

import (
	`flag`

	`github.com/craigmj/commander`
)

func SystemdInstallCommand() *commander.Command {
	fs := flag.NewFlagSet(`systemd-install`, flag.ExitOnError)
	name := fs.String(`name`,``,`Name of the service`)
	user := fs.String(`user`,``,`User to run as`)
	group := fs.String(`group`,``,`Group to run as`)
	workingDir := fs.String(`dir`,``,`Working directory`)

	return commander.NewCommand(`systemd-install`,
		`Install djaproxy'd system as a systemctl service`,
		fs,
		func(args []string) error {
			return SystemdInstall(*name, *user, *group, *workingDir, args)
		})
}