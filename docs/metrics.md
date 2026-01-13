# Cluster Network Addons Operator Metrics

| Name | Kind | Type | Description |
|------|------|------|-------------|
| kubevirt_cnao_cr_kubemacpool_deployed | Metric | Gauge | KubeMacpool is deployed by CNAO CR |
| kubevirt_cnao_cr_ready | Metric | Gauge | CNAO CR Ready |
| kubevirt_cnao_cr_kubemacpool_aggregated | Recording rule | Gauge | Total count of KubeMacPool manager pods deployed by CNAO CR |
| kubevirt_cnao_kubemacpool_duplicate_macs | Recording rule | Gauge | [DEPRECATED] Total count of duplicate KubeMacPool MAC addresses. This recording rule monitors VM MACs instead of running VMI MACs and will be removed in the next minor release. Use KubeMacPool's native VMI collision detection instead |
| kubevirt_cnao_kubemacpool_manager_up | Recording rule | Gauge | Total count of running KubeMacPool manager pods |
| kubevirt_cnao_operator_up | Recording rule | Gauge | Total count of running CNAO operators |

## Developing new metrics

All metrics documented here are auto-generated and reflect exactly what is being
exposed. After developing new metrics or changing old ones please regenerate
this document.
