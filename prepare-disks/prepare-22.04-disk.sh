#!/usr/bin/env bash

set -eu -o pipefail

cat << EOF | sudo chroot ubuntu-22.04
echo -e 'password\npassword' | passwd
rm -f /etc/legal
chmod -x /etc/update-motd.d/*
rm -rf /var/lib/apt/lists/*
chmod a-x /usr/lib/ubuntu-release-upgrader/release-upgrade-motd
echo fs.inotify.max_user_watches=524288 > /etc/sysctl.d/99-ubuntu-go.conf

apt-get remove -y fwupd cloud-guest-utils cloud-init* initramfs* ssh* openssh*
apt-get autoremove -y

cat <<ZZZZ >/etc/apt/apt.conf.d/90assumeyes
APT::Get::Assume-Yes "true";
APT::Periodic::Update-Package-Lists "0";
APT::Periodic::Unattended-Upgrade "0";
ZZZZ

rm /etc/resolv.conf
echo 'nameserver 10.111.1.1' > /etc/resolv.conf

echo 'ubuntu2204-layerfile' > /etc/hostname

rm -f /lib/systemd/system/sysinit.target.wants/systemd-random-seed.service
echo '127.0.1.1    ubuntu2204-layerfile' >> /etc/hosts
rm -f /root/.bash_history /boot/*

EOF
