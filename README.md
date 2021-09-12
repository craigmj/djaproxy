# djaproxy

Very simple proxy server written in Go for serving django applications.

Installing django on a webserver can be a royal PiTA, especially if you have other 
applications and services running on the server. djaproxy aims to make it as simple
as possible, taking responsibility for:

1. Running a webserver that will serve your django application (it proxies circus) and that will 
   server your django static assets;
1. Installing all requirements and handling the running of these in a custom python environment; 
1. Installing itself as a service with systemctl;
1. Providing a basic ansible script for installing an application.

## Installation

Should be enough to run `make`. Then `bin/djaproxy` should be available.

## Usage

### Installing Python

1. Go to the directory of your django application where `manage.py` is.
1. `djaproxy python install` will build and install a custom python version in a `python` subdirectory.

### Setting up Django with requirements

You should have a `requirements.txt` which should contain at least `Django` and `daphne`.

1. `$(djaproxy python env)` will configure your shell with the environment variables to use the installed python.
1. `djaproxy python run -- -m pip install -r requirements.txt` will install all requirements from your `requirements.txt` which should include Django, daphne, et al.
1. `python3 manage.py migrate` to perform any necessary migrations
1. `python3 manage.py createsuperuser` to create your superuser account

### Installing systemd service

1. `djaproxy systemd-install -name NAME_OF_SERVICE -user USER_TO_RUN_AS -group GROUP_TO_RUN_AS -- web -bind INTERFACE_TO_LISTEN_ON(eg :21091) -app DJANGO_APPLICATION -dir DJANGO_APPLICATION_DIR` will install a systemd service that will create static resources, and run your Django application on the given port.
