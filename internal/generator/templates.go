package generator

const ChartTemplate = `apiVersion: v2
name: {{ .AppName }}
description: A Helm chart for Kubernetes deployment of {{ .AppName }}
type: application
version: 0.1.0
appVersion: "1.0.0"
`

const ValuesTemplate = `# Default values for {{ .AppName }}.
# This is a YAML-formatted file.

replicaCount: {{ .ReplicaCount }}

image:
  repository: {{ .ImageRepository }}
  pullPolicy: IfNotPresent
  tag: "{{ .ImageTag }}"

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

serviceAccount:
  create: true
  annotations: {}
  name: ""

podAnnotations: {}

{{- if eq .TemplateQuality "enterprise" }}
podSecurityContext:
  fsGroup: 2000

securityContext:
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 1000
{{- else if eq .TemplateQuality "production" }}
podSecurityContext: {}

securityContext:
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 1000
{{- else }}
podSecurityContext: {}
securityContext: {}
{{- end }}

service:
  type: ClusterIP
  port: {{ .Port }}

{{- if .GenerateIngress }}
ingress:
  enabled: true
  className: ""
  annotations: {}
  hosts:
    - host: {{ .AppName }}.local
      paths:
        - path: /
          pathType: ImplementationSpecific
  tls: []
{{- end }}

{{- if .GenerateGateway }}
gateway:
  enabled: true
  name: "{{ .AppName }}-gateway"
  namespace: "{{ .Namespace }}"
  listenerName: "http"
  port: 80
  protocol: HTTP
  hostname: "{{ .AppName }}.local"
{{- end }}

{{- if .GenerateConfigMap }}
configMap:
  enabled: true
  data:
    APP_ENV: "production"
    LOG_LEVEL: "info"
{{- end }}

{{- if .GenerateExternalSecret }}
externalSecret:
  enabled: true
  backend: "{{ .SecretBackend }}" # vault, aws, gcp, azure
  secretStoreName: "{{ .AppName }}-store"
  secretStoreKind: "SecretStore"
  refreshInterval: "1h"
  data:
    - secretKey: "DATABASE_URL"
      remoteKey: "prod/db/url"
      property: "connectionString"
{{- end }}

{{- if .GenerateHPA }}
autoscaling:
  enabled: true
  minReplicas: {{ .HPAMinReplicas }}
  maxReplicas: {{ .HPAMaxReplicas }}
  targetCPUUtilizationPercentage: 80
  targetMemoryUtilizationPercentage: 80
{{- end }}

{{- if .GenerateVPA }}
vpa:
  enabled: true
  updateMode: "Auto"
{{- end }}

{{- if .GenerateKEDA }}
keda:
  enabled: true
  minReplicaCount: 1
  maxReplicaCount: 10
  cooldownPeriod: 300
  pollingInterval: 30
  triggers:
    - type: prometheus
      metadata:
        serverAddress: http://prometheus-k8s.monitoring.svc.cluster.local:9090
        metricName: http_requests_total
        query: sum(rate(http_requests_total{job="{{ .AppName }}"}[2m]))
        threshold: '100'
{{- end }}

{{- if .GenerateServiceMonitor }}
serviceMonitor:
  enabled: true
  interval: "30s"
  path: "/metrics"
  port: "http"
{{- end }}

{{- if .GeneratePDB }}
pdb:
  enabled: true
  minAvailable: 1
{{- end }}

{{- if .GenerateStatefulSet }}
statefulset:
  serviceName: "{{ .AppName }}-headless"
  storageClass: "standard"
  storageSize: "10Gi"
{{- end }}

{{- if .GenerateCronJob }}
cronjob:
  schedule: "*/5 * * * *"
  concurrencyPolicy: "Forbid"
  failedJobsHistoryLimit: 1
  successfulJobsHistoryLimit: 3
  command: ["/bin/sh", "-c", "echo task completed"]
{{- end }}

{{- if .GenerateArgoCD }}
argocd:
  enabled: true
  project: "default"
  repoURL: "https://github.com/ihyamarsdev/kgen-gitops.git"
  targetRevision: "HEAD"
  path: "charts/{{ .AppName }}"
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
{{- end }}

{{- if .GenerateIstio }}
istio:
  enabled: true
  gatewayName: "mesh"
  hosts:
    - "{{ .AppName }}.local"
{{- end }}

{{- if .GeneratePVC }}
pvc:
  enabled: true
  accessMode: "ReadWriteOnce"
  size: "10Gi"
  storageClass: "standard"
{{- end }}

{{- if .GenerateNetworkPolicy }}
networkPolicy:
  enabled: true
{{- end }}

{{- if eq .TemplateQuality "enterprise" }}
resources:
  limits:
    cpu: 1000m
    memory: 1024Mi
  requests:
    cpu: 500m
    memory: 512Mi
{{- else if eq .TemplateQuality "production" }}
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi
{{- else }}
resources: {}
{{- end }}

nodeSelector: {}
tolerations: []
affinity: {}
topologySpreadConstraints: []

{{- if or (eq .TemplateQuality "production") (eq .TemplateQuality "enterprise") }}
livenessProbe:
  httpGet:
    path: /
    port: http
  initialDelaySeconds: 15
  periodSeconds: 20

readinessProbe:
  httpGet:
    path: /
    port: http
  initialDelaySeconds: 5
  periodSeconds: 10
{{- else }}
livenessProbe: {}
readinessProbe: {}
{{- end }}
`

const HelpersTemplate = `{{/*
Expand the name of the chart.
*/}}
{{- define "kgen.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kgen.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "kgen.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kgen.labels" -}}
helm.sh/chart: {{ include "kgen.chart" . }}
{{ include "kgen.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kgen.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kgen.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "kgen.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "kgen.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
`

const DeploymentTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "kgen.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "kgen.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      {{- with .Values.podAnnotations }}
      annotations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      labels:
        {{- include "kgen.selectorLabels" . | nindent 8 }}
    spec:
      {{- with .Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      serviceAccountName: {{ include "kgen.serviceAccountName" . }}
      securityContext:
        {{- toYaml .Values.podSecurityContext | nindent 8 }}
      containers:
        - name: {{ .Chart.Name }}
          securityContext:
            {{- toYaml .Values.securityContext | nindent 12 }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - name: http
              containerPort: {{ .Values.service.port }}
              protocol: TCP
          {{- if .Values.livenessProbe }}
          livenessProbe:
            {{- toYaml .Values.livenessProbe | nindent 12 }}
          {{- end }}
          {{- if .Values.readinessProbe }}
          readinessProbe:
            {{- toYaml .Values.readinessProbe | nindent 12 }}
          {{- end }}
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
      {{- with .Values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
`

const ServiceTemplate = `apiVersion: v1
kind: Service
metadata:
  name: {{ include "kgen.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
spec:
  type: {{ .Values.service.type }}
  ports:
    - port: {{ .Values.service.port }}
      targetPort: http
      protocol: TCP
      name: http
  selector:
    {{- include "kgen.selectorLabels" . | nindent 4 }}
`

const IngressTemplate = `{{- if .Values.ingress.enabled -}}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ include "kgen.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
  {{- with .Values.ingress.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
spec:
  {{- if .Values.ingress.className }}
  ingressClassName: {{ .Values.ingress.className }}
  {{- end }}
  {{- if .Values.ingress.tls }}
  tls:
    {{- range .Values.ingress.tls }}
    - hosts:
        {{- range .hosts }}
        - {{ . | quote }}
        {{- end }}
      secretName: {{ .secretName }}
    {{- end }}
  {{- end }}
  rules:
    {{- range .Values.ingress.hosts }}
    - host: {{ .host | quote }}
      http:
        paths:
          {{- range .paths }}
          - path: {{ .path }}
            pathType: {{ .pathType }}
            backend:
              service:
                name: {{ include "kgen.fullname" $ }}
                port:
                  number: {{ $.Values.service.port }}
          {{- end }}
    {{- end }}
{{- end }}
`

const GatewayTemplate = `{{- if .Values.gateway.enabled -}}
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: {{ .Values.gateway.name | default (include "kgen.fullname" .) }}
  namespace: {{ .Values.gateway.namespace | default .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
spec:
  gatewayClassName: {{ .Values.gateway.name | default "gateway-class" }}
  listeners:
    - name: {{ .Values.gateway.listenerName | default "http" }}
      port: {{ .Values.gateway.port | default 80 }}
      protocol: {{ .Values.gateway.protocol | default "HTTP" }}
      allowedRoutes:
        namespaces:
          from: Same
{{- end }}
`

const HTTPRouteTemplate = `{{- if .Values.gateway.enabled -}}
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: {{ include "kgen.fullname" . }}-route
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
spec:
  parentRefs:
    - name: {{ .Values.gateway.name | default (include "kgen.fullname" .) }}
  rules:
    - backendRefs:
        - name: {{ include "kgen.fullname" . }}
          port: {{ .Values.service.port }}
{{- end }}
`

const ConfigMapTemplate = `{{- if .Values.configMap.enabled -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "kgen.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
data:
  {{- toYaml .Values.configMap.data | nindent 2 }}
{{- end }}
`

const ExternalSecretTemplate = `{{- if .Values.externalSecret.enabled -}}
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: {{ include "kgen.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
spec:
  refreshInterval: {{ .Values.externalSecret.refreshInterval | quote }}
  secretStoreRef:
    name: {{ .Values.externalSecret.secretStoreName }}
    kind: {{ .Values.externalSecret.secretStoreKind }}
  target:
    name: {{ include "kgen.fullname" . }}-synced
    creationPolicy: Owner
  data:
    {{- range .Values.externalSecret.data }}
    - secretKey: {{ .secretKey }}
      remoteRef:
        key: {{ .remoteKey }}
        property: {{ .property }}
    {{- end }}
{{- end }}
`

const HPATemplate = `{{- if .Values.autoscaling.enabled -}}
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{ include "kgen.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "kgen.fullname" . }}
  minReplicas: {{ .Values.autoscaling.minReplicas }}
  maxReplicas: {{ .Values.autoscaling.maxReplicas }}
  metrics:
    {{- if .Values.autoscaling.targetCPUUtilizationPercentage }}
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: {{ .Values.autoscaling.targetCPUUtilizationPercentage }}
    {{- end }}
    {{- if .Values.autoscaling.targetMemoryUtilizationPercentage }}
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: {{ .Values.autoscaling.targetMemoryUtilizationPercentage }}
    {{- end }}
{{- end }}
`

const VPATemplate = `{{- if .Values.vpa.enabled -}}
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: {{ include "kgen.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
spec:
  targetRef:
    apiVersion: apps/v1
    {{- if .Values.statefulset }}
    kind: StatefulSet
    {{- else }}
    kind: Deployment
    {{- end }}
    name: {{ include "kgen.fullname" . }}
  updatePolicy:
    updateMode: {{ .Values.vpa.updateMode | quote }}
{{- end }}
`

const ScaledObjectTemplate = `{{- if .Values.keda.enabled -}}
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: {{ include "kgen.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    {{- if .Values.statefulset }}
    kind: StatefulSet
    {{- else }}
    kind: Deployment
    {{- end }}
    name: {{ include "kgen.fullname" . }}
  minReplicaCount: {{ .Values.keda.minReplicaCount }}
  maxReplicaCount: {{ .Values.keda.maxReplicaCount }}
  cooldownPeriod: {{ .Values.keda.cooldownPeriod }}
  pollingInterval: {{ .Values.keda.pollingInterval }}
  triggers:
    {{- toYaml .Values.keda.triggers | nindent 4 }}
{{- end }}
`

const StatefulSetTemplate = `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: {{ include "kgen.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
spec:
  serviceName: {{ .Values.statefulset.serviceName | default (include "kgen.fullname" .) }}
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      {{- include "kgen.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "kgen.selectorLabels" . | nindent 8 }}
    spec:
      containers:
        - name: {{ .Chart.Name }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - name: http
              containerPort: {{ .Values.service.port }}
              protocol: TCP
          volumeMounts:
            - name: data
              mountPath: /data
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: [ "ReadWriteOnce" ]
        storageClassName: {{ .Values.statefulset.storageClass | quote }}
        resources:
          requests:
            storage: {{ .Values.statefulset.storageSize | default "10Gi" }}
`

const CronJobTemplate = `apiVersion: batch/v1
kind: CronJob
metadata:
  name: {{ include "kgen.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
spec:
  schedule: {{ .Values.cronjob.schedule | quote }}
  concurrencyPolicy: {{ .Values.cronjob.concurrencyPolicy }}
  failedJobsHistoryLimit: {{ .Values.cronjob.failedJobsHistoryLimit }}
  successfulJobsHistoryLimit: {{ .Values.cronjob.successfulJobsHistoryLimit }}
  jobTemplate:
    spec:
      template:
        spec:
          restartPolicy: OnFailure
          containers:
            - name: {{ .Chart.Name }}
              image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
              command:
                {{- toYaml .Values.cronjob.command | nindent 16 }}
`

const ArgoApplicationTemplate = `{{- if .Values.argocd.enabled -}}
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {{ include "kgen.fullname" . }}
  namespace: argocd
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
spec:
  project: {{ .Values.argocd.project | quote }}
  source:
    repoURL: {{ .Values.argocd.repoURL | quote }}
    targetRevision: {{ .Values.argocd.targetRevision | quote }}
    path: {{ .Values.argocd.path | quote }}
    helm:
      valueFiles:
        - values.yaml
  destination:
    server: https://kubernetes.default.svc
    namespace: {{ .Release.Namespace }}
  syncPolicy:
    {{- toYaml .Values.argocd.syncPolicy | nindent 4 }}
{{- end }}
`

const IstioVirtualServiceTemplate = `{{- if .Values.istio.enabled -}}
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: {{ include "kgen.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
spec:
  hosts:
    {{- toYaml .Values.istio.hosts | nindent 4 }}
  gateways:
    - {{ .Values.istio.gatewayName }}
  http:
    - route:
        - destination:
            host: {{ include "kgen.fullname" . }}
            port:
              number: {{ .Values.service.port }}
{{- end }}
`

const PdbTemplate = `{{- if .Values.pdb.enabled -}}
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: {{ include "kgen.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
spec:
  {{- if .Values.pdb.minAvailable }}
  minAvailable: {{ .Values.pdb.minAvailable }}
  {{- end }}
  {{- if .Values.pdb.maxUnavailable }}
  maxUnavailable: {{ .Values.pdb.maxUnavailable }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "kgen.selectorLabels" . | nindent 6 }}
{{- end }}
`

const ServiceMonitorTemplate = `{{- if .Values.serviceMonitor.enabled -}}
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: {{ include "kgen.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
spec:
  selector:
    matchLabels:
      {{- include "kgen.selectorLabels" . | nindent 6 }}
  endpoints:
    - port: {{ .Values.serviceMonitor.port }}
      interval: {{ .Values.serviceMonitor.interval }}
      path: {{ .Values.serviceMonitor.path }}
{{- end }}
`

const NetworkPolicyTemplate = `{{- if .Values.networkPolicy.enabled -}}
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: {{ include "kgen.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
spec:
  podSelector:
    matchLabels:
      {{- include "kgen.selectorLabels" . | nindent 6 }}
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - podSelector: {}
      ports:
        - protocol: TCP
          port: {{ .Values.service.port }}
  egress:
    - to:
        - ipBlock:
            cidr: 0.0.0.0/0
{{- end }}
`

const PVCTemplate = `{{- if .Values.pvc.enabled -}}
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{ include "kgen.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kgen.labels" . | nindent 4 }}
spec:
  accessModes:
    - {{ .Values.pvc.accessMode }}
  storageClassName: {{ .Values.pvc.storageClass }}
  resources:
    requests:
      storage: {{ .Values.pvc.size }}
{{- end }}
`
