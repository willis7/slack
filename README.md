# Slack

This project started as an embedded package in my Golossary project. After a positive code review from Brian Ketelsen I decided to open source this code as a standalone package.

## Design

This package was heavily influenced by the `net/http` package which allows users to define a `Handler` for a given pattern. The same principle is used here for events:

``` go
    mux := slack.NewEventMux()
    mux.Handle("message", slack.HandlerFunc(RTMMessage))
```

In the example above we use the user defined `RTMMessage` function, which satisfies the `Handler` interface with the `message` event. Similarly, we could apply our  `RTMHello` function to the `hello` event. This might look like:

``` go
    ...
    mux.Handle("hello", slack.HandlerFunc(RTMHello))
```

## Client Lifecycle

### Connecting to Slack

To create a new client we first need to run:

`client := slack.NewClient(token, mux)`

You're required to provide an `EventMux` here because it would be a useless client if it didn't know how to handle events. No default Event mux is provided.

To open a connection with Slack you will need to call:
``` go
client.Connect()
defer client.Close()
```

Remeber to close the connection! 

Next we need to tell the client to start dispatching messages with:

`go client.Dispatch()`

At this point its really up to the user to shutdown the connection gracefully. Here's an example of how I do it:

```go
	sigterm := make(chan os.Signal)
	signal.Notify(sigterm, os.Interrupt)
	for {
		select {
		case <-sigterm:
			log.Println("terminate signal recvd")
			err := client.Shutdown()
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-time.After(time.Second):
			}
			client.Close()
			return
		}
	}
```