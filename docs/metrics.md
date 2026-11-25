# Cluster Network Addons Operator Metrics

| Name | Kind | Type | Description |
|------|------|------|-------------|
| kubevirt_cnao_cr_kubemacpool_deployed | Metric | Gauge | [Deprecated] KubeMacpool is deployed by CNAO CR |
| kubevirt_cnao_cr_ready | Metric | Gauge | CNAO CR Ready |
| cluster:kubevirt_cnao_cr_kubemacpool_deployed:sum | Recording rule | Gauge | [Deprecated] The number of KubeMacPool manager pods deployed by CNAO CR |
| cluster:kubevirt_cnao_kubemacpool_manager_up:sum | Recording rule | Gauge | [Deprecated] The number of KubeMacPool manager pods that are up |
| cluster:kubevirt_cnao_operator_up:sum | Recording rule | Gauge | The number of CNAO pods that are up |
| kubevirt_cnao_operator_up | Recording rule | Gauge | [Deprecated] Total count of running CNAO operators |

## Developing new metrics

All metrics documented here are auto-generated and reflect exactly what is being
exposed. After developing new metrics or changing old ones please regenerate
this document.
