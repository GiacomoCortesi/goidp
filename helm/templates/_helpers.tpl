{{/*
Expand the name of the chart.
*/}}
{{- define "goidp.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "goidp.fullname" -}}
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

{{- define "goidp.goidprsa" -}}
{{- if .Values.jwt.secretName }}
{{- .Values.jwt.secretName | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s" (include "goidp.fullname" .) }}-rsa
{{- end }}
{{- end }}

{{- define "goidp.secretname" -}}
{{- if .Values.database.dbSecretOverride }}
{{- .Values.database.dbSecretOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s" (include "goidp.fullname" .) }}
{{- end }}
{{- end }}

{{- define "goidp.secretkeysname" -}}
{{- printf "%s" (include "goidp.fullname" .) }}-keys
{{- end }}

{{- define "goidp.gensecret" -}}
{{- $secret := lookup "v1" "Secret" .Release.Namespace (include "goidp.secretname" .) -}}
{{- if $secret -}}
{{/* Reusing value of secret if exist */}}
password: {{ $secret.data.password }}
postgres-password: {{ index $secret.data "postgres-password" }}
{{- else -}}
{{/*
    add new data
*/}}
password: {{ randAlphaNum 24 | b64enc | quote }}
postgres-password: {{ randAlphaNum 24 | b64enc | quote }}
{{- end -}}
{{- end -}}


{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "goidp.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "goidp.labels" -}}
helm.sh/chart: {{ include "goidp.chart" . }}
{{ include "goidp.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "goidp.selectorLabels" -}}
app.kubernetes.io/name: {{ include "goidp.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: "idp"
{{- end }}

{{/*
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "goidp.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "goidp.fullname" .) .Values.serviceAccount.name -}}
{{- else }}
{{- default "default" .Values.serviceAccount.name -}}
{{- end }}
{{- end }}

{{/*
goidp imagePullSecrets
*/}}
{{- define "goidp.pullSecrets" -}}
  {{- $pullSecrets := list }}
  {{- if .global }}
    {{- range .global.imagePullSecrets -}}
      {{- $pullSecrets = append $pullSecrets . -}}
    {{- end -}}
  {{- end -}}
  {{- range .images -}}
    {{- range .pullSecrets -}}
      {{- $pullSecrets = append $pullSecrets . -}}
    {{- end -}}
  {{- end -}}
  {{- if (not (empty $pullSecrets)) -}}
imagePullSecrets:
    {{- range $pullSecrets }}
  - name: {{ . }}
    {{- end }}
  {{- end }}
{{- end -}}

{{/*
Return the proper Docker Image Registry Secret Names
*/}}
{{- define "goidp.imagePullSecrets" -}}
{{- include "goidp.pullSecrets" (dict "images" (list .Values.image ) "global" .Values.global) -}}
{{- end -}}


{{/*
goidp image
*/}}
{{- define "goidp.image" -}}
{{- printf "%s/%s:%s" ( default .Values.image.registry .Values.global.imageRegistry) .Values.image.repository ( default .Chart.AppVersion .Values.image.tag ) }}
{{- end }}

{{/*
goidp init container image
*/}}
{{- define "goidp.initContainerImage" -}}
{{- printf "%s/%s:%s" ( default .Values.initContainer.image.registry .Values.global.imageRegistry) .Values.initContainer.image.repository ( default .Chart.AppVersion .Values.initContainer.image.tag ) }}
{{- end }}

{{/*
WindRiver Cloud Platform clusters only support an older API for certificates, detect the supported
level and create a macro to use later
*/}}
{{- define "certManager.certApi" -}}
{{- if eq ( .Capabilities.APIVersions.Has "cert-manager.io/v1" ) true -}}
{{- printf "cert-manager.io/v1" }}
{{- else if eq ( .Capabilities.APIVersions.Has "cert-manager.io/v1alpha3" ) true -}}
{{- printf "cert-manager.io/v1alpha3" }}
{{- else -}}
{{/*
If neither is true then we are likely running under "helm template", simulate the newer API level
*/}}
{{- printf "cert-manager.io/v1" }}
{{- end -}}
{{- end -}}
