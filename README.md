# GraphMetrics
[![Go Reference](https://pkg.go.dev/badge/github.com/graphmetrics/graphmetrics-go.svg)](https://pkg.go.dev/github.com/graphmetrics/graphmetrics-go)
[![Go Version](https://img.shields.io/github/go-mod/go-version/graphmetrics/graphmetrics-go)](https://github.com/GraphMetrics/graphmetrics-go)
[![Go Report](https://goreportcard.com/badge/github.com/GraphMetrics/graphmetrics-go)](https://goreportcard.com/report/github.com/GraphMetrics/graphmetrics-go)

This is the Go SDK for GraphMetrics.

## Usage
We provide middlewares that are easily to plug in your server. 
If your server is not currently supported, please open an issue so we can fix that.

### gqlgen
```go
import (
    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/graphmetrics/graphmetrics-go"
    graphmetricsgqlgen "github.com/graphmetrics/graphmetrics-go/gqlgen"
)
var srv *handler.Server

gm := graphmetricsgqlgen.NewExtension(&graphmetrics.Configuration{ 
	// SEE CONFIGURATION SECTION
})
defer gm.Close() // Keep a reference to the extension and call Close on server shutdown

srv.Use(gm)
```

## Configuration
The SDK needs a few elements to be properly configured using the `graphmetrics.Configuration`.

- `ApiKey`: Your environment api key
- `ServerVersion`: (Optional) The version of the server, necessary to catch regressions between releases
- `ClientExtractor`: (Optional) Function that retrieves the client details from the context, necessary to differentiate queries coming from different clients
- `Logger`: (Optional) A structure logger that respects the interface, otherwise golang "log" is used. Adapters are provided for popular logger, see the [logger-go package](https://github.com/GraphMetrics/logger-go).

### Client extractor

The client extractor fetches the client details from the context. By default, no details are fetched.
We provide helper functions for the Apollo client, please let us know if you would like to see other clients supported.

The first step of the extraction is to add the `http` middleware to your server. 
The exact way of doing that depends on your server implementation, but generally looks like:
```go
import (
    "github.com/go-chi/chi"
    "github.com/graphmetrics/graphmetrics-go/client"
)

r := chi.NewRouter()

r.Use(client.ApolloMiddleware)
```

Then you can pass the extractor function in the configuration:
```go
graphmetrics.Configuration{
    ClientExtractor: client.ApolloExtractor,
}
```

### Advanced configuration

- `FieldBufferSize`: As we do not want to slow down your queries, we process the field metrics async in a goroutine with a buffered channel in between. 
Usually this buffer is big enough to handle spikes, but it might not be if you have very large and fast queries. 
In which case, metrics are dropped and a warning is emitted. Please contact us if that happens and try increasing the buffer in the meantime. 
