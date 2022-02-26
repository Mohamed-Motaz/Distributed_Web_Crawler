# **Distributed Web Crawler**
![systemArchi drawio](https://user-images.githubusercontent.com/53558209/155814673-c201500d-7f48-4223-82a3-7bf9b7633190.png)
The main objective of the Distributed Web Crawler is to serve as a template for a system than can parallelize tasks and distribute the load, be it cpu and processing load, or even network load, as is the case of this Web Crawler. The system is [**distributed**](#system-components) on multiple machines, [**highly available**](#high-availability-and-reliability), [**reliable**](#high-availability-and-reliability) and [**fault tolerant**](#fault-tolerance). 

## **Table of Contents**
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
- [**Faults (Yup, and many of them)**](#faults)
- [**How To Run**](#how-to-run)

## **System Components**

- ### **Load Balancer**
    The load balancer of choice is HaProxy. The following highlights its main functionalities:
    - Be able to establish and maintain websocket connections between the client and the websocket servers.
    - Handle up to 50,000 (variable) concurrent connections at a time.
    
    Why I chose HaProxy:
    - The main reason I chose HaProxy over Nginx is because HaProxy fits my needs as a load balancer perfectly. Nginx would be overkill, and HaProxy uses Round Robin, which in my case, makes sense, since I wan't the websocket connections to be balanced amongst all websocket servers. It also supports websockets out of the box, so it was a perfect fit.


- ### **WebSocket Server**
    The client facing servers use websockets to communicate with their clients. The following highlights its main functionalities:
    - Responsible for establishing and mainting active websockets with the clients.
    - Responsible for cleaning up and closing all connections that have been idle for over a (variable) amount of time.
    - Publish jobs to the Assigned Jobs Queue if no results could be found in the cache.
    - Consume done jobs from the Done Jobs Queue and send them over to the clients.
    
    Why I chose websockets:
    - The main reason I chose websockets is well, because they are trendy! Obviously not just that, but I had 2 other choices, polling every 5 seconds or so with normal HTTP request-response, or use [Server-Sent Events](https://en.wikipedia.org/wiki/Server-sent_events#:~:text=Server%2DSent%20Events%20(SSE),client%20connection%20has%20been%20established.). Both would have been fine, but that is only because in my system, client requests usually take a few seconds up to a few minutes to complete, so polling wouldn't cause much overhead. I decided to stick with websockets though since I wanted to go with a more general solution that (in case requests actually only do take a few hundered milliseconds to be processed), and avoid having to constantly poll the server, which in some cases would actually cause more overhead than just maintaining one TCP connection over the client's lifecycle.


- ### **Cache**
    The cache of choice is Redis. The following highlights its main functionalities:
    - Serve as a key value store, where each key has a set [TTL](https://en.wikipedia.org/wiki/Time_to_live)
    - Each key is a url, and its value contains the depth crawled, and the crawled websites 2D list.
    - Eg. If a client asks for "google.com", with a depth of 2, then the cache must contain "google.com" with atleast a depth of 2, so the client can be served immediately without any additional delay.

    Why I chose Redis:
    - The main reason I chose Redis is because it supports clustering and replication, (I can implement it in the future), and it seems like a fairly popular choice, so why not?


- ### **Message Queue**
    The message queue of choice is RabbitMq. The following highlights its main functionalities:
    - Durable, so in case of a crash, "all" jobs in the queue can be restored from disk
    - Support multiple producers and consumers
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
    - Keep track of the progress at all times, and notify the Lock Server when they are done
    - Push done jobs to the Done Jobs Queue so that the Websocket servers can consume them and semd the results to the waiting clients.


- ### **Worker**
    Workers are the powerhouses of the system. They are completely stateless, and only know about their master's address. The following highlights their main functionalities:
    - Communicate with the masters, ask for jobs, and respond with the results.
    

- ### **Lock Server**
    The Lock Server is the server tasked with persisting all system jobs in the database, so that in case of failure, all jobs can still be recovered. To avoid being a single point of failure, which it is, it should be implemented on top of a [Raft cluster](https://en.wikipedia.org/wiki/Raft_(algorithm)). The following highlights its main functionalities:
    - Make quick decisions on whether a master should start a job, or if there is any higher priority job to be given.
    - Keep track of all current jobs, and re-assign any jobs that are delayed beyond a (variable) time.
    - Persist all jobs information to the database.
    - Be extremely performant, since every single jobs needs to pass by the Lock Server before it can be processed.


- ### **Database**
    The database of choice is PostgreSQL. The following highlights its main functionalities:
    - Persist data in case masters die, thats it. I bet you didn't expect much to be honest.
    
    Why I chose PostgreSQL:
    - The main reason I chose PostgreSQL is because it supports clustering (not natively) and replication, (I can implement it in the future). But I mean, everything supports clustering and replication nowadays, so I really just wanted to try it out.

## **#High Availability And Reliability**
- To the client, the system has really high availability. The only case where the system would be down is if all the load balancers, websocket servers, or message queues are down. Since all of this are/can be implemented in clusters, the system is indeed highly available.
- The system is highly reliable and consistent, since it doesn't rely on some kind of consensus among all of its masters, or workers, or any of the componenents. Each component is completely stateless, and the end result is a reliable system that would deliver the same result every time a user adds a job.

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
- Lock Server failure: The one true single point of failure, where if it fails, the queues keep filling up with jobs, and the system would run out of memory and crash. The solution is also (thankfully) simple, it can be implemented on top of a [Raft cluster](https://en.wikipedia.org/wiki/Raft_(algorithm)), which should prevent the lock server from being our single point of failure. Unfortunately, that would come at a cost, latency. The raft leader now has to check that enough servers have the latest data in the logs, before confirming that a job can take place, and would thus slow down the system considerably.

**Note**: The reason I decided to defer the implementation of all of the above solutions is because well, they are aren't really implementations. The solutions are fairly easy to implement, and the goal of this project is to focus more on the coding and design part of things.

## **Further Optimizations**

## **Faults**
The system is not perfect, and listed below are many faults which I should solve in the near future, if I'm not too lazy that is.
- Problem: A user can send a job that takes quite a bit of time, multiple times in succession. This causes the masters to become stuck while working on them. Thus, the Lock Server decides that since the masters are late, their jobs should be reassigned. It then reassignes the job to a master, even though there is absolutely no reason to. This would cripple the whole system
- Solution: 1- Prevent users from sending multiple requests at a time, only one per client at a time. DDos is still an issue though. 2- Allow the Lock Server access the cache as well. Now, if a master is late, before reassigning the late job, it firsts checks if the job results are in cached, and doesn't reassign them, since it understands they are already done. 3- Allow a channel of communication between Lock Server and Master where a Lock Server can inform a stuck master to stop processing a job if the results are already present and in cache.
- Problem: Lock Server reassigns jobs after a specific amount of time, not heartbeats between it and the masters. 
- Solution: Communicate via heartbeats with Master, and decide if Master is stuck and is actually not making any progress, before taking the decision to re-assign the pending jobs.
- Problem: Lock Server is a huge bottleneck, since all jobs have to pass by it before they can get processed. In my system, it doesn't make a difference since each job takes a minimum of atleast 5 seconds, but in a different system, it will definetly be a bottleneck.
- Solution: Rather than rely on the database for all queries, keep an in-memory cache of sorts, and respond to the master using this cache. Start a thread periodically every (variable) amount of seconds that pushes all the changes to the database, but the most important thing is to not rely on database queries for every single decision.

## **How To Run**
