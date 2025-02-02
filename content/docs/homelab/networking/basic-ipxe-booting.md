---
title: "Simple network (iPXE) booting with netboot.xyz"
date: "2025-02-01"
weight: 4
---

No matter what operating systems you install and how often or rarely you need to do it, setting up your local network for network booting is a big quality-of-life improvement - instead of needing to download ISOs and flashing USB sticks yourself, your machines can get the most up-to-date operating system images they need from the internet directly.

## What does 'network booting' / 'netbooting' actually mean and how does it work?

Instead of booting from a hard drive or a USB stick, the computer will start up, join a network, attempt to download specified booting media, and then boot directly from that media.

In order for this process to work, first the computer will need to have network access over a wired connection.

Second, on connecting to the network, the computer will send out a request asking for the location (the IP address) of another server in the network that will provide the boot media, as well as the specific file path to ask for.

Third, after receiving the target IP and path, your computer makes a request to the target server with [TFTP](https://en.wikipedia.org/wiki/Trivial_File_Transfer_Protocol), downloads the boot media, and starts the booting process as normal.

## Setting up network booting

1. Network router needs to know (a) the IP of the target server; (b) the file path of the boot image to provide.
2. The computer to be netbooted needs to have a wired network connection.
3. The target server needs to exist and be hosting a TFTP server.
4. The target server needs to have the requested boot image (e.g. a Linux ISO) available.

#1 is usually a simple configuration option that can be set in any router. #3 takes a bit of local setup and requires another server to be running on the local network. #4 is where netboot.xyz comes in.

## What does `netboot.xyz` do?

The process above still requires the target server to have a specific ISO of the operating system you may wish to install. Often, you may wish to try different versions or distributions, so this quickly becomes a large number of files to manage and maintain. 

As an alternative, pointing your boot process to `Netboot.xyz` first downloads a text-based menu for you to choose a desired operating system distribution and version (and also offers you various helper and recovery utilities for e.g. formatting disks), and downloads your chosen ISO from a centrally managed location.

This way, your local server only hosts a small file that points to `netboot.xyz`, and any OS image you may need is downloaded only when it is needed.

![img](https://netboot.xyz/assets/images/netboot.xyz-d976acd5e46c61339230d38e767fbdc2.gif)

*This is how the netboot menu looks. That's a lot of operating systems to host yourself.*

## Running a local TFTP server with a script pointing to Netboot

TFTP servers are extremely small and simple. The below commands provide a minimal example of installing `tftp` on a Debian-based system and setting it up with a file that will point the booting computer to netboot.xyz. Note that if you have [`dnsmasq` running, it can also acts as a TFTP server.](https://wiki.archlinux.org/title/Dnsmasq#TFTP_server)

```bash
# run all commands with `sudo` or as root
apt update
apt install tftpd-hpa
systemctl enable tftpd-hpa

cat <<EOF >>/srv/tftp/netboot
#!ipxe
echo Hello! This is an iPXE script that will point you to Netboot in 5 seconds, hold on...
sleep 5
chain --autofree http://boot.netboot.xyz
EOF
```

## Configuring your router

Assuming we ran the above TFTP server on `192.168.168.168`, we would configure the router's DHCP options on network booting to point to this address and to the `netboot` filepath (since we created a file called `netboot` in the TFTP server folder).

Configuring this varies by router, but generally these options can be found under DHCP-related settings, looking for terms like `PXE`, `iPXE`, `netboot` and `network booting`.

![](/content/unifi-dhcp-netboot.png)

## Testing it out

To network boot your computer, ensure that in BIOS the netboot or 'PXE boot' is the first option in your boot order settings.

## Going further

This setup is a good starting point, but in a corporate environment it is likely that:

- Machines cannot generally assume to have internet access, so the required operating system images need to be hosted and managed locally instead of using something like Netboot.xyz.
- Standardisation to one distribution and/or a set of configurations is desirable, so having a large variety of operating systems available for booting (versus a well-defined 'standard' setup) is not a priority.
- A custom iPXE menu or a set of scripts similar to Netboot.xyz can still be useful internally, to choose among e.g. OS versions, machine roles, offer utilities, or further configure installation specifics like hard drives, network interfaces or hostnames.
