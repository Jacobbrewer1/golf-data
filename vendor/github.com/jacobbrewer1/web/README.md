# Web

## Description

This repository contains tools that are commonly used between my golang projects. It allows for the projects to be
easily structured and maintained. The tools are designed to be used in a modular way, allowing for easy integration into
existing projects. The repository is easily extensible, allowing for new tools to be added as needed. When creating a
new web app, ensure you provide a logger; the logger can be accessed later in through the application.

## Usage

### Config

You can access the config file using the viper instance from the app. The config file is loaded from the Environment
Variable `CONFIG_LOCATION`; if no location is provided, the app will default to `config.json` in the present directory.
The config file can be in any format that viper supports. Before you can access it you will have to ensure that the app
loads the config with the `WithViperConfig` option. Once the config is loaded you can access it using the `app.Viper()`
method.

### Example

Please see a basic example of how to use the tools in a web app. The example is a simple web app that provides a Vault
client, database connection from Vault and a Http server.

```go
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jacobbrewer1/web"
	"github.com/jacobbrewer1/web/logging"
)

type App struct {
	base *web.App
}

func NewApp(l *slog.Logger) (*App, error) {
	base, err := web.NewApp(l)
	if err != nil {
		return nil, err
	}

	return &App{
		base: base,
	}, nil
}

func handler(l *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l.Debug("Received request", "method", r.Method, "url", r.URL.String())
		w.WriteHeader(http.StatusOK)
	})
}

func (a *App) Start() error {
	if err := a.base.Start(
		web.WithViperConfig(),
		web.WithVaultClient(),
		web.WithDatabaseFromVault(),
	); err != nil {
		return fmt.Errorf("start app: %w", err)
	}

	if err := a.base.StartServer("api", &http.Server{
		Addr:              ":8080",
		Handler:           handler(logging.LoggerWithComponent(a.base.Logger(), "api-handler")),
		ReadHeaderTimeout: 5 * time.Second,
	}); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	return nil
}

func (a *App) WaitForEnd() {
	// Wait for the application to finish
	a.base.WaitForEnd(a.Shutdown)
}

func (a *App) Shutdown() {
	a.base.Shutdown()
}

func main() {
	l := logging.NewLogger(
		logging.WithAppName("myapp"),
	)

	app, err := NewApp(l)
	if err != nil {
		panic(err)
	}

	if err := app.Start(); err != nil {
		panic(err)
	}

	app.WaitForEnd()
}

```

The config file for the above example is a simple JSON file that contains the following:

```json
{
  "vault": {
    "address": "http://vault-active.vault:8200",
    "database": {
      "path": "database-mount/creds",
      "role": "vault-database-role"
    }
  },
  "database": {
    "host": "host:port",
    "schema": "schema-name"
  }
}
```

### Metrics

If you want to expose metrics for your application, these are registered by default. You can use the
`web.WithMetricsEnabled` option to enable or disable the metrics. The metrics are exposed on the `/metrics` endpoint on
port `9090`. The metrics are registered with the `prometheus` package and are exposed in the `prometheus` format.

### Dynamic Health Checks

If you want to add health checks that state whether the application is healthy or not, you can use the
`web.WithHealthCheck` option. This option will take a list of health checks that will be run whenever the health check
endpoint is hit. There is **_NO_** specific endpoint for the health check, it is just the port that is exposed. If you
provide a path, it is re-routed to the `/` endpoint.

### Options

Please take a look at the `options.go` file for the available options. The options are used to configure the application
and are passed to the `app.Start()` method. When the application is started, the options are applied in the order they
are passed, so the order is important. For example, `WithVaultClient` uses the viper config so it should be passed after
the `WithViperConfig` option.

### Async Tasks

If your application is not using a server, you may be watching some configs for example, or fetching data and placing it
into a stream, you can use the `web.WithIndefiniteAsyncTask("name", func(ctx) {})` option to run a task that should not
exit until the application is stopped. If the task exits, the application will stop. The task will be run in a goroutine
and will be passed a context that will be cancelled when the application is stopped. The application will wait for the
task to finish before stopping.

### Deployment

When deploying the application, you will need to ensure that the following environment variables are set for a smooth
operation, this application makes use of the Downward Kubernetes API to set the environment variables.

The following environment variables should be set within the deployment yaml:

```yaml
- name: "SERVICE_ACCOUNT_NAME" # This is used to get the service account name; additionally, Vault will fail to initialize if this is not set.
  valueFrom:
    fieldRef:
      fieldPath: spec.serviceAccountName
- name: "POD_IP" # This is used to get the pod IP address
  valueFrom:
    fieldRef:
      fieldPath: status.podIP
- name: "NODE_NAME" # This is used to get the node name
  valueFrom:
    fieldRef:
      fieldPath: spec.nodeName
```

### Packages

If you require some Kubernetes helpers, you can use the `k8s` package. It contains a few helpers that are useful for
interacting with Kubernetes. Please look at the files within the `k8s` package for more information.

## Local Development

While the Web module is primarily designed to be run in Kubernetes, it can also be run locally (or without a cluster, in
a VM depending on your setup).

### Vault Override

If you want to run the application outside of Kubernetes, you will need to override the `VaultClient()` method that is
called to get the vaulty client. This method has been kept as a variable so that it can be overridden in this scenario.

Please see the example below for how to override the `VaultClient()` method. This will allow you to run the application

```go
package main

import (
	"context"
	"log/slog"

	"github.com/jacobbrewer1/vaulty"
	"github.com/jacobbrewer1/web"
	"github.com/spf13/viper"
)

func main() {
	web.VaultClient = func(ctx context.Context, l *slog.Logger, v *viper.Viper) (vaulty.Client, error) {
		vault, err := vaulty.NewClient(vaulty.WithLogger(l)) // Be sure to pass in the authentication method, etc.
		if err != nil {
			return nil, err
		}

		return vault, nil
	}
}

```

## Installation

```shell
go get github.com/jacobbrewer1/web
```