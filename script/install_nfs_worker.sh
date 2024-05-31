# 先执行 chmod 777 ./script/install_nfs_worker.sh
# 记得改为主机ip

sudo apt-get install rpcbind

mkdir /srv/mini-k8s
sudo chmod 777 /srv/mini-k8s/

mount 192.168.1.5:/srv/mini-k8s /srv/mini-k8s
sudo echo "192.168.172.128:/srv/mini-k8s /srv/mini-k8s nfs defaults 0 0" >> /etc/fstab

sudo systemctl restart rpcbind