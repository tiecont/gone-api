# Kubernetes & Docker Infrastructure Setup Guide

This README provides a comprehensive guide for setting up a Kubernetes cluster with Docker, Helm, Ingress Controller, Prometheus monitoring, and Rancher UI. It includes steps for preparing system resources, installing required tools, and configuring components across various nodes.

---

## Kubernetes System Overview

| #  | Server           | IP               | Host           | RAM | CPU | Disk | Domain                                      |
|----|------------------|------------------|----------------|-----|-----|------|---------------------------------------------|
| 1  | control-plane    | 192.168.100.110  | control-plane  | 4GB | 4   | 20GB |                                             |
| 2  | worker-node-1    | 192.168.100.111  | worker-1       | 4GB | 4   | 20GB |                                             |
| 3  | worker-node-2    | 192.168.100.112  | worker-2       | 4GB | 4   | 20GB |                                             |
| 4  | loadbalancer     | 192.168.100.115  | ingress-server | 2GB | 1   | 20GB | http://gone-onpre.tiesnode.vn/              |
| 5  | database-server  | 192.168.100.116  | database-server| 2GB | 1   | 20GB |                                             |
---

## Kubernetes Cluster Setup

### 1. System Update & User Creation
```bash
sudo apt update -y && sudo apt upgrade -y
adduser devops
su devops
cd /home/devops
```

### 2. Disable Swap
```bash
sudo swapoff -a
sudo sed -i '/swap.img/s/^/#/' /etc/fstab
```

### 3. Kernel Modules Configuration
```bash
echo -e "overlay\nbr_netfilter" | sudo tee /etc/modules-load.d/containerd.conf
sudo modprobe overlay
sudo modprobe br_netfilter
```

### 4. Network Configuration
```bash
echo "net.bridge.bridge-nf-call-ip6tables = 1" | sudo tee -a /etc/sysctl.d/kubernetes.conf
echo "net.bridge.bridge-nf-call-iptables = 1" | sudo tee -a /etc/sysctl.d/kubernetes.conf
echo "net.ipv4.ip_forward = 1" | sudo tee -a /etc/sysctl.d/kubernetes.conf
sudo sysctl --system
```

### 5. Install Docker & Containerd
```bash
sudo apt install -y ca-certificates curl gnupg lsb-release software-properties-common apt-transport-https
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo \"deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable\" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install -y containerd.io
containerd config default | sudo tee /etc/containerd/config.toml >/dev/null 2>&1
sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/g' /etc/containerd/config.toml
sudo systemctl restart containerd
sudo systemctl enable containerd
```

### 6. Kubernetes Installation
```bash
echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.32/deb/ /" | sudo tee /etc/apt/sources.list.d/kubernetes.list
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.32/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
sudo apt update
sudo apt install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
```

### 7. Kubernetes Initialization (Single/Multi-Master)
```bash
sudo kubeadm init --control-plane-endpoint "192.168.100.110:6443" --upload-certs
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

### 8. Install Calico CNI
```bash
kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.29.3/manifests/calico.yaml
```

### 9. Join Other Nodes
```bash
# On workers and additional masters (modify token, hash, cert accordingly)
sudo kubeadm join 192.168.100.110:6443 --token <token> --discovery-token-ca-cert-hash <sha> --control-plane --certificate-key <cert-key>
```

---

## Docker & Rancher Setup

### Install Docker Compose (Standalone)
```bash
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### Mount Disk for Rancher Data
```bash
sudo mkfs.ext4 -m 0 /dev/sdb
sudo mkdir /data
echo "/dev/sdb  /data  ext4  defaults  0  0" | sudo tee -a /etc/fstab
sudo mount -a
```

### Run Rancher
```bash
docker run --name rancher-server -d --restart=unless-stopped \
  -p 80:80 -p 443:443 \
  -v /data/rancher:/var/lib/rancher \
  --privileged rancher/rancher:latest
```

### Rancher Bootstrap Password
```bash
docker logs rancher-server 2>&1 | grep "Bootstrap Password:"
```

---

## Rancher Notes
- Ensure sufficient CPU & RAM
- Fix issue with:
```bash
kubectl create namespace cattle-fleet-system
```

---

## Install Helm
```bash
wget https://get.helm.sh/helm-v3.17.3-linux-amd64.tar.gz
tar xvf helm-v3.17.3-linux-amd64.tar.gz
sudo mv linux-amd64/helm /usr/bin/
helm version
```

## Ingress Controller Setup (NodePort)
```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm pull ingress-nginx/ingress-nginx
tar -xzf ingress-nginx-*.tgz
vi ingress-nginx/values.yaml
# Edit type: LoadBalancer => NodePort
# Set nodePort http: 30080, https: 30443
kubectl create ns ingress-nginx
helm -n ingress-nginx install ingress-nginx -f ingress-nginx/values.yaml ingress-nginx
```

## Monitoring: Metrics & Prometheus Stack
```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
kubectl patch deployment metrics-server -n kube-system --type='json' -p='[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]'

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install prometheus prometheus-community/kube-prometheus-stack --namespace monitoring --create-namespace
```

---

## Setup Nginx LoadBalancer
```bash
sudo apt install nginx
sudo nano /etc/nginx/conf.d/gone

# Add the following:
upstream my_servers {
    server 192.168.100.110:30080;
    server 192.168.100.111:30080;
    server 192.168.100.112:30080;
}

server {
    listen 80;

    location / {
        proxy_pass http://my_servers;
        proxy_redirect off;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## Setup MySQL Database Server
```bash
sudo apt update && sudo apt upgrade -y
sudo apt install mysql-server -y
sudo systemctl start mysql
sudo systemctl enable mysql
sudo mysql_secure_installation
sudo mysql -u root -p
```

Create DB & User:
```sql
CREATE DATABASE GoneDB;
CREATE USER 'admin'@'%' IDENTIFIED BY 'MysqlAdmin';
GRANT ALL PRIVILEGES ON GoneDB.* TO 'admin'@'%';
FLUSH PRIVILEGES;
```

Configure Remote Access:
```bash
sudo nano /etc/mysql/mysql.conf.d/mysqld.cnf
# Change:
bind-address = 0.0.0.0
mysqlx-bind-address = 0.0.0.0

sudo ufw allow 3306/tcp
sudo ufw allow 22
sudo ufw reload
sudo systemctl restart mysql
```

---

## DockerHub Token
```
dckr_pat_8NEV8QTh3UKrcBYPKYSNArElSA0
```

---

Happy Clustering & Shipping 🚀

