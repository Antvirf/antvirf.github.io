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
