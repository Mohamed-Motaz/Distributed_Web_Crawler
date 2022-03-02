# **Distributed Web Crawler**
![systemArchi drawio](https://user-images.githubusercontent.com/53558209/155814673-c201500d-7f48-4223-82a3-7bf9b7633190.png)
The main objective of the Distributed Web Crawler is to serve as a template for a system that can parallelize tasks and distribute the load, be it cpu and processing load, or even network load, as is the case for this Web Crawler. The system is [**distributed**](#system-components) on multiple machines, [**highly available**](#high-availability-and-reliability), [**reliable**](#high-availability-and-reliability) and [**fault tolerant**](#fault-tolerance). 

## **Table of Contents**
- [**A Journey Across the System**](#a-journey-across-the-system)
- [**System Components**](#system-components)
    * [Load Balancer](#load-balancer)
    * [WebSocket Server](#websocket-server)
    * [Cache](#cache)
    * [Message Queue](#message-queue)
    * [Master](#master)
    * [Worker](#worker)
    * [Lock Server](#lock-server)
    * [Database](#database)

- [**Availability And Reliability**](#high-availability-and-reliability)
- [**Fault Tolerance**](#fault-tolerance)
- [**Further Optimizations**](#further-optimizations)
    * [WebSocket Optimizations](#websocket-optimizations)
    * [Load Balancing Optimizations](#load-balancing-optimizations)
    * [General Optimizations](#general-optimizations)

- [**Faults (Yup, and many of them)**](#faults)
- [**How To Run**](#how-to-run)


## **A Journey Across the System**
 This section is meant to establish the journey of a job request from the moment the client requests it, up to the moment the results are delivered back to the client. The journey is as follows.
 - The client attempts to establishe a websocket with the one of the websocket servers.
 - The connection passes through the load balancer which then chooses one of the websocket servers to forward the websocket connection to, using the Round Robin algorithm.
 - The client sends a job request, containing the url to crawl, and the depth required to crawl.
 - After the websocket server assigned to said client receives the request, it first checks if a similar request is present in cache, and if there is, the result is sent back to the client immediately.
 - Assuming no results were present in cache, the websocket server then pushes the job to the Assigned Jobs Queue.
 - One of the multiple Masters then pulls the job from the queue.
 - Before starting to process the job, the Master first asks the Lock Server for permission to start the job.
 - If the Lock Server doesn't have any jobs with a higher priority (ie. late jobs with dead masters), it allows the master to start processing said job.
 - The Master starts coordinating its assigned Workers, in order to finish the job as quickly as possible.
 - After all the results have been collected, the Master then pushes the job into the Done Jobs Queue.
 - One of the websocket servers then pulls the job from the Done Jobs Queue, and if the job belongs to a client of said websocket server, the websocket server adds the results to cache.
 - Afterwards, the results are sent back to the appropriate client.


## **System Components**

- ### **Load Balancer**
    The load balancer of choice is HaProxy. The following highlights its main functionalities:
    - Be able to establish and maintain websocket connections between the client and the websocket servers.
    - Handle up to 50,000 (variable) concurrent connections at a time.
    
    Why I chose HaProxy:
    - The main reason I chose HaProxy over Nginx is because HaProxy fits my needs as a load balancer perfectly. Nginx would be overkill, and HaProxy uses Round Robin, which in my case, makes sense, since I want the websocket connections to be balanced amongst all websocket servers. It also supports websockets out of the box, so it was a perfect fit.


- ### **WebSocket Server**
    The client facing servers use websockets to communicate with their clients. The following highlights their main functionalities:
    - Responsible for establishing and mainting active websockets with the clients.
    - Responsible for cleaning up and closing all connections that have been idle for over a (variable) amount of time.
    - Publish jobs to the Assigned Jobs Queue if no results could be found in the cache.
    - Consume done jobs from the Done Jobs Queue and send them over to the clients.
    - Keep the cache up to date, and reset the TTL when a job's result is found in cache.
    
    Why I chose websockets:
    - The main reason I chose websockets is well, because they are trendy! Obviously not just that, but I had 2 other choices, polling every 5 seconds or so with normal HTTP request-response, or use [Server-Sent Events](https://en.wikipedia.org/wiki/Server-sent_events#:~:text=Server%2DSent%20Events%20(SSE),client%20connection%20has%20been%20established.). Both would have been fine, but that is only because in my system, client requests usually take a few seconds up to a few minutes to complete, so polling wouldn't cause much overhead. I decided to stick with websockets though since I wanted to go with a more general solution (in case requests actually only do take a few hundered milliseconds to be processed), and avoid having to constantly poll the server, which in some cases would actually cause more overhead than just maintaining one TCP connection over the client's lifecycle.


- ### **Cache**
    The cache of choice is Redis. The following highlights its main functionalities:
    - Serve as a key value store, where each key has a set [TTL](https://en.wikipedia.org/wiki/Time_to_live)
    - Each key is a url, and its value contains the depth crawled, and the crawled websites 2D array.
    - Eg. If a client asks for "google.com", with a depth of 2, then the cache must contain "google.com" with atleast a depth of 2, so the client can be served immediately without any additional delay.

    Why I chose Redis:
    - The main reason I chose Redis is because it supports clustering and replication, (I can implement it in the future), and it seems like a fairly popular choice, so why not?


- ### **Message Queue**
    The message queue of choice is RabbitMq. The following highlights its main functionalities:
    - Durable, so in case of a crash, "all" jobs in the queue can be restored from disk.
    - Support multiple producers and consumers per queue.
    - Jobs Assigned Queue Producers:       Websocket Servers
    - Jobs Assigned Queue Consumers:       Masters
    - Done Jobs Queue Producers:           Masters
    - Done Jobs Queue Consumers:           Websocker Servers
   
   Why I chose RabbitMq:
    - The main reason I chose RabbitMq is because it (also) supports clustering and replication, (I can implement it in the future). In addition, it uses the push model, and the consumers can set a prefetch limit (which I set to 1), so that they avoid getting overwhelmed. This helps in achieving low latency and maximal performance.


- ### **Master**
    Masters are the main job consumers. The following highlights their main functionalities:
    - Consume jobs from the Jobs Assigned Queue
    - Ask the Lock Server for permission to start said job, and accept if the Lock Server provides them with a different job
    - Communicate with the workers, and orchestrate the workload among them
    - Keep track of the job progress at all times, and notify the Lock Server when they are done
    - Push done jobs to the Done Jobs Queue so that the Websocket servers can consume them and send the results to the waiting clients.


- ### **Worker**
    Workers are the powerhouses of the system. They are completely stateless, and only know about their master's address. The following highlights their main functionalities:
    - Communicate with the masters, ask for jobs, and respond with the results.
    

- ### **Lock Server**
    The Lock Server is the server tasked with persisting all system jobs in the database, so that in case of failure, all jobs can still be recovered. To avoid being a single point of failure, which it is, it should be implemented on top of a [Raft cluster](https://en.wikipedia.org/wiki/Raft_(algorithm)). The following highlights its main functionalities:
    - Make quick decisions on whether a master should start a job, or if there is any higher priority job to be given.
    - Keep track of all current jobs, and re-assign any jobs that are delayed beyond a (variable) time.
    - Persist all jobs' information and status to the database.
    - Be extremely performant, since every single jobs needs to pass by the Lock Server before it can be processed.


- ### **Database**
    The database of choice is PostgreSQL. The following highlights its main functionalities:
    - Persist data in case masters die, thats it. I bet you didn't expect much to be honest.
    
    Why I chose PostgreSQL:
    - The main reason I chose PostgreSQL is because it supports clustering (not natively) and replication, (I can implement it in the future). But I mean, everything supports clustering and replication nowadays, so I really just wanted to try it out.

## **High Availability And Reliability**
- To the client, the system has really high availability. The only case where the system would be down is if all the load balancers, websocket servers, or message queues are down. Since all of these are/can be implemented in clusters and replicated, the system is indeed highly available.
- The system is highly reliable and consistent, since it doesn't rely on some kind of consensus among all of its masters, or workers, or any of the componenents. Each component is completely stateless, and the end result is a reliable system that would deliver the same result every time a user requests a job.

## **Fault Tolerance**
The system is designed with fault tolerance in mind. The system is able to handle the following types of faults:
- Master failures
- Worker failures
- WebSocket server failures
- Cache failure

All the above components can fail, and the system keeps running without a hitch. The failures that do affect the system are:
- Load Balancer failure: Could have a load balancer cluster to cover for the failure of a machine, and when the clients re-establish
the connections, they should use [exponential backoff](https://en.wikipedia.org/wiki/Exponential_backoff), so as not to cause a [thundering herd problem](https://en.wikipedia.org/wiki/Thundering_herd_problem).
- Message Queue failure: RabbitMq could be replicated and clustered, so fairly easy to deal with
- Database failure: Could also be replicted and clustered. 
- Lock Server failure: The one true single point of failure, where if it fails, the queues keep filling up with jobs, and the system would run out of memory and crash. The solution is also (thankfully) simple. It can be implemented on top of a [Raft cluster](https://en.wikipedia.org/wiki/Raft_(algorithm)), which should prevent the lock server from being our single point of failure. Unfortunately, that would come at a cost, latency. The raft leader now has to check that enough servers have the latest data in the logs, before confirming that a job can take place, and would thus slow down the system considerably.

**Note**: 
- The reason I decided to defer the implementation of all of the above solutions is because they don't reallu require much thought, just a matter of configuration and implementation. The solutions are fairly easy to implement, and the goal of this project is to focus more on the coding and design aspect of things.
- Long timeouts between containers are only due to docker-compose's depends-on command only waiting for the container to start, not to be ready to accept connections. In a real production system, this would probably be less of an issue.


## **Further Optimizations**
- ### **WebSocket Optimizations**
- The Problem: We start a websocket connection for each client. Each connection has read and write buffers that are 4kb each due to the net/http package, and the [gorilla package](#https://github.com/gorilla/websocket) also has its own additional buffers when upgrading to websockets. Furthermore, each goroutine’s stack is 4kb, but can vary according to the os. So starting a goroutine for each client would cost about 20kb, not to mention any of the internal structs and data I associate with each client goroutine in my app. This would never scale well, especially since connections are idle 99% of the time, and are just sitting on buffers that can’t be garbage collected, and so the memory usage increases beyond control. RAM becomes a bottleneck.

- The Solution: An optimization to be made on the websocket’s side of things is to use a lower level websockets api such as https://github.com/gobwas/ws in addition to dealing with the kernel from the application level using a system call like [epoll](https://man7.org/linux/man-pages/man7/epoll.7.html). This would mean that we would have a thread responsible for the async epoll call, and it only spawns new goroutines when the kernel signals that any of the file descriptors it was listening on does indeed have IO activity. So in total, we now only have as many goroutines as we do active connections, rather than a goroutine for each connection. (Will probably implement later)

- My Implemented Solution: A slight optimization I decided to implement is create one goroutine per connection, which is for the reader, and to only create one for the writer whenever there is a need to send data. This means that I have to keep a central map mapping each client’s Id with a pointer to its websocket connection struct. I have to use locks, but since this map would be very read heavy, and with very little writes (only in case of adding a connection or removing one) a read/write lock would be suitable. I also need to make sure that I close all connections after a specific idle timeout, to make sure the map doesn’t increase in size too much. 

- ### **Load Balancing Optimizations**

- The Problem: It is quite tricky to load balancer with websockets since they are stateful, not just a quick request response situation like HTTP requests are. The load balancer (if layer 7) would have to maintain 2 connections from both the client to the load balancer, and from the load balancer to the server. So if I do manage to build an extremely optimized and efficient server that can handle 100k or even a million connections, the load balancer would become a bottleneck since it has to maintain twice that number.
- The Solution:  There is an ingenious idea and it goes like this. We don’t even use a load balancer! [This article explains it beautifully](https://dzone.com/articles/load-balancing-of-websocket-connections). TLDR; Don't use a load balancer, build your own. Clients first ask it for an Ip Address to connect to. It then contacts all the websocket servers, and asks which of them would be able to handle an incoming connection. It then sends the reply back to the client. GENIUS!

- ### **General Optimizations**

- If a result for a job is partially present in cache (url requested with depth of 3, but only depth of 2 is present), don't discard the cache and start over, rather build upon the existing results. Easy to implement, probably will implement it soon.
- Decrease all field names' length in cache, urlToCrawl -> url. Also super easy to implement.
- Might as well decrease all field names' size, to decrease the message size over the wire. 
- Use normal HTTP requests between system components, and use [Protobuf](https://developers.google.com/protocol-buffers) to compress the data and decrease the "internal" network load.
- I currently just end all websocket connections after a timeout, which isn't optimal since a specific job may indeed take a considerable amount of time. A better solution would be to implement a ping/pong message system between the clients and the server, to prevent the connection from going idle. I should then close the connection after the client himself goes idle, or a job actually takes too much time, and closing the connection with the client would signal that his job is/would be cancelled.


## **Faults**
The system is not perfect, and listed below are many faults which I should definitely solve in the near future.
- Problem: A user can send a job that takes quite a bit of time, multiple times in succession. This causes the masters to become stuck while working on them. Thus, the Lock Server decides that since the masters are late, their jobs should be reassigned. It then reassignes the job to another master, even though there is absolutely no reason to. This would cripple the whole system
- Solution: 1- Prevent users from sending multiple requests at a time, only one per client at a time. DDos is still an issue though. 2- Allow the Lock Server access to the cache as well. Now, if a master is late, before reassigning the late job, it firsts checks if the job results are cached, and doesn't reassign them if the are present, since it understands they are already done. 3- Allow a channel of communication between Lock Server and Master where a Lock Server can inform a stuck master to stop processing a job if the results are already present in cache.
- Problem: Lock Server reassigns jobs after a specific amount of time, not through heartbeats between it and the masters. 
- Solution: Communicate via heartbeats with Master, and decide if Master is stuck and is actually not making any progress, before taking the decision to re-assign the pending jobs.
- Problem: Lock Server is a huge bottleneck, since all jobs have to pass by it before they can get processed. In my system, it doesn't make a difference since each job takes a minimum of atleast 5 seconds, but in a different system, it will definetly be a bottleneck.
- Solution: Rather than rely on the database for all queries, keep an in-memory cache of sorts, and respond to the master using this cache. Start a thread periodically every (variable) amount of seconds that pushes all the changes to the database, but the most important thing is to not rely on database queries for every single decision.
- Problem: A client establishes a websocket connection with websocket server S1. The server pushes the job into the Assigned Jobs Queue. By the time the message processing is done and the message is pushed to the Done Jobs Queue, S1 had died. The client may have or may have not re-established a connection with a different server. The problem lies in the fact that before a websocket server pulls a job from the Done Jobs Queue, it has to check if the client has an active websocket connection with it. If it doesn't, the message is pushed back into the queue so another websocket server picks it up and sends it the appropriate client. The issue is now apparent. If a client's connection has been terminated, his job would stay alive forever in the Done Jobs Queue, and these messages would build up and consume a considerable amount of memory.
- Solution: Each message should have a TTL in the message queue, and if it has passed this TTL, it is then discarded. In addition, websocket servers should be able to communicate with each other, and if a server gets a job for a client that isn't currently connected to him, it can then forward the job to the other servers to check to which server this client belongs to. There is also the option of every server storing his current clients in the cache cluster, where all servers can see each other's current active connections, so they can forward the job to the appropriate server. If no such client is found, the message is immediately discarded.

**Note**: 
- You need to have Go installed
- RabbitMq behaves weirdly in docker compose, and not all messages do make it into the queues. Try it out without docker-compose if you do face those issues

## **How To Run**

- To start the whole system:
```
cd /Distributed_Web_Crawler
docker-compose up --build --force-recreate
```
- To stop the whole system:
```
cd /Distributed_Web_Crawler
docker-compose stop
docker-compose rm -f 
docker-compose down --rmi local
```
- To create the web crawler's network:
```
docker network create Distributed_Web_Crawler
```
- To build a worker's image, run the container, then remove it afterwards:
```
cd /Distributed_Web_Crawler
docker build -f Server/Cluster/Worker/Dockerfile -t worker .
docker run --net=host --name workerContainer worker
docker rm -f workerContainer
```
- To build a master's image, run the container, then remove it afterwards:
```
cd /Distributed_Web_Crawler
docker build -f Server/Cluster/Master/Dockerfile -t master .
docker run --net=host --name masterContainer master
docker rm -f masterContainer
```
- To build a lockServer's image, run the container, then remove it afterwards:
```
cd /Distributed_Web_Crawler
docker build -f Server/LockServer/Dockerfile -t lock_server .
docker run --net=host --name lockServerContainer lock_server
docker rm -f lockServerContainer
```
- To build a clientFacingServer's image, run the container, then remove it afterwards:
```
cd /Distributed_Web_Crawler
docker build -f ClientFacingServer/Dockerfile -t client_facing_server .
docker run --net=host --name clientFacingServerContainer client_facing_server
docker rm -f clientFacingServerContainer
```
- To run a master:
```
cd /Distributed_Web_Crawler/server/cluster/master
MY_PORT=7777 LOCK_SERVER_PORT=9999 MQ_PORT=5672 go run -race master.go
```
- To run a worker:
```
cd /Distributed_Web_Crawler/server/cluster/worker
MASTER_PORT=8888 go run -race worker.go
```
- To run a lock server:
```
cd /Distributed_Web_Crawler/server/lockserver
MY_PORT=9999 DB_PORT=5432 go run -race lockServer.go
```
- To run a client-facing websocket server:
```
cd /Distributed_Web_Crawler/ClientFacingServer
MY_PORT=5555 MQ_PORT=5672 go run -race server.go
```
- To run RabbitMq:
```
docker run  --name rabbitmq-container -p 5672:5672 -p 15672:15672  rabbitmq:3-management
```
- To run PostgreSql:
```
docker run --name postgres-container -p 5432:5432 -e POSTGRES_PASSWORD=password postgres
```
