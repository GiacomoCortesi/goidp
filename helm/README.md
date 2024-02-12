{
    "title": "Chart Values",
    "type": "object",
    "properties": {
        "global": {
            "type": "object",
            "properties": {
                "imageRegistry": {
                    "type": "string",
                    "description": "Global Docker image registry",
                    "default": ""
                },
                "imagePullSecrets": {
                    "type": "array",
                    "description": "Global Docker registry secret names as an array",
                    "default": [],
                    "items": {}
                }
            }
        },
        "replicaCount": {
            "type": "number",
            "description": "Number of sftp replicas to deploy",
            "default": 1
        },
        "image": {
            "type": "object",
            "properties": {
                "registry": {
                    "type": "string",
                    "description": "goidp image registry",
                    "default": "docker.io"
                },
                "repository": {
                    "type": "string",
                    "description": "goidp image repository",
                    "default": "giacomocortesi/goidp"
                },
                "tag": {
                    "type": "string",
                    "description": "goidp image tag (immutable tags are recommended)",
                    "default": "1.0.0"
                },
                "pullPolicy": {
                    "type": "string",
                    "description": "goidp image pull policy",
                    "default": "IfNotPresent"
                },
                "pullSecrets": {
                    "type": "array",
                    "description": "goidp image pull secrets",
                    "default": [],
                    "items": {}
                }
            }
        },
        "initContainer": {
            "type": "object",
            "properties": {
                "image": {
                    "type": "object",
                    "properties": {
                        "registry": {
                            "type": "string",
                            "description": "container registry",
                            "default": "docker.io"
                        },
                        "repository": {
                            "type": "string",
                            "description": "container repository",
                            "default": "alpine"
                        },
                        "pullPolicy": {
                            "type": "string",
                            "description": "container pullPolicy",
                            "default": "IfNotPresent"
                        },
                        "tag": {
                            "type": "string",
                            "description": "container tag",
                            "default": "3.16.1"
                        }
                    }
                }
            }
        },
        "nameOverride": {
            "type": "string",
            "description": "String to partially override fullname template (will maintain the release name)",
            "default": ""
        },
        "fullnameOverride": {
            "type": "string",
            "description": "String to fully override fullname template",
            "default": ""
        },
        "extraEnvVars": {
            "type": "array",
            "description": "Array with extra environment variables to add to the pod",
            "default": [],
            "items": {}
        },
        "serviceAccount": {
            "type": "object",
            "properties": {
                "create": {
                    "type": "boolean",
                    "description": "Specifies whether a service account should be created",
                    "default": false
                },
                "annotations": {
                    "type": "object",
                    "description": "Annotations to add to the service account.",
                    "default": {}
                },
                "name": {
                    "type": "string",
                    "description": "The name of the service account to use. If not set and create is true, a name is generated using the fullname template",
                    "default": ""
                }
            }
        },
        "jwt": {
            "type": "object",
            "properties": {
                "jwtAccessExpireTime": {
                    "type": "string",
                    "description": "JWT expiration time.",
                    "default": "5m"
                },
                "jwtRefreshExpireTime": {
                    "type": "string",
                    "description": "JWT refresh espiration time.",
                    "default": "2400m"
                },
                "jwtRefresh": {
                    "type": "boolean",
                    "description": "Enable token refresh.",
                    "default": true
                },
                "jwtUseKey": {
                    "type": "boolean",
                    "description": "Enable JWT.",
                    "default": true
                },
                "trustedKeysMountPath": {
                    "type": "string",
                    "description": "File system location where trusted keys may be found",
                    "default": "/run/pubkeys"
                },
                "trustedKeys": {
                    "type": "object",
                    "description": "Array of trusted public keys.",
                    "default": {}
                },
                "secretName": {
                    "type": "string",
                    "description": "Name of the secret tha contains private and public key.",
                    "default": "goidp-rsa"
                },
                "publicKeySecretKey": {
                    "type": "string",
                    "description": "name of key within the jwt.secretName holding the public key",
                    "default": "tls.crt"
                },
                "privateKeySecretKey": {
                    "type": "string",
                    "description": "name of key within the jwt.secretName holding the private key",
                    "default": "tls.key"
                },
                "certManager": {
                    "type": "object",
                    "properties": {
                        "enabled": {
                            "type": "boolean",
                            "description": "If true, use the Kubernetes Cert-Manager for certificate creation",
                            "default": false
                        },
                        "issuerName": {
                            "type": "string",
                            "description": "If using Cert-Manager, what issuer should be used",
                            "default": "issuer-selfsigned-issuer"
                        },
                        "certName": {
                            "type": "string",
                            "description": "If using Cert-Manager, what certificate name should be used",
                            "default": "goidp-certificate"
                        }
                    }
                }
            }
        },
        "database": {
            "type": "object",
            "properties": {
                "dbName": {
                    "type": "string",
                    "description": "Database name to use/create.",
                    "default": "oranmgr"
                },
                "dbHost": {
                    "type": "string",
                    "description": "Host/Service where the database is listening on.",
                    "default": "goidp-postgres"
                },
                "dbPort": {
                    "type": "number",
                    "description": "Port where the database is listening on.",
                    "default": 5432
                },
                "dbSSLMode": {
                    "type": "string",
                    "description": "Database SSL mode",
                    "default": "disable"
                },
                "dbTimezone": {
                    "type": "string",
                    "description": "Database time standard.",
                    "default": "UTC"
                },
                "dbCleanupPeriod": {
                    "type": "string",
                    "description": "Database cleanup period specified in cron annotation.",
                    "default": "0 2 * * *"
                },
                "dbMaxEventsNumber": {
                    "type": "number",
                    "description": "Max events threshold on DB (chron-job is activated when events num. > th)",
                    "default": 100
                },
                "dbUser": {
                    "type": "string",
                    "description": "Database user name.",
                    "default": "oranmgr"
                },
                "dbPass": {
                    "type": "string",
                    "description": "Database user password. Auto-generated if not specified.",
                    "default": ""
                },
                "dbSecretOverride": {
                    "type": "string",
                    "description": "Forced name of the db secret.",
                    "default": "idp-password"
                },
                "dbSecretMountPath": {
                    "type": "string",
                    "description": "Path where to mount the password of the database as secret.",
                    "default": "/run/secrets/dbsecret"
                },
                "createSecret": {
                    "type": "boolean",
                    "description": "If enabled create db secret.",
                    "default": true
                },
                "dbCharset": {
                    "type": "string",
                    "description": "Database charset.",
                    "default": "utf8"
                },
                "dbParseTime": {
                    "type": "boolean",
                    "description": "If enabled db scan DATE and DATETIME automatically",
                    "default": true
                },
                "dbShowSql": {
                    "type": "boolean",
                    "description": "If enabled, start the dbms in debug mode. Show every query in the logs.",
                    "default": false
                }
            }
        },
        "app": {
            "type": "object",
            "properties": {
                "host": {
                    "type": "string",
                    "description": "Identity provider service name or external endpoint.",
                    "default": "0.0.0.0"
                },
                "port": {
                    "type": "number",
                    "description": "Identity provider service port.",
                    "default": 8889
                },
                "writeTimeout": {
                    "type": "number",
                    "description": "Write operations timeout in seconds.",
                    "default": 15
                },
                "readTimeout": {
                    "type": "number",
                    "description": "Read operations timeout in seconds.",
                    "default": 15
                },
                "idleTimeout": {
                    "type": "number",
                    "description": "Idle timeout in seconds.",
                    "default": 60
                },
                "logLevel": {
                    "type": "string",
                    "description": "Log level as a string [error,warning,info,debug]",
                    "default": "debug"
                }
            }
        },
        "service": {
            "type": "object",
            "properties": {
                "type": {
                    "type": "string",
                    "description": "Kubernetes service type for api service.",
                    "default": "ClusterIP"
                },
                "port": {
                    "type": "number",
                    "description": "ggoidp HTTP port.",
                    "default": 8889
                }
            }
        },
        "resources": {
            "type": "object",
            "description": "Resource values to add to deployment",
            "default": {}
        },
        "podAnnotations": {
            "type": "object",
            "description": "A collection of custom annotations desired for the pods in the deployment",
            "default": {}
        },
        "podSecurityContext": {
            "type": "object",
            "description": "Securitycontext of the pod template in the app deployment",
            "default": {}
        },
        "securityContext": {
            "type": "object",
            "description": "Securitycontext of the container template in the app deployment",
            "default": {}
        },
        "autoscaling": {
            "type": "object",
            "properties": {
                "enabled": {
                    "type": "boolean",
                    "description": "if true autoscaling is enabled",
                    "default": false
                },
                "minReplicas": {
                    "type": "number",
                    "description": "minimum number of pods for idp component",
                    "default": 1
                },
                "maxReplicas": {
                    "type": "number",
                    "description": "maximum number of pods for idp component",
                    "default": 100
                },
                "targetCPUUtilizationPercentage": {
                    "type": "number",
                    "description": "percentage of CPU usage",
                    "default": 80
                },
                "targetMemoryUtilizationPercentage": {
                    "type": "number",
                    "description": "percentage of memory usage",
                    "default": 80
                }
            }
        },
        "nodeSelector": {
            "type": "object",
            "description": "Provide custom pod nodeSelector values",
            "default": {}
        },
        "tolerations": {
            "type": "array",
            "description": "Provide custom pod toleration value(s)",
            "default": [],
            "items": {}
        },
        "affinity": {
            "type": "object",
            "description": "Provide custom pod affinity value(s)",
            "default": {}
        },
        "livenessProbe": {
            "type": "object",
            "properties": {
                "httpPath": {
                    "type": "string",
                    "description": "Target path for liveness probe.",
                    "default": "/versions"
                },
                "initialDelaySeconds": {
                    "type": "number",
                    "description": "Delay before start the liveness probe check.",
                    "default": 15
                },
                "periodSeconds": {
                    "type": "number",
                    "description": "Number of seconds between liveness probe check.",
                    "default": 30
                },
                "enabled": {
                    "type": "boolean",
                    "description": "Enable livenessProbe.",
                    "default": true
                },
                "httpHeaders": {
                    "type": "array",
                    "description": "",
                    "items": {
                        "type": "object",
                        "properties": {
                            "name": {
                                "type": "string",
                                "description": ""
                            },
                            "value": {
                                "type": "string",
                                "description": ""
                            }
                        }
                    }
                }
            }
        },
        "readinessProbe": {
            "type": "object",
            "properties": {
                "httpPath": {
                    "type": "string",
                    "description": "Target path for readiness probe.",
                    "default": "/versions"
                },
                "initialDelaySeconds": {
                    "type": "number",
                    "description": "Delay before start the readiness probe check.",
                    "default": 5
                },
                "periodSeconds": {
                    "type": "number",
                    "description": "Number of seconds between readiness probe check.",
                    "default": 5
                },
                "enabled": {
                    "type": "boolean",
                    "description": "Enable readinessProbe.",
                    "default": true
                },
                "httpHeaders": {
                    "type": "array",
                    "description": "",
                    "items": {
                        "type": "object",
                        "properties": {
                            "name": {
                                "type": "string",
                                "description": ""
                            },
                            "value": {
                                "type": "string",
                                "description": ""
                            }
                        }
                    }
                }
            }
        },
        "postgresql": {
            "type": "object",
            "properties": {
                "image": {
                    "type": "object",
                    "properties": {
                        "registry": {
                            "type": "string",
                            "description": "Host providing the Docker registry from which to pull Postgres SQL",
                            "default": "docker.io"
                        },
                        "repository": {
                            "type": "string",
                            "description": "Repository on host from which to pull Postgres SQL",
                            "default": "bitnami/postgresql"
                        },
                        "tag": {
                            "type": "string",
                            "description": "Image tag of Postgres SQL version to pull",
                            "default": "14.2.0-debian-10-r88"
                        }
                    }
                },
                "enabled": {
                    "type": "boolean",
                    "description": "Enable postgresql.",
                    "default": true
                },
                "fullnameOverride": {
                    "type": "string",
                    "description": "Host/Service where the database is listening on.",
                    "default": "goidp-postgres"
                },
                "timezone": {
                    "type": "string",
                    "description": "Database time standard.",
                    "default": "UTC"
                },
                "tls": {
                    "type": "object",
                    "properties": {
                        "enabled": {
                            "type": "boolean",
                            "description": "Enable TLS.",
                            "default": false
                        }
                    }
                },
                "ssl_mode": {
                    "type": "string",
                    "description": "Database SSL mode.",
                    "default": "disable"
                },
                "auth": {
                    "type": "object",
                    "properties": {
                        "username": {
                            "type": "string",
                            "description": "Database user name.",
                            "default": "oranmgr"
                        },
                        "existingSecret": {
                            "type": "string",
                            "description": "Forced name of the db secret.",
                            "default": "idp-password"
                        },
                        "database": {
                            "type": "string",
                            "description": "Database name to use/create.",
                            "default": "oranmgr"
                        }
                    }
                },
                "primary": {
                    "type": "object",
                    "properties": {
                        "persistence": {
                            "type": "object",
                            "properties": {
                                "enabled": {
                                    "type": "boolean",
                                    "description": "Enable DB persistence.",
                                    "default": false
                                }
                            }
                        }
                    }
                }
            }
        },
        "helmTest": {
            "type": "object",
            "properties": {
                "image": {
                    "type": "object",
                    "properties": {
                        "registry": {
                            "type": "string",
                            "description": "",
                            "default": "docker.io"
                        },
                        "repository": {
                            "type": "string",
                            "description": "",
                            "default": "busybox"
                        },
                        "tag": {
                            "type": "string",
                            "description": "",
                            "default": "1.34.1"
                        }
                    }
                }
            }
        }
    }
}