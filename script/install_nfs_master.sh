sudo apt-get install nfs-kernel-server
sudo apt-get install nfs-common
sudo apt-get install rpcbind
sudo apt-get install nfs-server

# mkdir /share
# cd /share
# echo "welcome to NFS">test

mkdir /srv/mini-k8s
cd /srv/mini-k8s
echo "welcome to NFS" > test

sudo systemctl restart nfs-server.service
sudo systemctl restart rpcbind

# cat /etc/exports
# 注意exports里面(rw,sync,no_root_squash)不能加空格

sudo cp /home/zyc/Desktop/mini-k8s/assets/test_persistent/exports /etc/

sudo /etc/init.d/nfs-kernel-server restart

showmount -e