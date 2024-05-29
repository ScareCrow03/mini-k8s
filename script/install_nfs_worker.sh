sudo apt-get install rpcbind

mkdir /srv/mini-k8s

su root

sudo echo "192.168.172.128:/srv/mini-k8s /srv/mini-k8s nfs defaults 0 0" >> /etc/fstab

sudo systemctl restart rpcbind