
# Kubernetes Cluster Setup Guide

This guide provides a step-by-step process for setting up a Kubernetes cluster on Ubuntu servers, including configuration of Docker, containerd, and Calico for networking. The setup supports both single-master and multi-master (HA) configurations.

## Table of Contents
1. [System Update and User Creation](#system-update-and-user-creation)
2. [Swap Configuration](#swap-configuration)
3. [Kernel Modules Configuration](#kernel-modules-configuration)
4. [Network Configuration](#network-configuration)
5. [Docker and Kubernetes Installation](#docker-and-kubernetes-installation)
6. [Kubernetes Cluster Initialization](#kubernetes-cluster-initialization)
7. [Resetting Kubernetes Cluster](#resetting-kubernetes-cluster)
8. [Multi-Master Configuration](#multi-master-configuration)
9. [Conclusion](#conclusion)

---

## System Update and User Creation

1. **Update and upgrade the system**:
   ```bash
   sudo apt update -y && sudo apt upgrade -y
   ```

2. **Create `devops` user and switch to it**:
   ```bash
   adduser devops
   su devops
   cd /home/devops
   ```

---

## Swap Configuration

1. **Turn off swap**:
   ```bash
   sudo swapoff -a
   ```

2. **Prevent swap from re-enabling after reboot**:
   ```bash
   sudo sed -i '/swap.img/s/^/#/' /etc/fstab
   ```

---

## Kernel Modules Configuration

1. **Configure the kernel modules**:
   Edit the file `/etc/modules-load.d/containerd.conf` and add the following:
   ```text
   overlay
   br_netfilter
   ```

2. **Load the kernel modules**:
   ```bash
   sudo modprobe overlay
   sudo modprobe br_netfilter
   ```

---

## Network Configuration

1. **Set up sysctl parameters for Kubernetes networking**:
   ```bash
   echo "net.bridge.bridge-nf-call-ip6tables = 1" | sudo tee -a /etc/sysctl.d/kubernetes.conf
   echo "net.bridge.bridge-nf-call-iptables = 1" | sudo tee -a /etc/sysctl.d/kubernetes.conf
   echo "net.ipv4.ip_forward = 1" | sudo tee -a /etc/sysctl.d/kubernetes.conf
   ```

2. **Apply sysctl settings**:
   ```bash
   sudo sysctl --system
   ```

---

## Docker and Kubernetes Installation

1. **Install necessary packages and add Docker repository**:
   ```bash
   sudo apt install -y curl gnupg2 software-properties-common apt-transport-https ca-certificates
   sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmour -o /etc/apt/trusted.gpg.d/docker.gpg
   sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
   ```

2. **Install containerd**:
   ```bash
   sudo apt update -y
   sudo apt install -y containerd.io
   ```

3. **Configure containerd**:
   ```bash
   containerd config default | sudo tee /etc/containerd/config.toml >/dev/null 2>&1
   sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/g' /etc/containerd/config.toml
   ```

4. **Restart and enable containerd**:
   ```bash
   sudo systemctl restart containerd
   sudo systemctl enable containerd
   ```

5. **Add Kubernetes repository**:
   ```bash
   echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.30/deb/ /" | sudo tee /etc/apt/sources.list.d/kubernetes.list
   curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.30/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
   ```

6. **Install Kubernetes packages**:
   ```bash
   sudo apt update -y
   sudo apt install -y kubelet kubeadm kubectl
   sudo apt-mark hold kubelet kubeadm kubectl
   ```

---

## Kubernetes Cluster Initialization

### Reset Cluster (if needed)

1. **Reset Kubernetes cluster**:
   ```bash
   sudo kubeadm reset -f
   sudo rm -rf /var/lib/etcd
   sudo rm -rf /etc/kubernetes/manifests/*
   ```

### Single-Master Setup (1 master, 2 workers)

1. **On master node (k8s-master-1)**, initialize Kubernetes:
   ```bash
   sudo kubeadm init
   ```

2. **Configure kubectl for the master node**:
   ```bash
   mkdir -p $HOME/.kube
   sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
   sudo chown $(id -u):$(id -g) $HOME/.kube/config
   ```

3. **Install Calico networking**:
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/calico.yaml
   ```

4. **On worker nodes (k8s-master-2, k8s-master-3)**, join the cluster:
   ```bash
   sudo kubeadm join 192.168.1.111:6443 --token your_token --discovery-token-ca-cert-hash your_sha
   ```

---

## Multi-Master Setup (3 masters)

1. **On master node (k8s-master-1)**, initialize Kubernetes with control-plane endpoint:
   ```bash
   sudo kubeadm init --control-plane-endpoint "192.168.1.111:6443" --upload-certs
   ```

2. **Configure kubectl for the master node**:
   ```bash
   mkdir -p $HOME/.kube
   sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
   sudo chown $(id -u):$(id -g) $HOME/.kube/config
   ```

3. **Install Calico networking**:
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/calico.yaml
   ```

4. **On other master nodes (k8s-master-2, k8s-master-3)**, join the cluster:
   ```bash
   sudo kubeadm join 192.168.1.111:6443 --token your_token --discovery-token-ca-cert-hash your_sha --control-plane --certificate-key your_cert
   ```

---

## Conclusion

With the steps above, you should now have a working Kubernetes cluster configured on Ubuntu servers. Depending on your architecture, you can opt for a single master or multi-master setup for high availability. Ensure you verify the status of your nodes and pods using `kubectl get nodes` and `kubectl get pods -A`.

Happy Kubernetes managing! 🚀

# 🐳 Docker & Rancher Setup Guide (Ubuntu)

This guide walks you through installing Docker & Docker Compose, mounting a new disk, and running Rancher server using Docker.

---

## ✅ Step 1: Install Docker Engine

Update packages and install dependencies:

```bash
sudo apt update
sudo apt install -y ca-certificates curl gnupg lsb-release
```

Add Docker’s official GPG key:

```bash
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | \
  sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
```

Add Docker repository:

```bash
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
```

Install Docker:

```bash
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

---

## ✅ Step 2: Install `docker-compose` CLI (Standalone)

```bash
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
  -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

Verify installation:

```bash
docker --version
docker-compose --version
```

---

## ✅ Step 3: Run Docker Without `sudo` (Optional)

```bash
sudo usermod -aG docker $USER
newgrp docker
```

Then test:

```bash
docker ps
```

---

## ✅ Step 4: Mount and Prepare Disk for Rancher

Format and mount disk (replace `/dev/sdb` with your device):

```bash
sudo mkfs.ext4 -m 0 /dev/sdb
sudo mkdir /data
echo "/dev/sdb  /data  ext4  defaults  0  0" | sudo tee -a /etc/fstab
sudo mount -a
sudo df -h
```

---

## ✅ Step 5: Run Rancher Server via Docker

Pull and run Rancher with data volume:

```bash
docker run --name rancher-server -d --restart=unless-stopped \
  -p 80:80 -p 443:443 \
  -v /data/rancher:/var/lib/rancher \
  --privileged rancher/rancher:latest
```

---

## 🔐 Get Rancher Bootstrap Password

```bash
docker logs rancher-server 2>&1 | grep "Bootstrap Password:"
```

Use this password to log into Rancher for the first time via the web UI.

---

## 🎉 Done!

You now have Docker, Docker Compose, disk storage, and Rancher running!

