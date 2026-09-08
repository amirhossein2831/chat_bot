# Deployments

Docker Compose on 2–3 VMs for Phase 1. Kubernetes is deliberately deferred:
with one gateway, one worker and one GPU box there is no scheduling problem to
solve, and the Compose files translate to manifests when a second gateway node
earns it.

Services: postgres:17-alpine · redis:7-alpine · minio · gateway · caddy

Two settings that are easy to miss and break streaming:

- Caddy's reverse proxy needs `flush_interval -1`, or SSE tokens are buffered
  and every answer arrives in one burst at the end.
- MinIO needs a dedicated volume plus a scheduled `mc mirror` for backup, and a
  lifecycle rule expiring raw uploads after 90 days. Extracted text is small
  and worth keeping; original PDFs are not.

The relay is deployed separately, outside Iran, on at least two hosts with
different provider accounts. It is not part of this stack.
