+++ 
date = 2026-01-18
title = "On-demand OpenVPN setup on Linode"
author = "Antti Viitala"
tags = [
    "terraform",
    "linode",
    "vpn"
]
+++

A VPN can be useful occasionally, but as I'm too cheap to continuously pay for a service I rarely use, I wanted to set up something that could be spun up and down quickly with a few commands when needed. Having used the OpenVPN offering from Linode marketplace before, this setup is based on that:

- Terraform to create and configure an OpenVPN server in a particular Linode region, with a public IP
- Scripts (executed also via Terraform) to configure and extract `.ovpn` configs
- `openvpn3` installed locally and used as a client

The terraform setup looks like this:

```
.
├── modules
│   └── vpn
│       ├── main.tf
│       ├── outputs.tf
│       ├── provider.tf
│       └── variables.tf
└── vpns.tf
```

The the `vpns.tf` file is straightforward:

```hcl
# Linode provider setup
terraform {
  required_providers {
    linode = {
      source  = "linode/linode"
      version = "2.33.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.6.3"
    }
  }
}

provider "linode" {
  token = YOUR_LINDE_API_TOKEN
}


# VPN Config
# You can access the UI via $IP:943
# or directly connect with the created profile files that appear in the current dir, with:
# start : openvpn3 session-start --config ./$REGION-g6-nanode-1-client.ovpn
# stop  : openvpn3 session-manage --disconnect --config ./$REGION-g6-nanode-1-client.ovpn

locals {
  vpns = {
    # Uncomment a VPN server you wish to create. Key = region, value = VM SKU
    # "ap-southeast" = "g6-nanode-1" # Sydney
    # "ap-south" = "g6-nanode-1"     # Singapore
    # "ap-northeast" = "g6-nanode-1" # Tokyo
    # "de-fra-2" = "g6-nanode-1"     # Frankfurt 2
    # "se-sto" = "g6-nanode-1"       # Stockholm
    # "nl-ams" = "g6-nanode-1"       # Amsterdam
    # "gb-lon" = "g6-nanode-1"       # London 2
  }
}

module "linode_openvpn" {
  for_each      = local.vpns
  source        = "./modules/vpn"
  region        = each.key
  instance_type = each.value
  allow_access_from = ["$MY_IP/32"]
}
```

The usage is simple:

1. Uncomment / add to the `locals.vpns` map a server based on the desired region, e.g. `ap-south`
2. `terraform apply` to create it, takes about 5 minutes.
3. Connect to it: `openvpn3 session-start --config ./ap-south-g6-nanode-1-client.ovpn`
4. Disconnect when done: `openvpn3 session-manage --disconnect --config ./ap-south-g6-nanode-1-client.ovpn`
5. Tear down the instance(s) with `terraform destroy`

That's it. A `g6-nanode-1` costs `0.0075` USD per hour of usage and comes with 1TB of network transfer at the time of writing - very cost effective!

## VPN Module contents and future improvements

The module implements a very basic security measure of whitelisting only your current IP for inbound access to the instance. This is not super convenient if you move around a lot, but works well in a home VPN setup where your own public IP is relatively static. A more "properly" locked down instance would open up specific OpenVPN ports to allow access from anywhere and use stronger credentials, but that's beyond the scope of this simple setup.

Contents of the module are included below:

```hcl
# providers.tf
terraform {
  required_providers {
    linode = {
      source  = "linode/linode"
      version = "2.33.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.6.3"
    }
  }
}

# variables.tf
variable "region" {
  type = string
}

variable "instance_type" {
  type = string
}

variable "ssh_pub_key" {
  type    = string
  default = "" # YOUR SSH PUBLIC KEY
}

variable "vpn_user_name" {
  type    = string
  default = "" # DESIRED USERNAME
}

variable "soa_email_address" {
  type    = string
  default = "" # YOUR EMAIL HERE
}

variable "allow_access_from" {
  type    = list(any)
  default = ["0.0.0.0/0"] # Allow from anywhere
}

# Random vars
resource "random_string" "root_pass" {
  length  = 48
  special = false
}


# main.tf
resource "linode_instance" "opvnpn_instance" {
  label           = "${var.region}-${var.instance_type}-openvpn"
  image           = "linode/ubuntu22.04"
  region          = var.region
  type            = var.instance_type
  root_pass       = random_string.root_pass.result
  authorized_keys = [chomp(var.ssh_pub_key)]

  stackscript_id = 401719 # OpenVPN One Click (the marketplace OpenVPN one)


  # Config data
  # Check available values with `curl -H "Authorization: Bearer $LIN_TOKEN" https://api.linode.com/v4/linode/stackscripts/401719 | jq`
  stackscript_data = {
    user_name         = var.vpn_user_name
    disable_root      = "No" # Whether to disable root access over SSH
    soa_email_address = var.soa_email_address

    # DNS settings if using custom DNS
    subdomain      = "" # used for subdomains, e.g. www
    domain         = "" # used for domain, e.g. example.com
    token_password = "" # used for DNS records management in Linode, if domain is configured
  }
}

resource "time_sleep" "wait_after_provisioning" {
  depends_on      = [linode_instance.opvnpn_instance]
  create_duration = "300s"
}

resource "null_resource" "fetch_openvpn_credential" {
  depends_on = [time_sleep.wait_after_provisioning]
  triggers = {
    always_run = timestamp()
  }

  provisioner "local-exec" {
    command = <<-EOT
        ssh -o StrictHostKeyChecking=no \
            -o ConnectTimeout=10 \
            -i ~/.ssh/id_rsa \
            root@${linode_instance.opvnpn_instance.ip_address} \
            "sudo cat /usr/local/openvpn_as/init.log 2>/dev/null | grep 'To login' 2>&1" | awk -F'"' '{print "<auth-user-pass>\n"$2"\n"$4"\n</auth-user-pass>"}' \
            > "${path.root}/${var.region}-${var.instance_type}-openvpn.secret"
        EOT
  }
}

data "local_file" "vpn_credential" {
  depends_on = [null_resource.fetch_openvpn_credential]
  filename   = "${path.root}/${var.region}-${var.instance_type}-openvpn.secret"
}


resource "null_resource" "create_client_profile" {
  depends_on = [null_resource.fetch_openvpn_credential]
  triggers = {
    always_run = timestamp()
  }
  provisioner "local-exec" {
    command = <<-EOT
        ssh -o StrictHostKeyChecking=no \
            -o ConnectTimeout=10 \
            -i ~/.ssh/id_rsa \
            root@${linode_instance.opvnpn_instance.ip_address} \
            "sacli --prefer-tls-crypt-v2 --user openvpn GetUserlogin" \
            > "${path.root}/${var.region}-${var.instance_type}-client.ovpn"
        EOT
  }
}

resource "null_resource" "combined_client_profile" {
  depends_on = [
    time_sleep.wait_after_provisioning,
    null_resource.create_client_profile,
    null_resource.fetch_openvpn_credential
  ]
  provisioner "local-exec" {
    command = <<-EOT
      cat "${path.root}/${var.region}-${var.instance_type}-openvpn.secret" >> "${path.root}/${var.region}-${var.instance_type}-client.ovpn"
      EOT
  }
}

data "local_file" "client_profile" {
  depends_on = [time_sleep.wait_after_provisioning]
  filename   = "${path.root}/${var.region}-${var.instance_type}-client.ovpn"
}

resource "linode_firewall" "vpn_firewall" {
  label = "${var.region}-${var.instance_type}-ovpn"

  inbound {
    label    = "allow-inbound-tcp"
    action   = "ACCEPT"
    protocol = "TCP"
    ports    = "1-1000"
    ipv4     = var.allow_access_from
  }

  inbound_policy  = "DROP"
  outbound_policy = "ACCEPT"
  linodes         = [linode_instance.opvnpn_instance.id]
}
```
