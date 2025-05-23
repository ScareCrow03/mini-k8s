# 先执行 chmod 777 ./script/install_nfs_master.sh

sudo apt-get install nfs-kernel-server
sudo apt-get install nfs-common
sudo apt-get install rpcbind
sudo apt-get install nfs-server

# mkdir /share
# cd /share
# echo "welcome to NFS">test

mkdir /srv/mini-k8s
sudo chmod 777 /srv/mini-k8s/
sudo echo "welcome to mini-k8s NFS" > test
# sudo bash -c "echo \"welcome to mini-k8s NFS\" > test"

sudo systemctl restart nfs-server.service
sudo systemctl restart rpcbind

# cat /etc/exports
# 注意exports里面(rw,sync,no_root_squash)不能加空格
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

sudo cp ${SCRIPT_DIR}/../assets/test_persistent/exports /etc/

sudo /etc/init.d/nfs-kernel-server restart

showmount -e