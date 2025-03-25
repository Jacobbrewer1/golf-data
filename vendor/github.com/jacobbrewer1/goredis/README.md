# GO Redis

Golang Redis Pool Implementation. This is a simple implementation of a Redis Pool in Golang. It is a simple wrapper
around the `github.com/gomodule/redigo/redis` package.

## Installation

To install the package, run the following command:

```bash
go get -u github.com/jbrewer/goredis
```

## Basic Usage

Please find the basic usage example below:

```go
package main

import (
	"github.com/gomodule/redigo/redis"
	goredis "github.com/jacobbrewer1/goredis/redis"
)

func main() {
	if err := goredis.NewPool(
		goredis.WithDefaultPool(),
		goredis.WithAddress("localhost:6379"),
		goredis.WithNetwork(goredis.NetworkTCP),
		goredis.WithDialOpts(
			redis.DialDatabase(0),
			redis.DialUsername("username"),
			redis.DialPassword("password"),
		),
	); err != nil {
		panic(err)
	}
}
```

## Usage

Please find our usage examples in the `examples` [directory](./examples).

If you are using a JSON configuration file, you can pass in a viper instance `goredis.FromViper(viperInstance)`
connection option. This connection option will allow you to pass in a viper instance to read your configuration for the
basic authentication and database to connect to, with the addition of the MaxIdle, MaxActive, IdleTimeout and Address.

Please find an example of a JSON configuration file below:

```json
{
  "redis": {
    "address": "localhost:6379",
    "password": "password",
    "username": "username",
    "db": 0,
    "max_idle": 3,
    "max_active": 5,
    "idle_timeout_secs": 240
  }
}
```

You are able to pass additional options with the Connection option `goredis.WithDialOpts`. This option allows you to
pass in a slice of `redis.DialOption` to the connection pool. This is useful for setting the read and write timeout.
