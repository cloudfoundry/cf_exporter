# go-cf-clients-helper

Convenient way to have all cloud foundry go client api from https://github.com/cloudfoundry/cli .

The Cloud foundry cli provide go client to call cloud foundry api like v2/v3 api, uaa api,
network policy api, logs and metrics api, tcp router api ...
But, how to build and call them can be very difficult and is not documented.

This repo leverage this situation and is an extract of the code we do for [cloud foundry terraform provider](https://github.com/cloudfoundry-community/terraform-provider-cf).

## Usage

```go
package main

import (
	"github.com/cloudfoundry-community/go-cf-clients-helper" // package name is clients
	"fmt"
)

func main() {
	config := clients.Config{
		Endpoint: "https://api.my.cloudfoundry.com",
		User: "admin",
		Password: "password",
	}
	session, err := clients.NewSession(config)
	if err != nil {
		panic(err)
	}

	// get access to cf api v2 (incomplete api)
	session.V2()

	// get access to cf api v3 (complete and always up to date api)
	session.V3()

	// Get access to api uaa (incomplete)
	session.UAA()

	// Get access to TCP Routing api
	session.TCPRouter()

	// Get access to networking policy api
	session.Networking()

	// Get an http client which pass authorization header to call api(s) directly
	session.Raw()

	// Get config store for client which need, for example, current access token (e.g.: NOAA)
	session.ConfigStore()
}
```
