# Message Forwarder

This system contains two components: 

* Streamer: The receiver of relevant messages from a platform that will poll/subscribe to incomming messages and write them to a Kafka topic.
The example implementation publishes a simple message with the time periodically.
* Client: A websocket server that reads these messages and displays messages on a webpage in real time.

To run, first run Kafka and Zookeeper using Docker compose:
```shell
make kafkazk
```

And to run streamer and client:
```shell
make streamer
make client
```

When the client app is ready, you can access the webpage on port 8080: 
http://localhost:8080