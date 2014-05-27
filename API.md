# API
Current implementation don't care which HTTP method is used (GET/POST/WHATEVER).
Also, the token/service association is *not* checked.

The association key/IP address will be checked for every requests but
discover.

## Discover
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

## Update
Update the key of the service. If one needs to update more, it
may very well be safer to register a new service and ask admins
to delete the old one.

  /api/update key=...

Argument:

* key : current key

Returns

* new key
* ko : wrong key

## Info

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

## Login
Login a user:

  /api/login key=... login=... 

Arguments:

* key : service key
* login : either token, name or email

Returns:

* new : a new token has been generated for the name/email
* ok : the token was a valid one (effective login)
* ko : wrong key, wrong token, name, email

## Chaining
Chain a token (return a new token from a valid one)

  /api/chain key=... token=...

Arguments:

* key : service key
* token : user token

Returns:

* a new token
* ko : wrong key, wrong token

## Logout
Logout an user:

  /api/logout key=... token=... 

Arguments:

* key : service key
* token : user token

Returns

* ok

## Bridge
Make a bridge between two services. A bridge is a token
chain associated to an user. It allows services to provides
means to each others on behalf of an authentication server.

  /api/bridge name=... token=... key=...

Arguments:

* name : name of the other side of the bridge
* token : token identifying the user
* key : key of the service who want to established the bridge

Returns

* a token to be send to the 'name' service
* ko : wrong key, wrong token, unknown service

