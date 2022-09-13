# What do Layerfiles do?

Layerfiles are a format for quickly building, starting, and stopping VMs.

They're particularly useful for local developer environments and running acceptance tests.

When you build a Layerfile, it automatically takes snapshots along the way and re-uses those the next time you need to build it.

Layerfiles are based on Dockerfiles, so they should be immediately familiar to anyone that's used Docker.

# Commands

## ./lf build (directory containing Layerfile ...)

Builds one or more Layerfiles into VMs.

## ./lf vm list

Lists VMs on your computer

## ./lf image list

Lists VM images on your computer

## ./lf start [vm]

Starts a VM image to create a new VM

## ./lf stop [vm]

Pauses and stops running VM

## ./lf deploy [image]

Deploys a VM to the cloud
