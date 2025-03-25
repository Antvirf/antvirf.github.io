---
title: "Even simpler network booting with Pixiecore"
date: "2025-03-25"
weight: 10
---

*See earlier article on [basic iPXE booting](https://aviitala.com/docs/homelab/networking/basic-ipxe-booting/) for more details.*

Turns out, there's an even easier way. [Pixiecore](https://github.com/danderson/netboot/tree/main/pixiecore) is a simple open source tool that makes netbooting trivial.

1. Make sure your router *isn't* serving any PXE boot or TFTP related options. It won't be by default, so if you have not configured any of this - *do not configure any of this*.
2. Install `pixiecore` with `go install go.universe.tf/netboot/cmd/pixiecore@latest`
3. Run `$GOPATH/pixiecore quick xyz --dhcp-no-bind`
4. Boot a machine on the same network, and it will now load to [Netboot.xyz](https://netboot.xyz/) for you to pick an OS. That's what the `xyz` argument in the command is for, but many other operating systems like CentOS and Debian are included out of the box.

No need to run a TFTP server, no need to configure your router, just run pixiecore. Beyond a static setup like showcased above, you can also run [pixiecore in API mode](https://github.com/danderson/netboot/tree/main/pixiecore#pixiecore-in-api-mode), where it will ask a REST API that you implement *how to boot a machine*, given its MAC address - so you can customise things even further. For inspiration on what you can do with Pixiecore, [check out this excellent article from Railway](https://blog.railway.com/p/data-center-build-part-two), and this [quick mockup](https://github.com/Antvirf/metal-control-plane-mockup) on how to do it yourself.

