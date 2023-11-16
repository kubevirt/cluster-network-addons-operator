# Cluster Network Addons Operator Metrics

### kubevirt_cnao_cr_kubemacpool_aggregated
Total count of KubeMacPool manager pods deployed by CNAO CR. Type: Gauge.

### kubevirt_cnao_cr_kubemacpool_deployed
KubeMacpool is deployed by CNAO CR. Type: Gauge.

### kubevirt_cnao_cr_ready
CNAO CR Ready. Type: Gauge.

### kubevirt_cnao_kubemacpool_duplicate_macs
Total count of duplicate KubeMacPool MAC addresses. Type: Gauge.

### kubevirt_cnao_kubemacpool_manager_up
Total count of running KubeMacPool manager pods. Type: Gauge.

### kubevirt_cnao_operator_up
Total count of running CNAO operators. Type: Gauge.

## Developing new metrics

All metrics documented here are auto-generated and reflect exactly what is being
exposed. After developing new metrics or changing old ones please regenerate
this document.
