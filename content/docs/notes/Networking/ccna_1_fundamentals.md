---
title: "CCNA Part 1: Networking Fundamentals"
---

# CCNA Part 1: Networking Fundamentals

## Overview of networking model

### `TCP/IP` Networking model

Layer | Name | Example technology | What's being transmitted| Function | OSI-model equivalent layer
-- | -- | -- | -- | -- | --
1 | (Link) Physical | RJ-45 cable | Electrical signals | Transmit information (=bits) over some physical medium from one device to another | Same
2 | (Link) Data  | Ethernet protocol, Point-to-Point Protocol (PPP) | "Frames" / L2 PDU | Encapsulation and addressing (MAC addresses) | Same
3 | Network | IP Protocol | "Packets" / L3 PDU | Addressing (IP addresses) and routing | Same
4 | Transport | TCP, UDP, QUIC | "Segments" / L4 PDU | Error recovery ? | Same
5-7 | Application | HTTP, SMTP, FTP, SSH | Application dependent, e.g. HTTP request | | Corresponds to Session/Presentation/Application layers 5/6/7 of OSI

When data is transmitted, each layer *encapsulates* its own data before passing on the message to the level below:

1. Application data is encapsulated first (e.g. HTTP headers and status)
2. Data provided from L5 is encapsulated in e.g. an L4 TCP segment
3. Data provided from L4 is encapsulated in an L3 IP packet
4. Data provided from L3 is encapsulated in an L2 ethernet frame
5. Physical transmission of the bits of the ethernet frames occurs to move data

Receiving data follows the above steps in the reverse order, where each layer *de-encapsulates* data before passing it to a higher layer.

### Layers, protocols and devices

Layer | Protocol | Common devices in this layer
-- | -- | --
5-7 | HTTP, SSH, SMTP | Hosts/servers, firewalls
4 | TCP, UDP, QUIC | Hosts/servers, firewalls
3 | IP | Routers
2 | Ethernet, HDLC | LAN switches, wireless access points, modems
1 | RJ-45 | Cables, LAN hubs, LAN repeaters

## Local Area Networks (LANs)

A typical home or small office LAN usually consists of a switch, router, a wireless access point, and a modem. In many cases the first three at least are the same device (switch/router/AP), and just referred to as a "router". The same device may also be a modem.

More complex "enterprise" LANs differ from home/small office LANs in the number of devices and level of specialisation of most of the networking equipment - in a large LAN it is less likely for example to combine a switch and a router in one device, as many switches may need to be chained together to cover the larger number of client devices.

{{<mermaid>}}
flowchart BT

PC1
PC2
PC3
PC4
Phone1
Phone2

subgraph stuff["The stuff that a home 'router' does in one device at a smaller scale"]
    sw1["Switch #1"]
    sw2["Switch #2"]
    ap["Wireless access point\nWIFI/WLAN"]
    swd["Distribution switch"]
    rt["Router"]
end

Phone1 --> ap --> swd --> rt --> Internet
Phone2 --> ap
PC1 --> sw1 --> swd
PC2 --> sw1
PC3 --> sw2 --> swd
PC4 --> sw2

{{</mermaid>}}

### Ethernet physical layer standards

Speed | Name | IEEE Standard (informal) | IEEE Standard (formal) | Cable type
-- | -- | -- | -- | --
10 Mbps | Ethernet | 10BASE-T | 802.3 | Copper, UTP (Unshielded Twisted Pair)
100 Mbps | Fast Ethernet | 100BASE-T | 802.3u | Copper UTP
1000 Mbps | Gigabit Ethernet | 100BASE-LX | 802.3z | Fiber
1000 Mbps | Gigabit Ethernet | 100BASE-T | 802.3ab | Copper UTP
10 Gbps | 10 Gig Ethernet | 10GBASE-T  | 802.3an | Copper UTP

### Ethernet cable stuff

- Up to 802.3u, RJ-45s have two pairs of wires. The more capable/modern standards have four pairs to improve bandwidth for a total of 8 wires/pins
- RJ-45 cables can be "straight-through" or "crossover" cables, depending on whether the pins on both ends map directly, or have been crossed, respectively
  - One set of pins transmits in one direction; the other set transmits in a different direction
  - PC NICs, Routers and Wireless APs transmit on pins 1,2
  - Hubs, Switches transmit on pins 3,6
  - As a result, you could use a straight-through cable between a PC and a Switch; but to connect two switches, you'd need a crossover cable
  - In reality, today 99% of devices don't care about this anymore, and support "auto MDI-X"/"auto crossover", making this a non-issue

### Ethernet as a data-link protocol

- Devices that "have Ethernet" have Network Interfaces, usually Network Interface Cards (NICs) that handle the connection
- Ethernet frame has a *header* and a *trailer* (data to be transmitted sits in between them)
- Contents of the header include:
  - Preamble for synchronisation
  - Start Frame Delimiter (SFD)
  - Destination MAC address
  - Source MAC address
  - Type of protocol inside the frame, usually IPv4 or IPv6
  - Data (with padding if needed): 46-1500 bytes
- Contents of the *trailer* include just *FCS* (Frame Check Sequence) which is used to confirm whether the data was transmitted correctly based on a hash/checksum-type operation. Malformed frames are discarded; Ethernet does not provide error recovery but expects higher layers to do that (e.g. TCP).
- "Full duplex": Devices can send and receive at the same time
- "Half duplex": Devices can only either send or receive at any one time. Collisions are possible, and must be handled by CSMA/CD (Carrier-Sense Multiple Access with Collision Detection)

### MAC (Media Access Control) addresses

- AKA: LAN address, Ethernet address, hardware address, burned-in address, physical address, universal address
- 48-bit long binary number, unique in the universe for each device, where "device" here means a NIC.
  - This is an administrative thing; manufacturers get their own unique 3-byte code called the Organizationally Unique Identifier (OUI), and each MAC address "burned" on the devices a company produces must start with those 3 bytes.
- *Unicast* address means an address of an individual device
- *Multicast* address means an address listened on by multiple devices
- *Broadcast* address is the address listened on by every device (`FFFF.FFFF.FFFF.FFFF`) on the LAN

### Wide Area Networks (WANs)

You own LANs, but often have to lease WANs to connect your local network to the broader (inter)net. Telecom companies make this happen with various technologies, including HDLC (High-level Data Link Control), PPP (Point-to-Point Protocol) and Ethernet (Ethernet emulation, or Internet over Multi-Protocol Label Switching - EoMPLS).

#### Leased-line WAN

- AKA: Leased circuit, circuit, serial link/line, point-to-point link/line, T1, WAN link, link, private line
- Direct connection from one LAN to another
- Physical details are unknown to the end client and up to the telco, but usually ~all buildings have connections wired up and if/when tenants purchase connections from the telco, these connections are activated
- Various L2 protocols can be used as mentioned above, which may then require different/specialized equipment on both ends of the leased line
- Regardless of protocol, a separate set of headers and trailers is used for the leased-line connection over WAN
  - Ethernet frame used in the local LAN is discarded, and the content is repackaged into a new Ethernet frame for transmission over the line

{{<mermaid>}}
flowchart LR
subgraph LAN1
PC1
R1
end

subgraph LAN2
PC2
R2
end

PC1 <--> R1 <-->WAN["WAN Connection"]<--> R2 <--> PC2

{{</mermaid>}}

#### Other alternatives besides leased-lines

This is what consumers mostly use, though many small(er) businesses also use these for their internet access.

- Fiber
- 4G/5G routers to access the internet via mobile data networks
- Cable - CATV cable
- DSL - telephone line

#### Internet as a large WAN

- End users connect their LANs via their ISP (Internet Service Provider) to the ISP's WAN
- ISPs are then connected with each other, in a 'broader WAN, allowing routing across the entire internet

## IPv4 Addressing & Routing

- Routing: Hosts/routers forwarding IP packets (L3PDUs), while relying on underlying network to forward the bits
- IP addressing: (Grouped) Addresses used to identify destination and source
- IP routing protocol: A way for routers to dynamically 'learn' about what IP address groups should be routed where
- Other utilities: DNS, ARP (Address Resolution Protocol)

### Basic routing logic flow

1. If the target is in the same LAN as me (I can reach them directly), send the packet to them
2. Otherwise, if I'm a host, send the packet to my `default gateway` - the router will figure out what to do with it
3. If I'm a router, send the packet to the right place as its next *hop* based on my *IP routing table* (assuming FCS produced no errors; if we had an error with the Ethernet frame, discard the frame.)

Depending on the connectivity between routers, packets will be de-encapsulated and encapsulated multiple times over HDLC, Ethernet, or other protocols.

### IP Addressing and routing

IPv4 addresses are usually written out in "dotted decimal notation" or DDN format, like so: `_._._._`, e.g. `127.0.0.1`, where each section is one octet (an 8-bit binary number). IP addresses are 'naturally' grouped since each section takes the range `0-255`, so e.g.:

- `192.168.0.1`
- `192.168.0.100`
- `192.168.0.125`

Can all be "grouped" under `192.168.0.0` - a set of consecutive addresses, or an *IP network*. IP addresses in the same group must not be separated from each other by a router. An address ending in `0` is considered the identifier of the network, and the address ending in `255` is a special address that broadcasts to all network participants. Hence the usable values for hosts are `1-254`, inclusive, for each octet.

### IP Network classes

The classes are listed below by the value of the first octet. The fraction refers to the proportion of IPv4 addresses belonging to that class, out of all available addresses.

Range | Class | Size | Usage
-- | -- | -- | --
`1-126` | Class A | 1/2 | Unicast
`127` | Reserved | (part of class A's 1/2) | `localhost`/loopback usage
`128-191` | Class B | 1/4 | Unicast
`192-223` | Class C | 1/8 | Unicast
`224-239` | Class D | 1/16 | Multicast
`240-255` | Class E | 1/16 | Reserved

### IP Subnetting

Subnetting divides an IP network into smaller groups, so that less IPs may go unused/wasted. Not much magic or detail here yet.

### Routing protocols

Hosts rely on routers to know where to send packets, but network structures can change all the time. Routing protocols are what routers use to communicate with each other and figure out which router(s) can handle which network groupings. Routing protocols:

- Dynamically update and 'learn' each subnet in the network
- Try to optimise and provide the 'best' route for a given packet
- Deprecate invalid or no longer existing routes
- Prevent routing loops

The basic process of a routing protocol is:

1. Each router adds a route to its own routing table about subnets directly connected to that router.
2. Each router sends all neighbouring routers the information in its routing table (this is called a *routing update*), including the routes from step #1 as well as any routers learned from other routers.

### DNS

- UDP over port 53, ask for the IP of a server based on hostname

### ARP - Address Resolution Protocol

- Method for hosts and routers to learn the MAC address corresponding to a server's current IP address
- Send out a request, "if this is your IP, please reply with your MAC", store the result in cache
- Sent over multicast/broadcast address to everyone on the network
- Try `arp -a` to check the current contents of the ARP cache

### ICMP Echo - `ping`

- Ping (Packe Internet Groper) uses Internet Control Message Protocol (ICMP) called *ICMP echo request* to a particular IP, to test basic connectivity of the IP network
