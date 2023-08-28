**CDB must be installed on a windows system and added to the environment variable**

# How to use
- **CDB must be installed on a windows system and added to the environment variable**
- Website is included in src but must be hosted using a seperate server like apache2 or nginx
- Edit upload.js at the top if you want to point the site to a different port or ip

# Server options
```
ip = IP address to bind, defaults to localhost
port = port number to listen on, defaults to 25478
tlsport = port number to listen on with TLS, defaults to 25443
upload_limit = max size of uploaded file (byte), default max of 15 mb
token = specify the security token, defaults to a f9403fc5f537b4ab332d if left blank will be random
protectedMethodFlag = specify methods intended to be protect by the security token, defaults to none
logLevelFlag = logging level, defaults to info
certFile = path to certificate file, defaults to no TLS
keyFile = path to key file, defaults to no TLS
corsEnabled = if true, add ACAO header to support CORS, Defaults to true
blank := Where to save the dumps
```

# TODO
- Make backend host frontend
- Add linux support for crash analysis server
- ~~Check validity of dump before analysis~~
- Add support for other dump analysis tools
- Add support for other dump formats
