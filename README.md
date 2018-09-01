# gRPC Web Crawler

The application consists of a command line client and a local service
which does the actual web crawling. Client and server communicate via gRPC[1].

The client supplies one or more URLs to the crawler, which creates a "site
tree" -- a tree of links wuth the root of the tree being the supplied root
URL.

The crawler will not follow links off the domain of the root URL, but will 
record those offsite links.

# Client

The client provides the following operations:

 - crawl boot
  - Starts up the crawler service. Does nothing if the service is running.
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
 - crawl shutdown
  - Stops the crawl server, terminating all crawling.

