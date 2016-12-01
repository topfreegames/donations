Hosting Donations
=================

There are three ways to host Donations: docker, binaries or from source.

## Docker

Running Donations with docker is rather simple. Our docker container image comes bundled with the API binary. All you need to do is load balance all the containers and you're good to go. The API runs at port `8080` in the docker image.

Donations uses MongoDB to store clans information. The container takes environment variables to specify this connection:

* `DONATIONS_MONGO_HOST` - MongoDB host to connect to;
* `DONATIONS_MONGO_PORT` - MongoDB port to connect to;
* `DONATIONS_MONGO_DB` - Database name of the MongoDB Server to connect to.

Donations uses Redis for global locks. The container takes environment variables to specify this connection:

* `DONATIONS_REDIS_URL` - Redis URL to connect to;
* `DONATIONS_REDIS_MAXIDLE` - Max Idle connection to Redis;
* `DONATIONS_REDIS_IDLETIMEOUTSECONDS` - Number of seconds to consider a connection idle.

Other than that, there are a couple more configurations you can pass using environment variables:

* `DONATIONS_NEWRELIC_KEY` - If you have a [New Relic](https://newrelic.com/) account, you can use this variable to specify your API Key to populate data with New Relic API;
* `DONATIONS_NEWRELIC_APPNAME` - If you have a [New Relic](https://newrelic.com/) account, you can use this variable to specify the name of the application to use in your New Relic dashboard;
* `DONATIONS_SENTRY_URL` - If you have a [sentry server](https://docs.getsentry.com/hosted/) you can use this variable to specify your project's URL to send errors to.

If you want to expose Donations outside your internal network it's advised to use Basic Authentication. You can specify basic authentication parameters with the following environment variables:

* `DONATIONS_BASICAUTH_USERNAME` - If you specify this key, Donations will be configured to use basic auth with this user;
* `DONATIONS_BASICAUTH_PASSWORD` - If you specify `BASICAUTH_USERNAME`, Donations will be configured to use basic auth with this password.

### Example command for running with Docker

```
    $ docker pull tfgco/donations
    $ docker run -t --rm -e "DONATIONS_MONGO_HOST=<mongoDB host>" -e "DONATIONS_MONGO_PORT=<mongoDB port>" -p 8080:8080 tfgco/donations
```

## Binaries

Whenever we publish a new version of Donations, we'll always supply binaries for both Linux and Darwin, on i386 and x86_64 architectures. If you'd rather run your own servers instead of containers, just use the binaries that match your platform and architecture.

The API server is the `donations` binary. It takes a configuration yaml file that specifies the connection to MongoDB and some additional parameters. You can learn more about it at [default.yaml](https://github.com/topfreegames/donations/blob/master/config/default.yaml).

## Source

Left as an exercise to the reader.
