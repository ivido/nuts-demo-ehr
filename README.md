_Please note that this project is currently a work in progress. Not everything
will fully work yet_

# Nuts demo EHR system

This application is pretending to be an electronic health record system. You can
use it to demo how healthcare professionals can work together by sharing
information with colleagues through the Nuts nodes.

This version is an updated version using vue.js as frontend framework and a Golang backend. It's based on the [nuts-registry-admin-demo](https://github.com/nuts-foundation/nuts-registry-admin-demo).

Go to the [master](https://github.com/nuts-foundation/nuts-registry-admin-demo/tree/master/) branch to find the previous version of the app.

**NOTE THAT THIS APPLICATION IS NOT INTENDED FOR USE WITH REAL MEDICAL
INFORMATION! IT IS IN NO WAY DEVELOPED TO BE SAFE, STABLE OR EVEN USABLE FOR
SUCH PURPOSE.**

## Building and running
### Production
To build for production:

```shell
$ npm install
$ npm run build
$ go run .
```

This will serve the front end from the embedded filesystem.
### Development

During front-end development, you probably want to use the real filesystem and webpack in watch mode:

```shell
$ npm install
$ npm run watch
$ go run . live
```

The API and domain types are generated from the `api/api.yaml`.
```shell
$ oapi-codegen -generate server -package api api/api.yaml > api/generated.go
$ oapi-codegen -generate types -package domain -o domain/generated_types.go api/api.yaml

```

### Docker
```shell
$ docker run -p 1304:1304 nutsfoundation/nuts-demo-ehr:main
```

## Configuration
When running in Docker without a config file mounted at `/app/server.config.yaml` it will use the default configuration.

## Technology Stack

Frontend framework is vue.js 3.x

Icons are from https://heroicons.com

CSS framework is https://tailwindcss.com
