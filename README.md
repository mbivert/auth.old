# General description
Auth is an authentication server for HTTP services.
It works without password, using one-time tokens only.

One time-token reduces replay attacks. Note however that
services using the auth are not required to chain the token
each time.

# Options
Here are the supported options:

	(earth)% ./auth -help
	Usage of ./auth:
	  -conf="config.json": Configuration file

Configuration is detailled later.

# Files description
## Sample deployment : INSTALL
INSTALL describes how to completely setup a working AAS along
with the example service.

## API Description : API
This file contains description of the HTTP API provided
by the AAS under /api/, plus some implementation details

## Next steps : TODO
Contains what should/shall be done soon or later.

## Certificates generation : genkey.sh
The genkey.sh use OpenSSL to generate a x509 certificate (default: cert.pem)
and an associated private key (default: key.pem). It also copies those
to example, and add the certificate in example/conf/.

The key is NOT password protected; Go seems to have trouble with that.

## Configuration : config.json
Configuration file is in JSON. Bad configuration file results
in undefined behavior(s). Fields are:

* URL : URL for the auth server
* Port : HTTP listening port for service
* Name : Name of the auth service
* AdminEmail : Email of the first administrator
* Mode : Service registration mode (either Automatic, Manual, Disable)
* Timeout : Lifetime of a token, (seconds)
* LenToken : Length of a token (bytes)
* LenKey : Length of a service key (bytes)
* VerifyCaptcha : Check captcha validity
* SSL : Service accessible via HTTPs
* Certificate : x509 certificate
* PKey : Key associated to Certificate
* SMTPServer : Server from which tokens are sent to users
* SMTPPort : Port to use for SMTPServer
* AuthEmail : Email for SMTPServer
* AuthPasswd : Password for AuthEmail

The service registration mode can be either

1. Automatic : every services can register and get a key from /api/discover
2. Manual : every services can register. Getting key and activate done by an administrator.
3. Disable : no services can register

## Sources
### webauth.go
Main file. Contains the HTTP handlers.

### auth.go
Contains the various function used for authentication. Usually
called from webauth.go

### database.go
Contains everything related to connecting and querying the database.
Also, contains a shaky services cache.

Usually called from auth.go

### utils.go
Contains some utility functions, mainly used in webauth.go

### token.go
Token management. Action on tokens are represented by interface Msg,
which contains a single process() method.

A goroutine ProcessMsg read actions one by one from a goroutine and the
execute their process() method. This garantee nice concurrent access.
A goroutine Timeouts is in charge of deleting token whose exceeded their
lifetime. Both goroutine are launched in webauth.go:/^func main

The message sending is hidden in a few utility functions whithin this files.

### config.go
Code to load the configuration file

### common.go
Some common data structures and errors.

### example/main.go
This is a sample service, configured through a conf/ directory.
It starts by loading a set of authentication services from all
the conf/*.conf files. Format is:

  variable=name

Variables are:

* url : URL to the auth server (eg. https://localhost:8080/)
* key : key to identify the service to the AAS (eg. retrieved from /api/discover)
* cert : name of the certificate of the AAS (will be searched in conf/)

A sample file can be found at example/conf/auth.conf.

Remind that genkey.sh will generate/install x509 pair for both
the AAS and example. It will also copy the certificate to example/conf/.

It then starts a web server on https://localhost:8082 (-port to change port,
SSL is default), from which a user can choose an authentication
server and login.
