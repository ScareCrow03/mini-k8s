# 先执行 chmod 777 install_nfs_worker.sh

sudo apt-get install rpcbind

if test -d "/srv/mini-k8s"; then
    echo "/srv/mini-k8s exists"
else
    mkdir /srv/mini-k8s
fi

sudo echo "192.168.172.128:/srv/mini-k8s /srv/mini-k8s nfs defaults 0 0" >> /etc/fstab

sudo systemctl restart rpcbind