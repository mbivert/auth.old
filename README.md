= General description
Auth is an authentication server for HTTP services.
It works without password, using one-time tokens only.

One time-token reduces replay attacks. Note however that
services using the auth are not required to chain the token
each time.

= Certificates generation
The genkey.sh use OpenSSL to generate a x509 certificate (cert.pem)
and an associated private key (key.pem). It also copies those
to example, and add the certificate in example/conf/.

The key is NOT password protected; Go seems to have trouble with that.

= Options
Here are the supported options:
  (earth)% ./auth -help
  Usage of ./auth:
    -cert="cert.pem": x509 Certificate
    -email="admin@whatev.er": First administrator email
    -from="admin@whatev.er": Email to send tokens from
    -key="key.pem": Private key to sign certificate
    -passwd="wrong password": Password for email
    -port="8080": Listening HTTP port
    -smtpp="587": Port of SMTP server
    -smtps="smtp.gmail.com": SMTP server
    -url="https://auth.awesom.eu": URL for AAS
    -vc=true: Verify captcha

You should definitely have a look at -smtpp -smtps -email -from -passwd
in order to do some basic deployment.

= Building and launching

== Database (PostgreSQL)
Assuming you installed the latest PostgreSQL following
their INSTALL file.

Creating and starting the auth database:

  (earth)% createdb auth

Connect to this database, and add a new role:

  (earth)% psql auth
  psql (9.3.4)
  Type "help" for help.
  
  auth=# CREATE ROLE auth PASSWORD 'auth' LOGIN;
  CREATE ROLE
  auth=# \q

== Auth & example
You may refer to the next § for more informations.

After setting the correct options:

  (earth)% go build . && (./auth $OPTIONS &)
  2014/05/17 21:08:16 Launching on https://localhost:8080
  (earth)% cd example/ && go build .
  # -k to blindly accept certificate
  (earth)% curl -k --data 'name=example&url=http://example.awesom.eu/&address=127.0.0.1&email=admin@whatev.er' https://localhost:8080/api/discover

Curl should reply 'ok'. This means the key has been generated.
The server does not send the key by default (see the three modes
later)

Then, login to https://localhost:8080 using your admin account, go
to the admin page, fetch the key for the service 'example' and
add it to example/conf/auth.conf.

You may then launch example/:

  (earth)% ./example
  2014/05/17 22:12:34 Launching on https://localhost:8082

Browse to https://localhost:8082 and:

* Use your admin account ('admin' username)
* Go to the AAS and fetch the token for 'example'
* Enter the token on 'example'.
* ...
* PROFIT.

= Detailled usage
(please, ignore what templates/index.html say)

== Registration
Registration process requires the user to

* have a valid email account
* choose an unused, valid nickname (no spaces and @ allowed)
* (eventually, be able to read captcha)

Once those informations are supplied on /register,
the user will be sent a token which he may use to login.

== Login
=== To the AAS
If the user don't have a token, he simply gives the AAS its
nickname or email and wait for a new token to be generated and
sent to his email address.

The user fetch this token, and then enter it to be logged-in.

=== To other services (eg. example/)
The process is similar: if the user have no token for the
service, he gives either his nickname or email. He then
go to his AAS account, fetch the token and connect to the
service.

=== Administrator
Administrator may view the various registered services, and
eventually enable/disable them.

He may also set the registration mode to:

* Automatic : server accept every services without administrator intervention
* Manual (default) : server accept every services, but administrator shall manually activate services and send the key to the services
* Disable : no services are allowed to register.

An email is sent to every administrator when Automatic mode is enable.
(this is overkill)

== API
Current implementation don't care which HTTP method is used (GET/POST/WHATEVER).
Also, the token/service association is *not* checked.

The association key/IP address will be checked for every requests but
discover.

=== Discover
To register a new service:

  /api/discover name=... url=... address=... email= 

Arguments:

* name : name of the service
* url : url of the service
* address : IP address of the service
* email : who to contact (eg. to send the key)

Returns:

* a key identifying the service (Automatic mode only)
* ok : service registered (Manual mode only)
* ko : Wrong paramaters or Disable mode

=== Update (NOT IMPLEMENTED)
Update the key of the service

  /api/update key=...

Argument:

* key : current key

Returns

* new key
* ko : wrong key

=== Info

Retrieve info about the owner of a token:

  /api/info key=... token=...

Argument:

* key : service key
* token : token of the user

Returns:

* one per line and in this order
    * id
    * name
    * email
* ko : wrong key, wrong token)

=== Login
Login a user:

  /api/login key=... login=... 

Arguments:

* key : service key
* login : either token, name or email

Returns:

* new : a new token has been generated for the name/email
* ok : the token was a valid one (effective login)
* ko : wrong key, wrong token, name, email

=== Chaining
Chain a token (return a new token from a valid one)

  /api/chain key=... token=...

* key : service key
* token : user token

Returns:

* a new token
* ko : wrong key, wrong token

=== Logout
Logout an user:

  /api/logout key=... token=... 

* key : service key
* token : user token

Returns nothing.

== Few words on example/
The example/ directory contains a basic service (main.go) which
supports authentication with multiple auth servers. Those are
supplied through conf/*.conf files. Format is:

  variable=name

Variables are:

- url : URL to the auth server (eg. https://localhost:8080/)
- key : key to identify the service to the AAS (eg. retrieved from /api/discover)
- cert : name of the certificate of the AAS (file will be searched in conf/)

A sample file can be found at example/conf/auth.conf.
genkey.sh will generate/install x509 pair for both the AAS and
example. It will also copy the certificate to example/conf/.

= TODO
== Small steps
By a semblance of order of importance

* Clean things.
* captcha : don't remember.
* Wrong navbar when connected.
* Avoid the double-captcha at login.
* Rework templates/index.html
* Avoid captcha in the general case. Two-steps login is already painful.
* Use a configuration file instead of the 9000 options. (also, add -timeout)
* Clean variables naming/code "isolation"
* /api/update is not implemented.
* /unregister is not implemented.
* Use SSL to communicate with PostgreSQL (not important assuming isolated communication)
* Maybe use different storage than PostgreSQL (keep reading why it may not be that good)

== Heavy load
=== On Website and API
To be seen in production, but one-time token should generate
quite some traffic.

Multiple auth servers could be launched, communicating to
the same database, and sharing the same token cache.
The cache may be implemented using [memcached](http://memcached.org/)
or [groupcache](https://github.com/golang/groupcache)

Balancing the load on the multiple auth servers can easily
be done with [Nginx](http://nginx.org/en/docs/http/load_balancing.html).

=== On Database
PostgreSQL already comes with some interesting
[features](http://www.postgresql.org/docs/9.4/static/high-availability.html)
which would help managing the load on the Database if any.

=== On SMTP
For now, we use a single, external SMTP server (for developping).
One should setup his own SMTP server in production.

However, if needed, multiple SMTP server may be used, emails
being dispatched following who they are sending to (eg. use
gmail's SMTP to communicate with @gmail.com addresses).

== Enforce chaining
It might be good to enforce token chaining by services within
the "protocol".
This could easily be done by removing info, and having login
returns those data.
