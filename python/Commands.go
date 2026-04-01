package python

import (
	`bufio`
	`fmt`
	`flag`
	`io`
	`os`
	`path/filepath`
	`strings`

	`github.com/craigmj/commander`
)

const PYTHON_VERSION="3.14.3"

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
	ver := fs.String(`version`, ``, `Python version to install (will default to ` + PYTHON_VERSION + ` if unspecified)`)
	out := fs.String(`out`,``, `output enviornment vars to file`)
	shebang := fs.Bool(`shebang`,false,`Output bash shebang`)
	return commander.NewCommand(
		`env`,
		`Outputs python env for shell sourcing`,
		fs,
		func(args []string) error {
			var stdout io.Writer
			if ``!=*out {
				outf, err := os.OpenFile(*out, os.O_CREATE|os.O_WRONLY, 0755)
				if nil!=err {
					return err
				}
				defer outf.Close()
				outb := bufio.NewWriter(outf)
				defer outb.Flush()
				stdout = outb
			} else {
				stdout = os.Stdout
			}
			if *shebang {
				fmt.Fprintln(stdout, "#!/usr/bin/bash")
			}
			pairs, err := ParseEnvIntoSetStrings(strings.Join(args, `,`))
			if nil!=err {
				return err
			}
			py, err := New(*dir, *ver)
			if nil!=err {
				return err
			}
			for _, e := range py.Env(pairs) {
				fmt.Fprintln(stdout, `export`, e)
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
	ver := fs.String(`version`, ``, `Python version to install (will default to ` + PYTHON_VERSION + ` if unspecified)`)
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
	ver := fs.String(`version`, ``, `Python version to install (will default to ` + PYTHON_VERSION + ` if unspecified)`)
	return commander.NewCommand(
		`run`,
		`Runs arguments on python`,
		fs,
		func(args []string) error {
			py, err := New(*dir, *ver)
			if nil!=err {
				return err
			}
			return py.Command(os.Environ(), args...).Run()
		})
}
