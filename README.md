# HTTP Server-Side Events

`sse` is a Go package that implents HTTP server-side events handling.

```sh
$ go get github.com/lorciv/sse
```

## Usage

The main element is `Stream`, a struct that implements the http.Handler interface.

To create a stream use `NewStream` and then register it for a path.

```go
s := sse.NewStream()
s.Logger = log.Default() // optionally set the logger
http.Handle("/stream", s)
```

Clients that issue a GET request to the given path will receive all new events.

Use the `Send` method to send events.

```go
s.Send("42")
```

If you wish to specify the event type, use `SendEvent` instead.

```go
s.SendEvent("count", "42")
```
