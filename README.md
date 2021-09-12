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

