package main

const SERVICE_TEMPLATE = `
apiVersion: v1
kind: Service
metadata:
 name: {{ .Metadata.Name }}
spec:
 selector:
   app: {{ .Metadata.Name }}
 ports:
 - port: {{ .Spec.Port }}
   targetPort: {{ .Spec.Port }}`

const DEPLOYMENT_TEMPLATE = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Metadata.Name }}
spec:
  selector:
    matchLabels:
      app: {{ .Metadata.Name }}
  template:
    metadata:
      labels:
        app: {{ .Metadata.Name }}
    spec:
      containers:
      - name: {{ .Metadata.Name }}
        image: {{ .Spec.Image }}
        resources:
          limits:
            {{ if eq .Spec.Size "large" }}
            memory: "1024"
            cpu: "1"
            {{ else if eq .Spec.Size "medium" }}
            memory: "512Mi"
            cpu: "500m"
            {{ else }}
            memory: "256Mi"
            cpu: "250m"
            {{ end }}
        ports:
        - containerPort: {{ .Spec.Port }}`
