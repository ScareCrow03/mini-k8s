# 先执行 chmod 777 ./script/install_nfs_worker.sh

sudo apt-get install rpcbind

mkdir /srv/mini-k8s

sudo echo "192.168.172.128:/srv/mini-k8s /srv/mini-k8s nfs defaults 0 0" >> /etc/fstab

sudo systemctl restart rpcbind