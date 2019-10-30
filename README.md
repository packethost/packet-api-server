# Packet API Server

This repository contains a library and command-line of a subset of the Packet API Server. At this point, it is **not** intended to be
a full-fledged replacement server. Its primary purpose is a reusable server to test various clients that speak to the Packet API.

Eventually, it may become a full-fledged server.

It also builds in a metadata handler, for convenience.

## Using

1. Create a `DataStore` that will handle all of the requests.
1. Create an `ErrorHandler` that will handle any errors. It is your choice as to whether it should exit immediately as fatal, or return something else. This makes it equally easy to use in tests or long-running systems.
1. Optionally, set the device ID for which metadata will return data
1. import and instantiate
1. Enjoy!

For your convenience, a memory store is included. It allows you to modify the items directly
in the `struct`, or to use convenience functions to seed items.

```go
import (
    "github.com/packethost/packet-api-server/pkg/server"
    "github.com/packethost/packet-api-server/pkg/store"
)

type errorHandler struct {}
func (e *errorHandler) Error(err error) {
    fmt.Println(err)
    os.Exit(1)
}

server := &server.PacketServer{
    Store:        store.NewMemory(), // we are using the in-memory handler
    ErrorHandler: &errorHandler{},
    MetadataDevice: "abcd156", // the device that metadata will answer for
}

srv := &http.Server{
    Handler:      server.CreateHandler(),
    Addr:         "127.0.0.1:8000",
}

srv.ListenAndServe()
```

