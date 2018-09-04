# gRPC Web Crawler

The application consists of a command line client and a local service
which does the actual web crawling. Client and server communicate via gRPC[1].

The client supplies one or more URLs to the crawler, which creates a "site
tree" -- a tree of links with the root of the tree being the supplied root
URL.

The crawler will not follow links off the domain of the root URL, but will 
record those offsite links.

# Client

The client provides the following operations:

 - crawl start www.example.com [-trace]
  - Starts crawling at www.example.com, only following links on example.com.
    If -trace is supplied, then the crawler dumps status information for each
    URL it visits as it runs.
 - crawl stop www.example.com
  - Stops crawling of example.com.
 - crawl status [-site www.example.com] [-all]
   - Shows how many crawls are in progress and summarizes their current
     status. If `-site` is supplied, we dump the site tree for that site.
     If `-all` is supplied, we dump the site tree for all sites scanned or
     being scanned.

# Building it

All of the external dependencies are in the `dep` configuration; install `dep`
and run `dep ensure` to install them. `make` will build the CLI client and the
server.

# Testing it

The tests use `ginkgo`, so you will need to install it:

```bash
go get -u github.com/onsi/ginkgo/ginkgo  # installs the ginkgo CLI
go get -u github.com/onsi/gomega/...     # fetches the matcher library
```

This will be installed in your `$GOPATH/bin`; if that's not in your `PATH`,
you can run the tests with `$GOPATH/bin/ginkgo -r`.

# Running it

`make run` will build the client and server, and also kill any old server
and start a new one for you.

```
./crawl start <url>    # Starts up a new crawl
./crawl stop <url>     # Pauses a crawl
./crawl status <url>   # Prints a tree representation of the URLs crawled.
```

# Notes

The version at `HEAD` is still incomplete; the server runs, and the CLI can
talk to it, but the server does not yet actually crawl anything. The crawler
works, but currently can't be started and stopped. 

# External dependencies of note

 - [github.com/disiqueira/gotree](https://github.com/disiqueira/gotree)
   tree recording and formating
 - [github.com/sirupsen/logrus](https://github.com/sirupsen/logrus)
   Logging
 - [github.com/spf13/cobra](https://github.com/spf13/cobra)
   CLI base library and code generator
 - [google.golang.org/grpc](https://google.golang.org/grpc)
   Client/server communications
 - [github.com/joemcmahon/logcap](https://github.com/joemcmahon/logcap)
   Forked version of logcap compatible with current versions of `logrus`
 - [dep](https://github.com/golang/dep)
   Dependency management

