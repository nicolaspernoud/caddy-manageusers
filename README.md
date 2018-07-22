# caddy-manageusers
Caddy plugin designed to work with http.login (loginsrv) plugin. Allow creation, update, deletion of users in .htpasswd file and user-file.
Based on https://github.com/foomo/htpasswd for htpasswd management.

## Setup
Add a directive in your Caddyfile :
manageusers /[route to access API] [.htpasswd file name] [optional : users.json file]

Example : manageusers /manageusers .htpasswd users.json

## Usage

### Retrieve user list
```http
GET http://[hostname:port]/manageusers  HTTP/1.1
```

### Add or alter an user
```http
POST http://[hostname:port]/manageusers HTTP/1.1
content-type: application/json

{
    "username": "user",
    "password": "pwd",
    "origin": "htpasswd",
    "role": "user",
    "name": "Us",
    "surname": "ER"
}
```

### Delete an user
```http
DELETE http://[hostname:port]/manageusers/[username] HTTP/1.1
```

## NB
The loginsrv (http.login) plugin is supposed to work with a yaml file, but since json is valid yaml, it works flawlessly.
The user management route must be protected by jwt to disable unwanted users creation (which is a major security flaw).