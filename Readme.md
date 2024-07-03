# Presetter - add ResourcePresets to your K8s deployments
This is a custom controller, that includes ResourcePreset CRD's, which allow to create dynamic memory and CPU limits/requests configurations and automatically add them to specifically labeled Deploys/Pods.  

## Installation
For DEV
```
make manifests
make install
make run
k create -f config/samples/presetter_v1_resourcepreset.yaml
```

## Usage
```
apiVersion: v1
kind: Pod
metadata:
  name: nginx-test-pod
  namespace: presets
  labels:
    app: nginx
    presetter.xamma.dev/preset: "minimal"  # Beispiel-Label, das für deine Ressourcenvorgaben verwendet werden könnte
spec:
  containers:
  - name: nginx
    image: nginx:latest
    ports:
    - containerPort: 80

```