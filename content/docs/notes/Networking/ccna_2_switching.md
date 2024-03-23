---
title: "CCNA 2: Ethernet Switching"
---

# CCNA 2: Ethernet Switching

*This section of the CCNA also covers switch configuration basics, but this is omitted in my notes.*

## Why switches?

Before switches, physical buses were the most common way to connect multiple devices onto a LAN. However, buses are L1 devices (they just combine the electrical signals), and frames sent from different devices could collide. Devices within the same bus/collision domain also shared the same bandwidth, and broadcasts from any one device went to *all devices on the LAN*.

When talking about switches, "interface" and "port" are often interchangeable.

**Switches** are much smarter, and primarily:

- Provide a large number of interfaces, breaking up each device into its own collision domain - basically eliminating the collision problem entirely
- Each interface/switch port provides dedicated bandwidth
- Each interface/switch port can use full duplex logic if the individual connected device supports it, without slowing down the whole network

*Symmetric* switches provide the same bandwidth to all connections. *Asymmetric* switches have different bandwidths available at different ports.

## Switching logic

1. Learn MAC addresses: Any received frame has a source MAC address, so the switch learns what MAC addresses it has connections to
    - If the switch knows that specific MAC address: it **forwards** that frame to that specific address, while also **filters** it for others - only one recipient gets the frame
    - If the switch doesn't know that specific MAC address, it forwards the frame to everyone connected to the switch ("flooding")
1. Given a frame, decide based on the destination MAC address whether to forward or not (filter) the frame
1. Prevent routing loops with Spanning Tree Protocol (STP)

Switch memory for MAC addresses is about 5 minutes, which is the default in Cisco IOS.

### L2 and L3 switching

A layer 2 switch makes decisions solely based on MAC addresses, and its functionality is "invisible" or transparent to all protocols and user applications. A layer 3 switch can get more involved and also use IP address information for routing decisions, which can then reduce the need for dedicated routers.

In a home network, the wireless access point is also the router and also the switch.

### Collision and Broadcast domains

Collision domain is a set of LAN interfaces whose frames can collide with each other. Switches resolve this for the most part, but not for broadcast domains. When a frame arrives at the switch and must be sent to all devices, collisions are possible.

## Frame forwarding

- Store-and-forward switching: Receive the frame, keep it in memory and analyse it, check data integrity with cyclic redundancy check (CRC), and only forward if the CRC passes.
- Cut-through-switching: Switch only reads/buffers the MAC address, and starts immediately sending out the entire frame to the destination while. No error checking.
- Fragment-free mode: Wait for collision window (64-bytes) before forwarding, ensuring that no fragmentation has occurred. Better error checking than cut-through, with practically no increase in latency.

While processing incoming data, switches store the frames briefly. This is done either with *port-based memory*, where frames are stored in queues by port, or via *shared memory* that is shared between all ports of the switch.
