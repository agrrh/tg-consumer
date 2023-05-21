# tg-consumer

Universal Telegram consumer daemon.

- Consumes events from Telegram API
- Sends JSON to NATS JetStream topic

## Sample ArgoCD application

```yaml
---

kind: Application
apiVersion: argoproj.io/v1alpha1
metadata:
  name: tg-my-consumer
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/agrrh/tg-consumer
    targetRevision: master
    path: helm/
    helm:
      parameters:
        - name: app.name
          value: my-bot               # TODO: Change
        - name: app.nats.addr
          value: nats.namespace:4222  # TODO: Change
        - name: app.nats.prefix
          value: my-bot               # TODO: Change
  destination:
    namespace: tg-my-consumer
    server: https://kubernetes.default.svc
  syncPolicy:
    automated:
      selfHeal: true
      prune: true
```
