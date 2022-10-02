![Layerfiles](https://raw.githubusercontent.com/webappio/assets/main/github-header.svg)

# Layerfiles are Dockerfiles that build VMs

They're particularly useful for local developer environments and running acceptance tests.

When you build a Layerfile, it automatically takes snapshots along the way and re-uses those the next time you need to build it.

Layerfiles are based on Dockerfiles, so they should be immediately familiar to anyone that's used Docker.

## Installation

Layerfiles are built as a single static binary, available here:

- [Linux x86-64](https://github.com/webappio/assets/raw/main/lf)
- Other distributions TBD

## Project status

Layerfiles is currently in alpha - it's possible to build VMs but not run them. There also aren't builds available for Mac or Windows.

If the project interests you and you'd like to contribute, we do have issues marked "good first issue" [here](https://github.com/webappio/layerfiles/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)

## Commands

### ./lf build (directory containing Layerfile ...)

Builds one or more Layerfiles into VMs.

### ./lf vm list [not implemented]

Lists VMs on your computer

### ./lf image list [not implemented]

Lists VM images on your computer

### ./lf start [vm] [not implemented]

Starts a VM image to create a new VM

### ./lf stop [vm] [not implemented]

Pauses and stops running VM

### ./lf deploy [image] [not implemented]

Deploys a VM to the cloud


## Learn more

Visit [layerfile.com](https://layerfile.com) for more information