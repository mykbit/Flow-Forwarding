# Flow Forwarding
[Features](#features) | [Topology](#topology) | [Building and Running](#building-and-running) | [Customization](#customization)

Flow Forwarding is an implementation of a reactive routing approach, serving as an overlay
to establish routing information through the broadcasts of requests for the locations of destinations absent
from current routing tables. This approach finds practical application in scenarios such as ad-hoc
networks deployed in regions where traditional infrastructure is unavailable, such as those affected by natural
disasters. 

The design of forwarding mechanisms aims to reduce the communication necessary to establish forwarding information while providing flexibility and fault-tolerance. This project is a proof-of-concept implementation of this approach.

## Features
Flow Forwarding features two main components: endpoints and routers. The following are the capabilities of each component:

### Endpoint
* Broadcast various requests.
* Receive requests from other endpoints.
* Send data packets to other endpoints.
* Process incoming data packets.
* Provide different responses depending on the type of received packet.

If you are interested in the types of requests and responses the endpoints can send and receive, please refer to `protocol.md`.

### Router
* Process incoming requests and data packets.
* Forward requests and data packets to other routers or endpoints.
* Keep record of the paths to each endpoint.
* Clearing the routing table of expired entries.

The entities are self-sufficient and can be deployed in any topology. The following section describes the default topology used in this project.

## Topology
The topology used for testing the implementation includes 4 endpoints and 5 routers. The following image shows the described topology:

<p align="center">
  <img src="https://github.com/mykbit/Flow-Forwarding-Go/assets/96201443/b0a93fc4-c440-43ef-9334-82b7e96cb812" alt="Testing Topology">
</p>

However, the implementation is not limited to this topology. It can be as simple as a single endpoint residing in its own network which no router is connected to, or as complex as a mesh of routers, endpoints and networks in between them. The implementation is flexible enough to handle any topology without any changes to the code. 

## Building and Running
Requirements:
 - Docker
 - Go 1.21.4 or higher

1. Clone the repository: `git clone https://github.com/mykbit/Flow-Forwarding-Go.git`
2. Navigate to the project directory: `cd path/to/Flow-Forwarding-Go`
3. Build the project: `docker compose up --build`

The implementation delivers output in 2 forms: a stream of messages in the terminal and `.pcap` files, which can be found in the `/<entity_name>` directory of each container. Although the user is limited in terms of interaction with the program, it is possible to customize the information you wish to transfer from one endpoint to another. The following section describes what parameters can be changed to modify the behavior of the mechanism.

## Customization
The solution is highly flexible and can handle various changes to the topology and the behavior of the entities. The following parameters can be changed to modify the behavior of the entities:
- The number of endpoints and routers by modifying the `docker-compose.yml` file.
- The number of networks and their addresses by modifying the `docker-compose.yml` file.
- The networks to which the routers and endpoints are connected by modifying the `docker-compose.yml` file, which allows to build any topology.
- The frames and audio files sent by the endpoints by inserting the directory of format `/<any_name>/frames` and `/<any_name>/audio` into the endpoint folder and modifying the `.env` file, where you need to specify the path to your data by changing the `DATA_PATH` variable.
- The DNS addresses of the endpoints by modifying the `.env` file, where you need to change the `SOURCE_ID` and `DEST_ID` variables with respect to your topology. That also means that you can change the destination address of the data packets by changing the `DEST_ID` variable.
