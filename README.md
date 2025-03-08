# Kubernetes Setup Guide (Fixing kubelet issues to Deployment Success)

This guide provides step-by-step instructions to set up a Kubernetes cluster, fix `kubelet` issues, install networking (Flannel), and deploy applications successfully.

## 🛠 1. Fixing kubelet Issues

If `kubelet` is failing to start, try the following:

### 🔹 Check kubelet status
```bash
systemctl status kubelet
```
If it shows errors, proceed with the following fixes:

### 🔹 Restart kubelet
```bash
systemctl restart kubelet
systemctl enable kubelet
```

### 🔹 Reset Kubernetes (if necessary)
```bash
kubeadm reset -f
systemctl restart kubelet
```

## 🚀 2. Reinitialize Kubernetes Cluster
```bash
kubeadm init --pod-network-cidr=10.244.0.0/16
```

### 🔹 Set up kubeconfig for `kubectl`
```bash
mkdir -p $HOME/.kube
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
chown $(id -u):$(id -g) $HOME/.kube/config
```

## 🌐 3. Install Flannel Networking
```bash
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
```

### 🔹 Verify Flannel is running
```bash
kubectl get pods -n kube-flannel
```
If any pod is in `Error` state, check logs:
```bash
kubectl logs -n kube-flannel <pod-name>
```

## 🏗 4. Remove Taint to Allow Workloads on Control Plane (Optional for Single-Node Setup)
```bash
kubectl taint nodes --all node-role.kubernetes.io/control-plane-
```

## 🏗 5. Deploy Applications

### 🔹 Deploy MySQL
```yaml
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mysql
  template:
    metadata:
      labels:
        app: mysql
    spec:
      containers:
      - name: mysql
        image: mysql:5.7
        env:
        - name: MYSQL_ROOT_PASSWORD
          value: rootpassword
EOF
```

### 🔹 Deploy Spring Boot Application
```yaml
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: spring-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: spring-app
  template:
    metadata:
      labels:
        app: spring-app
    spec:
      containers:
      - name: spring-app
        image: spring-app:latest
        ports:
        - containerPort: 8080
EOF
```

## 🔍 6. Check Deployment Status
```bash
kubectl get pods -A
```

## 🌎 7. Expose Services
```bash
kubectl expose deployment mysql --type=ClusterIP --port=3308
kubectl expose deployment spring-app --type=NodePort --name=spring-service
```

### 🔹 Get service details
```bash
kubectl get svc
```

## 🎉 8. Access Application

Find the NodePort assigned to `spring-service`:
```bash
kubectl get svc spring-service
```
Access it using:
```bash
http://<NODE_IP>:<NODE_PORT>
```

🚀 Your Kubernetes cluster is now fully set up with a working application! 🎉

