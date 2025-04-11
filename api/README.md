# Cars API

This is a simple cars API backed with NodeJS and Express. The app runs on localhost and uses port 3000.
At any time you can change the port to whatever you want, see [Instructions](##Instructions).

## Installation of required packages
- Install [NodeJS](http://nodejs.org)
- Install [NPM](https://www.npmjs.com/package/npm) package manager
- Install required packages: 
```bash 
make build
```

## Instructions

To run the server you need simply to execute the following command:
```bash
make run
```
Or if you want to customize the port of the server (port 3001 for example), run:
```bash
PORT=3001 make run
```

Then you can access the data via the url: `http://localhost:${PORT}/api`

The api exposes the following endpoints:
```
GET /api/models
GET /api/models/{id}
GET /api/manufacturers
GET /api/manufacturers/{id}
GET /api/categories
GET /api/categories/{id}
```

The `image` property relates to an image for a `carModel`, and can be found in the `/api/images` directory.
