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

Clients that issue a GET request to the given path will receive events.

Use the `Send` method to send events. Send accepts data as a slice of bytes.

```go
s.Send([]byte("42"))
```

If you wish to specify the event type, use `SendEvent` instead.

```go
s.SendEvent("count", []byte("42"))
```

Arbitrarily complex data can be sent to the client as long as it is encoded as a byte slice.
For example, structs can be encoded to json on the server (using Go's `json.Marshal`) and then decoded back to objects on the client (e.g., using Javascript's `JSON.parse(...)`).

## Client

On a Javascript client, use the EventSource object to listen to events.
For example, the following code listens to the stream shown above and prints the data on console:

```js
const stream = new EventSource('/stream');
stream.onmessage = function(m) {
    console.log(m.data);
};
```

