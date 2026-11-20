# CNAO render POC

## Purpose

CNAO remains the source of truth for network component manifests, defaults, and image versions. The `cnao-render` CLI lets kubevirtci and HCO consume those components without installing CNAO. Consumers use the renderer from the CNAO release they pin.

## Use

The CLI reads a YAML or JSON file that selects a platform (`kubernetes` or `openshift`), namespace, enabled components, and any cluster capabilities needed for rendering. Image overrides are optional; CNAO defaults apply when they are omitted. `render.yaml` at the repository root is an example for Multus and KubeMacPool.

```sh
make cnao-render
build/_output/bin/cnao-render --input render.yaml --manifest-dir data \
  --output-dir rendered --hco
```

The CLI calls CNAO's existing component renderers. It does not connect to a cluster or apply resources.

## Output

The CLI writes one directory per enabled component. Each directory contains `manifests.yaml`; `--hco` also writes `hco.go` with Go constructors for that component's rendered objects.

```text
rendered/
  multus/
    manifests.yaml
    hco.go
  kubeMacPool/
    manifests.yaml
    hco.go
```

Each component directory contains the objects returned by that component's CNAO renderer. Shared monitoring resources and network policies are included in every selected component directory. No Namespace is synthesized when a component renderer does not emit one.

The CLI removes `creationTimestamp: null` and the labels CNAO cleanup removes from Namespace objects. It preserves other component labels and fields, and does not add labels applied later by CNAO's controller. The HCO output omits zero-valued fields and `TypeMeta`; it generates object constructors only. HCO handler and operand wiring stays in HCO. Resource kinds without a supported Go type use `unstructured.Unstructured`.

## Limits

This is a POC. The input is a CNAO spec subset, not a `NetworkAddonsConfig` resource. Platform capabilities must be supplied in the input. The generated HCO constructors are object definitions; HCO still needs one-time integration code to call and manage them.
