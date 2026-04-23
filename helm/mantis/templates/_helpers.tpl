{{/*
Common labels
*/}}
{{- define "mantis.labels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" }}
{{- end -}}

{{/*
Namespace
*/}}
{{- define "mantis.namespace" -}}
{{- .Values.global.namespace | default "mantis" -}}
{{- end -}}

{{/*
Database URL built from secret entries
*/}}
{{- define "mantis.databaseUrl" -}}
postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@postgres:{{ .Values.postgres.port }}/$(POSTGRES_DB)?sslmode=disable
{{- end -}}

{{/*
imagePullSecrets block
*/}}
{{- define "mantis.imagePullSecrets" -}}
{{- with .Values.global.imagePullSecrets }}
imagePullSecrets:
{{- toYaml . | nindent 2 }}
{{- end }}
{{- end -}}
