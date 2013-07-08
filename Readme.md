This is just an experimental project to cover T Dispatch's Auto Dispatch service,
written in Go language and using the Fleet API to request available drivers,
bookings picking up soon and to dispatch bookings.

## Setup

1. You first need to install **Go lang**
1. You must have MongoDB installed with a T Dispatch database

## How to run it in Linux or MacOSX

1. Run the environment with

```
. envrc
```

1. If this is the first time, you have to install packages:

```
go get labix.org/v2/mgo
go install github.com/TDispatch/fleet-api
go install github.com/TDispatch/auto-dispatch
```

1. Run the bot with:

```
go run auto_dispatch_bot.go
```

1. To compile and run the binary, you have to install "github.com/TDispatch/auto-dispatch" and run:

```
./bin/auto-dispatch
```
