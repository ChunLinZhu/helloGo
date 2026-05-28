{{/* 通用标签 */}}
{{- define "hellogo.labels" -}}
app.kubernetes.io/name: {{ .name }}
app.kubernetes.io/instance: {{ .release }}
app.kubernetes.io/managed-by: Helm
{{- end }}

{{/* init 容器：等待 MySQL 和 Redis 就绪 */}}
{{- define "hellogo.initWaitDeps" -}}
initContainers:
  - name: wait-mysql
    image: busybox:1.36
    command: ['sh', '-c', 'until nc -z mysql {{ .Values.mysql.port }}; do echo waiting for mysql; sleep 2; done']
  - name: wait-redis
    image: busybox:1.36
    command: ['sh', '-c', 'until nc -z redis {{ .Values.redis.port }}; do echo waiting for redis; sleep 2; done']
{{- end }}

{{/* 环境变量注入（ConfigMap + Secret） */}}
{{- define "hellogo.env" -}}
- name: APP_ENV
  valueFrom:
    configMapKeyRef:
      name: hellogo-config
      key: APP_ENV
- name: DB_TYPE
  valueFrom:
    configMapKeyRef:
      name: hellogo-config
      key: DB_TYPE
- name: DB_HOST
  value: "mysql"
- name: DB_PORT
  value: "{{ .Values.mysql.port }}"
- name: DB_NAME
  valueFrom:
    configMapKeyRef:
      name: hellogo-config
      key: DB_NAME
- name: DB_USER
  valueFrom:
    configMapKeyRef:
      name: hellogo-config
      key: DB_USER
- name: DB_PASS
  valueFrom:
    secretKeyRef:
      name: hellogo-secrets
      key: db-password
- name: REDIS_HOST
  value: "redis"
- name: REDIS_PORT
  value: "{{ .Values.redis.port }}"
- name: JWT_SECRET
  valueFrom:
    secretKeyRef:
      name: hellogo-secrets
      key: jwt-secret
- name: JWT_EXPIRES
  valueFrom:
    configMapKeyRef:
      name: hellogo-config
      key: JWT_EXPIRES
- name: JWT_REFRESH_EXPIRES
  valueFrom:
    configMapKeyRef:
      name: hellogo-config
      key: JWT_REFRESH_EXPIRES
- name: LOGIN_MAX_FAILS
  valueFrom:
    configMapKeyRef:
      name: hellogo-config
      key: LOGIN_MAX_FAILS
- name: LOGIN_LOCK_TTL
  valueFrom:
    configMapKeyRef:
      name: hellogo-config
      key: LOGIN_LOCK_TTL
- name: USER_SERVICE_ADDR
  value: "user-service:50001"
- name: AUTH_SERVICE_ADDR
  value: "auth-service:50002"
- name: PERMISSION_SERVICE_ADDR
  value: "permission-service:50003"
- name: BIZ_SERVICE_ADDR
  value: "biz-service:50004"
{{- end }}
