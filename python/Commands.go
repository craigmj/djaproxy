package python

import (
	`fmt`
	`flag`
	`os`
	`path/filepath`

	`github.com/craigmj/commander`
)

func PythonCommand() *commander.Command {
	return commander.NewCommand(`python`,
		`python related commands`,
		nil,
		func(args []string) error {
			return commander.Execute(args,
				EnvironmentCommand,
				InstallCommand,
				RunCommand,
				)
		})
}

func EnvironmentCommand() *commander.Command {
	fs := flag.NewFlagSet(`env`, flag.ExitOnError)
	wd, err := os.Getwd()
	if nil!=err {
		panic(err)
	}
	dir := fs.String(`dir`, filepath.Join(wd,`python`), `Directory in which to install python`)
	ver := fs.String(`version`, ``, `Python version to install (will default to 3.9.7 if unspecified)`)
	return commander.NewCommand(
		`env`,
		`Outputs python env for shell sourcing`,
		fs,
		func(args []string) error {
			py, err := New(*dir, *ver)
			if nil!=err {
				return err
			}
			for _, e := range py.Env() {
				fmt.Println(`export`, e)
			}
			return nil
		})
}

func InstallCommand() *commander.Command {
	fs := flag.NewFlagSet(`install`, flag.ExitOnError)
	wd, err := os.Getwd()
	if nil!=err {
		panic(err)
	}
	dir := fs.String(`dir`, filepath.Join(wd,`python`), `Directory in which to install python`)
	ver := fs.String(`version`, ``, `Python version to install (will default to 3.9.7 if unspecified)`)
	return commander.NewCommand(
		`install`,
		`Installs a local python interpreter into given directory`,
		fs,
		func(args []string) error {
			py, err := New(*dir, *ver)
			if nil!=err {
				return err
			}
			return py.Install()
		})
}

func RunCommand() *commander.Command {
	fs := flag.NewFlagSet(`run`, flag.ExitOnError)
	wd, err := os.Getwd()
	if nil!=err {
		panic(err)
	}
	dir := fs.String(`dir`, filepath.Join(wd,`python`), `Directory in which to install python`)
	ver := fs.String(`version`, ``, `Python version to install (will default to 3.9.7 if unspecified)`)
	return commander.NewCommand(
		`run`,
		`Runs arguments on python`,
		fs,
		func(args []string) error {
			py, err := New(*dir, *ver)
			if nil!=err {
				return err
			}
			return py.Command(nil, args...).Run()
		})
}