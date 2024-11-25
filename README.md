# Technical test: Go developer
```
The following test is to be implemented in Go and while you can take as much time as you need,
it's not expected that you spend more than 3 or 4 hours on it.
```

This exercise was a very good learning experience for me, as I'd not used protobuf or gRPC, and we had not used channels and semaphores ~~much~~ at all at WhiteHat. I spent quite a bit of time learning how to use them in this context, and if nothing else, I certainly learned a lot of new idioms and techniques.

```
The test consists of implementing a "Web Crawler as a gRPC service". The application consists of
a command line client and a local service which runs the actual web crawling. The communication
between client and server should be defined as a gRPC service (*). For each URL, the Web Crawler,
creates a "site tree", which is a tree of links with the root of the tree being the root URL. The 
crawler should only follow links on the domain of the provided URL and not follow external links.
Bonus points for making it as fast as possible.
```

Notes on the implementation below.

# gRPC Web Crawl Server

```
    server [-tls --tls_cert_file=<cert> --tls_key_file=<key>]
           [--port=10000] [-debug] [-mock]
```

If TLS is to be used, all three of the TLS items (`tls`, `tls_cert_file', `tls_key_file`) must be supplied. Port defaults to 10000 unless otherwise specified. (2024 followup note: this was before Let's Encrypt was easy to use, so I was doing this all by hand. I'd certainly use it now.)

If `debug` is supplied, the server's debug logging is enabled. (Setting the `TESTING` environment variable to a non-null value also enables debugging.)

If `mock` is supplied, the server uses the mock URL fetcher for all URL operations. This can be useful if you need to verify operation of the crawler without actually crawling any sites.

The application consists of a command line client and a local service
which does the actual web crawling. Client and server communicate via gRPC[1].

The client supplies one or more URLs to the crawler, which creates a "site
tree" -- a tree of links with the root of the tree being the supplied root
URL.

The crawler will not follow links off the domain of the root URL, but will 
record those offsite links. Links that can't be followed or parsed will be
recorded .

# gRPC CLI Client

The client provides the following operations:

 - `crawl start www.example.com`
  - Starts crawling at `www.example.com`, only following links on `example.com`.
 - `crawl stop www.example.com`
  - Stops crawling of `example.com`.
 - `crawl status www.example.com`
  - Shows the crawl status for the supplied URL.
 - `crawl show` 
  - Displays the crawled URLs as a tree structure.

The CLI uses the `Cobra` CLI library, allowing us to have a CLI similar to Docker or Kubernetes.

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
./crawl status <url>   # Shows status of the URL crawl.
./crawl show <url>     $ Displays a tree representation of the crawled URLs.
```

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
 - [colly](http://go-colly.org/)
   Web scraper library

# Notes on the implementation

The basic architecture of the crawler is based on the one in the Go Tour; I've switched up the recursion limit check to be an off-domain check. Integrating `gotree` to record the structure made the process of creating the site tree very simple.

Having the `Fetcher` be an interface made it much easier to work through the implementation process, and made it very easy to swap implementations at runtime for the server. 

`Colly` made the actual web fetch and parse very simple, and I plan to use it again in later projects.

# Notes to readers in 2024

This test is six years old at this point, but is not-terrible enough that I wanted to post it as an example of
a non-trivial use of gRPC, channels, and the like. I found it actually of some use in verifying static sites.

I was ghosted post submission. 

There was no NDA, so I feel free to post the code at this point. I was a relatively
inexperienced Go programmer at the time -- our microservices at WhiteHat were all basically HTTP servers 
(not even HTTPS!) that ran SQL queries and either updated the database or returned JSON -- so it was a 
_massive_ effort to learn all the new technologies in a few days, which I did in good faith; at the time, 
I was mystified and not a little disappointed to hear nothing back in response. Now, in 2024, I see it was 
just a foreshadowing of recruiting processes to come.

For other software engineers, remember that the interview process tells _you_ about the employer
as well as telling them about you: uncompensated take-home tests tell you that their attitude is that your
time isn't valuable; that your free time is theirs to put claims on.

Getting ghosted also tells you that you are not respected, even to the point of
providing the minimum "thanks, appreciate your effort here, but we've decided not to move forward". 

I would note that I had at the time been evaluating the company's product to see if it was going to work for me as 
a Dropbox alternative. I decided to not "move forward" too, deleted it, and haven't looked at it since.
