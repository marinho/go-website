This is just an experimental project of a simple CMS system written in Go language
and AngularJS framework. The main goal with this project is to explore Angular JS
to understand how it works.

## Setup

1. You first need to install **Go lang**
1. You must have MongoDB installed

## How to run it in Linux or MacOSX

1. Run the environment with

```
. envrc
```

1. If this is the first time, you have to install packages:

```
go get labix.org/v2/mgo
go get github.com/gorilla/mux
go get github.com/gorilla/sessions
```

1. Run the bot with:

```
go run src/github.com/marinho/go-website/server.go
```

1. To compile and run the binary, you have to install "github.com/marinho/go-website" and run:

```
./bin/server
```

## To do

1. Image upload tool
1. Image thumbnail function
1. Template editor and uploader
1. Better configuration tools
1. Fixtures loader command

