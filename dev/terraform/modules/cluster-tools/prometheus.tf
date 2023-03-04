resource "argocd_application" "prometheus" {
  metadata {
    name      = "prometheus"
    namespace = var.argocd_namespace
    # finalizers = ["resources-finalizer.argocd.argoproj.io"]
  }
  spec {
    project = argocd_project.tools.metadata.0.name
    source {
      repo_url        = "https://prometheus-community.github.io/helm-charts"
      chart           = "prometheus"
      target_revision = "19.7.2"
      helm {
        release_name = "prometheus"
        values       = <<EOT
server:
  nodeSelector:
    node-pool: ${var.node_pool}
  persistentVolume:
    enabled: false
  ingress:
    enabled: true
    hosts:
      - prometheus.localhost
alertmanager:
  persistence:
    enabled: false
  nodeSelector:
    node-pool: ${var.node_pool}
kube-state-metrics:
  nodeSelector:
    node-pool: ${var.node_pool}
prometheus-pushgateway:
  nodeSelector:
    node-pool: ${var.node_pool}
global:
  evaluation_interval: 30s
  scrape_interval: 5s
scrape_configs:
- job_name: otel
  honor_labels: true
  static_configs:
  - targets:
    - 'otelcol:9464'
- job_name: otel-collector
  static_configs:
  - targets:
    - 'otelcol:8888'
EOT
      }
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = var.namespace
    }

    sync_policy {
      automated = {
        prune     = true
        self_heal = true
      }
    }
  }
}
