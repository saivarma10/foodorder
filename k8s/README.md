# Kubernetes Deployment for Food Order System

This directory contains Kubernetes manifests for deploying the Food Order microservices application.

## Architecture

The application consists of:
- **Go Service**: Main application service (HTTP + gRPC)
- **Java Service**: Backend service
- **Python Service**: Delivery service
- **PostgreSQL**: Database
- **Kafka + Zookeeper**: Message queue

## Prerequisites

- Kubernetes cluster (minikube, kind, GKE, EKS, AKS, etc.)
- kubectl configured to connect to your cluster
- Docker images pushed to Docker Hub:
  - `varmasai10/go:latest`
  - `varmasai10/java:latest`
  - `varmasai10/python:latest`

## Deployment Order

Deploy the services in the following order to ensure dependencies are met:

### 1. Deploy PostgreSQL
```bash
kubectl apply -f postgres-deployment.yaml
```

Wait for PostgreSQL to be ready:
```bash
kubectl wait --for=condition=ready pod -l app=postgres --timeout=120s
```

### 2. Deploy Kafka & Zookeeper
```bash
kubectl apply -f kafka-deployment.yaml
```

Wait for Kafka to be ready:
```bash
kubectl wait --for=condition=ready pod -l app=kafka --timeout=120s
```

### 3. Deploy Python Delivery Service
```bash
kubectl apply -f python-deployment.yaml
```

Wait for Python service to be ready:
```bash
kubectl wait --for=condition=ready pod -l app=python-delivery --timeout=60s
```

### 4. Deploy Java Service
```bash
kubectl apply -f java-deployment.yaml
```

Wait for Java service to be ready:
```bash
kubectl wait --for=condition=ready pod -l app=java-service --timeout=120s
```

### 5. Deploy Go Service
```bash
kubectl apply -f go-deployment.yaml
```

## Quick Deployment (All at Once)

```bash
kubectl apply -f .
```

## Verify Deployment

Check all pods are running:
```bash
kubectl get pods
```

Check all services:
```bash
kubectl get services
```

## Access the Application

### Get Go Service URL (LoadBalancer)
```bash
# For cloud providers (AWS, GCP, Azure)
kubectl get service go-service

# For minikube
minikube service go-service --url

# For port-forward
kubectl port-forward service/go-service 8081:8081
kubectl port-forward service/go-service 51058:51058
```

### Access Individual Services (for debugging)
```bash
# Python service
kubectl port-forward service/python-delivery 8010:8010

# Java service
kubectl port-forward service/java-service 8085:8085

# PostgreSQL
kubectl port-forward service/postgres 5432:5432

# Kafka
kubectl port-forward service/kafka 9092:9092
```

## Scaling

Scale individual services:
```bash
kubectl scale deployment go-service --replicas=3
kubectl scale deployment java-service --replicas=3
kubectl scale deployment python-delivery --replicas=3
```

## Logs

View logs for each service:
```bash
kubectl logs -l app=go-service -f
kubectl logs -l app=java-service -f
kubectl logs -l app=python-delivery -f
kubectl logs -l app=postgres -f
kubectl logs -l app=kafka -f
```

## Cleanup

Remove all resources:
```bash
kubectl delete -f .
```

Or delete individually:
```bash
kubectl delete -f go-deployment.yaml
kubectl delete -f java-deployment.yaml
kubectl delete -f python-deployment.yaml
kubectl delete -f kafka-deployment.yaml
kubectl delete -f postgres-deployment.yaml
```

## Configuration

### Database Credentials
Edit `postgres-deployment.yaml` to change database credentials:
- ConfigMap: `postgres-config`
- Secret: `postgres-secret`

### Environment Variables
Each deployment file contains environment variables that can be customized:
- **go-deployment.yaml**: DB connection, Kafka broker, Java service URL
- **java-deployment.yaml**: DB connection, Python service URL
- **python-deployment.yaml**: Flask environment

### Resource Limits
Adjust CPU and memory limits in each deployment file under `resources` section.

## Troubleshooting

### Pods not starting
```bash
kubectl describe pod <pod-name>
kubectl logs <pod-name>
```

### Service connectivity issues
```bash
# Test from within the cluster
kubectl run -it --rm debug --image=busybox --restart=Never -- sh
# Then: wget -O- http://go-service:8081/health
```

### Database connection issues
```bash
# Check PostgreSQL is ready
kubectl exec -it <postgres-pod-name> -- psql -U postgres -c "SELECT 1"
```

## Production Considerations

For production deployments, consider:

1. **Secrets Management**: Use external secret managers (AWS Secrets Manager, HashiCorp Vault, etc.)
2. **Persistent Storage**: Configure proper PersistentVolumes with backup strategies
3. **Ingress Controller**: Add an Ingress resource instead of LoadBalancer for better routing
4. **Health Checks**: Implement proper HTTP health endpoints for all services
5. **Resource Limits**: Fine-tune based on actual usage patterns
6. **High Availability**: Run multiple replicas and use pod anti-affinity rules
7. **Monitoring**: Add Prometheus/Grafana for monitoring
8. **Logging**: Configure centralized logging (ELK, Loki, etc.)
9. **Network Policies**: Implement network policies for security
10. **TLS/SSL**: Enable TLS for external-facing services
