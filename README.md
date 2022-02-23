# **Distributed Web Crawler**
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
- [**How To Run**](#how-to-run)

## **System Components**

- ### **Load Balancer**
    The load balancer of choice is HaProxy. The following highlights its main functionalities:
    - Be able to establish and maintain websocket connections between the client and the websocket servers.
    - Handle up to 50,000 (variable) concurrent connections at a time.
    
    Why I chose HaProxy:
    - The main reason I chose HaProxy over Nginx is because HaProxy fits my needs as a load balancer perfectly. Nginx would be overkill, and HaProxy uses Round Robin, which in my case, makes sense, since I wan't the websocket connections to be balanced amongst all websocket servers. It also supports websockets out of the box, so it was a perfect fit.


- ### **WebSocket Server**
    The load balancer of choice is HaProxy. The following highlights its main functionalities:
    - Be able to establish and maintain websocket connections between the client and the websocket servers.
    - Handle up to 50,000 (variable) concurrent connections at a time.
    - The main reason I chose HaProxy over Nginx is because HaProxy fits my needs as a load balancer perfectly. Nginx would be overkill, and HaProxy uses Round Robin, which in my case, makes sense, since I wan't the websocket connections to be balanced amongst all websocket servers. It also supports websockets out of the box, so it was a perfect fit.


- ### **Cache**
    The load balancer of choice is HaProxy. The following highlights its main functionalities:
    - Be able to establish and maintain websocket connections between the client and the websocket servers.
    - Handle up to 50,000 (variable) concurrent connections at a time.
    - The main reason I chose HaProxy over Nginx is because HaProxy fits my needs as a load balancer perfectly. Nginx would be overkill, and HaProxy uses Round Robin, which in my case, makes sense, since I wan't the websocket connections to be balanced amongst all websocket servers. It also supports websockets out of the box, so it was a perfect fit.


- ### **Message Queue**
    The load balancer of choice is HaProxy. The following highlights its main functionalities:
    - Be able to establish and maintain websocket connections between the client and the websocket servers.
    - Handle up to 50,000 (variable) concurrent connections at a time.
    - The main reason I chose HaProxy over Nginx is because HaProxy fits my needs as a load balancer perfectly. Nginx would be overkill, and HaProxy uses Round Robin, which in my case, makes sense, since I wan't the websocket connections to be balanced amongst all websocket servers. It also supports websockets out of the box, so it was a perfect fit.


- ### **Master**
    The load balancer of choice is HaProxy. The following highlights its main functionalities:
    - Be able to establish and maintain websocket connections between the client and the websocket servers.
    - Handle up to 50,000 (variable) concurrent connections at a time.
    - The main reason I chose HaProxy over Nginx is because HaProxy fits my needs as a load balancer perfectly. Nginx would be overkill, and HaProxy uses Round Robin, which in my case, makes sense, since I wan't the websocket connections to be balanced amongst all websocket servers. It also supports websockets out of the box, so it was a perfect fit.


- ### **Worker**
    The load balancer of choice is HaProxy. The following highlights its main functionalities:
    - Be able to establish and maintain websocket connections between the client and the websocket servers.
    - Handle up to 50,000 (variable) concurrent connections at a time.
    - The main reason I chose HaProxy over Nginx is because HaProxy fits my needs as a load balancer perfectly. Nginx would be overkill, and HaProxy uses Round Robin, which in my case, makes sense, since I wan't the websocket connections to be balanced amongst all websocket servers. It also supports websockets out of the box, so it was a perfect fit.


- ### **Lock Server**
    The load balancer of choice is HaProxy. The following highlights its main functionalities:
    - Be able to establish and maintain websocket connections between the client and the websocket servers.
    - Handle up to 50,000 (variable) concurrent connections at a time.
    - The main reason I chose HaProxy over Nginx is because HaProxy fits my needs as a load balancer perfectly. Nginx would be overkill, and HaProxy uses Round Robin, which in my case, makes sense, since I wan't the websocket connections to be balanced amongst all websocket servers. It also supports websockets out of the box, so it was a perfect fit.


- ### **Database**
    The load balancer of choice is HaProxy. The following highlights its main functionalities:
    - Be able to establish and maintain websocket connections between the client and the websocket servers.
    - Handle up to 50,000 (variable) concurrent connections at a time.
    - The main reason I chose HaProxy over Nginx is because HaProxy fits my needs as a load balancer perfectly. Nginx would be overkill, and HaProxy uses Round Robin, which in my case, makes sense, since I wan't the websocket connections to be balanced amongst all websocket servers. It also supports websockets out of the box, so it was a perfect fit.


## **Availability And Reliability**
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
the connections, they use [exponential backoff](https://en.wikipedia.org/wiki/Exponential_backoff), so as not to cause a [thundering herd problem](https://en.wikipedia.org/wiki/Thundering_herd_problem).
- Message Queue failure: RabbitMq could be replicated and clustered, so fairly easy to deal with
- Database failure: Could also be replicted and clustered. 
- Lock Server failure: The one true single point of failure, where if it fails, the queues keep filling up with jobs, and the system would run out of memory and crash. The solution is also (thankfully) simple, it can be implemented on top of a [Raft cluster](https://en.wikipedia.org/wiki/Raft_(algorithm)), which should prevent the lock server from being our single point of failure. Unfortunately, that would come at a cost, latency. The raft leader now has to check that enough servers have the latest data in the logs, before confirming that a job can take place, and would thus slow down the system considerably.

**Note**: The reason I decided to defer the implementation of all of the above solutions is because well, they are aren't really implementations. The solutions are fairly easy to implement, and the goal of this project is to focus more on the coding and design part of things.

## **Further Optimizations**


## **How To Run**
