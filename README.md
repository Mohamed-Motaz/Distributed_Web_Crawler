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
