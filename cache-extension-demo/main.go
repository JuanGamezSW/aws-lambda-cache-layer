// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package main

import (
	"cache-extension-demo/extension"
	"cache-extension-demo/ipc"
	"cache-extension-demo/plugins"
	"context"
	"os"
	"os/signal"
	"syscall"
)

var (
	extensionClient = extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "debug" {
		mainDebug()
	} else {
		mainLambda()
	}
}

func mainDebug() {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		println(plugins.PrintPrefix, "Received", s)
		println(plugins.PrintPrefix, "Exiting")
	}()
	println(plugins.PrintPrefix, "Begin register...")

	// Initialize all the cache plugins
	extension.InitCacheExtensions()

	// Start HTTP server
	ipc.Start("4000")

	// Will block until shutdown event is received or cancelled via the context.
	processEventsDebug(ctx)
}

func mainLambda() {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		println(plugins.PrintPrefix, "Received", s)
		println(plugins.PrintPrefix, "Exiting")
	}()
	println(plugins.PrintPrefix, "Begin register...")
	_, err := extensionClient.Register(ctx, plugins.ExtensionName)
	if err != nil {
		println(plugins.PrintPrefix, "Error registering: "+err.Error())
		panic(err)
	}

	// Initialize all the cache plugins
	extension.InitCacheExtensions()

	// Start HTTP server
	ipc.Start("4000")

	// Will block until shutdown event is received or cancelled via the context.
	processEvents(ctx)
}

// Method to process events
func processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			res, err := extensionClient.NextEvent(ctx)
			if err != nil {
				println(plugins.PrintPrefix, "Error:", err)
				println(plugins.PrintPrefix, "Exiting")
				return
			}

			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				return
			}
		}
	}
}

// Method to process events debug mode
func processEventsDebug(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}
