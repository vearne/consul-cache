# consul-cache
consul-cache is a local cache disguised as consul agent and has stronger performance.

## Purpose
In the consul cluster, when the app initiates a watch request, 
the consul agent will forward the request to the consul server, 
which will put a huge burden on the consul server. 
The purpose of consul-cache is to separate this burden from the consul server.

## Architecture
consul-cache has 2 components.
### fetcher
Discover service changes from consul and build all instances corresponding to the service in Redis
### cache
cache disguises itself as a consul agent to facilitate service discovery by business programs


![Architecture](./img/arch.png) 