## GO JWT Identity Provider
Implementation of a JWT-based identity provider written in go.
Main features:
 - REST API for user CRUD operations
 - REST API for JWT token generation/renewal
 - support for user/password based token generation and m2m key based token generation
 - openapi documentation
 - data layer ORM based on gorm library
 - docker compose and helm chart deployment

The microservice can be easily embedded in an application in order to provide authentication services.
It is sufficient to:
 - read the public loaded into the goidp application
 - use the key to validate incoming API request (if the token can be decoded with the provided public key, it is trusted)

## Test
```
 go test ./... -v
```

### Configure
The identity provider application use environment variables for configuration
```
# db
DB_NAME=test                              # Name of database
DB_USER=root                              # Database username
DB_PASS=example                           # Database password
DB_HOST=10.150.4.190                      # Database host
DB_PORT=3306                              # Database port
DB_CHARSET=utf8                           # Database charset
DB_PARSE_TIME=True                        # Database parse time
DB_SHOW_SQL=True                          # Print all SQL database queries

# jwt
# either set JWT_KEY_PATH or JWT_SECRET based on the type of jwt auth you need
# if both are set, key authentication is used
# JWT_SECRET=/path/to/secret              # set the jwt secret
JWT_USE_KEY=True                          # whether to use key auth or not
JWT_PRIVATE_KEY=./private.pem             # private key
JWT_PUBLIC_KEY=./public.pem               # public key
JWT_ACCESS_EXPIRE_TIME=5m                 # access token expire time
JWT_REFRESH=True                          
JWT_REFRESH_EXPIRE_TIME=120

# app
APP_HOST=0.0.0.0                          # auth server host
APP_PORT=8889                             # auth server listen port
APP_WRITE_TIMEOUT=15                      # auth server write timeout
APP_READ_TIMEOUT=15                       # auth server read timeout
APP_IDLE_TIMEOUT=60                       # auth server idle timeout
APP_LOG_LEVEL=debug                       # auth server log level
```

## Run
Generate the key pair required for JWT authentication with the helper script generate_secrets.sh
Therefore run the app.

### helm
values json schema generated through readme-generator-for-helm.

Install locally with:

```helm install --kubeconfig $KUBECONFIG  idp ./ --namespace gcortesi --create-namespace```

### docker-compose
The docker compose deployment run:
 - the identity provider application
 - a postgres database
 - the adminer GUI for the postgres database

`docker-compose up -d --build`

### manual
To manually run the identity provider application:
```
cd idp/cmd
go build -o idp
./idp
```

## Documentation
REST API specification of the identity provider application is provided in jsonapi format.
Display the documentation with swagger ui:
```
docker run -d -p 9999:8080 -e URL=idp.yaml --rm --name openapi -v $PWD/idp.yaml:/usr/share/nginx/html/idp.yaml:Z swaggerapi/swagger-ui
```


