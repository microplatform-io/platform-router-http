// Package main ...
package main

import (
	"encoding/hex"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/microplatform-io/platform"
)

func getDefaultPublisher(connectionManager *platform.AmqpConnectionManager) platform.Publisher {
	publisher, err := platform.NewAmqpPublisher(connectionManager)
	if err != nil {
		log.Fatalf("Could not create publisher. %s", err)
	}

	return publisher
}

func getDefaultSubscriber(connectionManager *platform.AmqpConnectionManager, queue string) platform.Subscriber {
	subscriber, err := platform.NewAmqpSubscriber(connectionManager, queue)
	if err != nil {
		log.Fatalf("Could not create subscriber. %s", err)
	}

	return subscriber
}

func manageRouterState(routerConfigList *platform.RouterConfigList) {
	//routerConfigListBytes, _ := platform.Marshal(routerConfigList)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Emit a router offline signal if we catch an interrupt
	go func() {
		select {
		case <-sigc:
			//publisher.Publish("router.offline", routerConfigListBytes)
			os.Exit(0)
		}
	}()

	// Wait for the servers to come online, and then repeat the router.online every 30 seconds
	time.AfterFunc(10*time.Second, func() {
		//publisher.Publish("router.online", routerConfigListBytes)

		for {
			time.Sleep(30 * time.Second)
			//publisher.Publish("router.online", routerConfigListBytes)
		}
	})
}

// GetProtoBytes - Will return platform message as bytes
func GetProtoBytes(message platform.Message) []byte {
	protoBytes, _ := platform.Marshal(message)
	return protoBytes
}

// ErrorResponse - Used by handlers so we can return back errors
func ErrorResponse(msg string) []byte {
	return []byte(hex.EncodeToString(GetProtoBytes(&platform.Request{
		Routing:   platform.RouteToUri("resource:///router-http/reply/error"),
		Payload:   GetProtoBytes(&platform.Error{Message: platform.String(msg)}),
		Completed: platform.Bool(true),
	})))
}
