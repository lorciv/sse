# HTTP Server-Sent Events

Sse is a Go package that implents HTTP server-sent events (SSE) handling.
The standard for HTTP SSE can be found here: [https://html.spec.whatwg.org/multipage/server-sent-events.html#server-sent-events]().

```sh
$ go get github.com/lorciv/sse
```

## Usage

The package exposes a function `NewStream` that can be used to instantiate a `Stream`.

A Stream is a http.Handler, so it can be registered on a http.ServeMux.
Clients that issue a GET request to the given path will receive events as they are sent by the server.

```go
s := sse.NewStream()
http.Handle("/stream", s)
```

To send events on a stream, use the `Send` method.
Send accepts the data to be sent as a slice of bytes.

```go
s.Send([]byte("42"))
```

If you wish to specify the event type, use `SendEvent` instead.
(A call to Send is equivalent to SendEvent with event type == "message".)

```go
s.SendEvent("apples", []byte("42"))
```

Arbitrarily complex data can be sent to the client as long as it is encoded as a byte slice.
For example, structs can be encoded to json on the server using `json.Marshal` and then sent to the client.

## Client example

Clients receive events as they are sent by the server.

For example, to implement Javascript client that runs in a browser, one can use use the EventSource object.
The following code listens to the stream at url "/stream" (the one shown above) and prints the data on console as it arrives:

```js
const stream = new EventSource('/stream');
stream.onmessage = function(m) {
    console.log(m.data);
};
```

## Other functionalities

A stream can optionally be assigned a logger.

```go
s.Logger = log.Default()
```
