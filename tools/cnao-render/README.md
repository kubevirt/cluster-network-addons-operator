# CNAO render POC

`cnao-render` uses CNAO's existing component render path to produce manifests without connecting to a Kubernetes cluster. Its input is a YAML or JSON file containing the target platform, namespace, CNAO spec fields, platform capabilities, and image overrides.

Build it with:

```sh
make cnao-render
```

Example input:

```yaml
platform: kubernetes
namespace: kube-system
spec:
  multus: {}
  imagePullPolicy: IfNotPresent
images:
  multus: quay.io/kubevirt/multus-cni:latest
```

Render each component into its own output folder:

```sh
build/_output/bin/cnao-render --input render.yaml --manifest-dir data --output-dir rendered
```

The command writes clean `manifests.yaml` files under each component directory. It removes `creationTimestamp: null` and CNAO's cleanup labels from Namespace objects, while preserving other component fields and labels. It does not add labels applied later by CNAO's controller. Each component's rendered objects are preserved; shared monitoring and network policy resources are included in every selected component's output. Add `--hco` to also write an `hco.go` file beside each manifest:

```sh
build/_output/bin/cnao-render --input render.yaml --manifest-dir data \
  --output-dir rendered --hco
```

Each generated Go file uses package `generated` and contains one `New<Component><Kind><Name>` constructor per object. It contains no HCO handler or operand wiring; that integration stays in HCO. Struct literals omit zero-valued fields and Kubernetes `TypeMeta`. Custom resource kinds without a supported struct mapping use an `unstructured.Unstructured` literal. The generated Go source and manifest output come from the same CNAO render invocation.

For OpenShift, set `platform: openshift` and provide cluster capability values under `cluster`. Components whose manifests are owned by OpenShift platform operators are omitted by CNAO's existing render logic. `openShiftNetwork` may be supplied when rendering depends on the OpenShift Network configuration.

This is a POC: consumers should treat the input and generated Go package as an initial contract pending confirmation of HCO's exact integration shape.
