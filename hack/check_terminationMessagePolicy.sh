#!/bin/bash

set -uo pipefail
shopt -s globstar nullglob

ROOT_DIR="data"

main() {
  echo "🔍 Scanning manifests under $ROOT_DIR..."
  scan_output="$(scan_manifests)"

  if [[ -z "$scan_output" ]]; then
      echo "✅ All manifests passed the terminationMessagePolicy checks."
    else
      echo "❌ There were terminationMessagePolicy issues in some manifests:"
      echo
      echo "$scan_output"
      exit 1
  fi
}

scan_manifests() {
  for file in "$ROOT_DIR"/**/*.yaml; do
    [[ -f "$file" ]] || continue
    process_file "$file"
  done
}

process_file() {
  local file="$1"
  local cleaned_yaml
  cleaned_yaml=$(preprocess_yaml "$file")
  [[ -z "$cleaned_yaml" ]] && return

  local docs
  if ! docs=$(echo "$cleaned_yaml" | yq -o=json eval 'select(.kind != null)' 2>/dev/null); then
    return
  fi

  local output_lines=()

  while IFS= read -r doc; do
    while IFS= read -r line; do
      [[ -n "$line" ]] && output_lines+=("$line")
    done < <(parse_and_validate_document "$doc")
  done < <(echo "$docs" | jq -c '.')

  if [[ "${#output_lines[@]}" -gt 0 ]]; then
    echo "📄 File: $file"
    printf '%s\n' "${output_lines[@]}"
    echo
  fi
}

preprocess_yaml() {
  local file="$1"
  sed 's/{{[^}]*}}//g' "$file"
}

parse_and_validate_document() {
  local doc="$1"
  local kind name container_path

  kind=$(echo "$doc" | jq -r '.kind // empty')
  name=$(echo "$doc" | jq -r '.metadata.name // "unnamed"')
  [[ -z "$kind" ]] && return

  case "$kind" in
    Pod) container_path=".spec.containers" ;;
    Deployment|DaemonSet) container_path=".spec.template.spec.containers" ;;
    *) return ;;
  esac

  echo "$doc" | jq -c "$container_path[]?" | while read -r container; do
    validate_container_policy "$kind" "$name" "$container"
  done
}

validate_container_policy() {
  local kind="$1"
  local resource_name="$2"
  local container="$3"

  local cname policy
  cname=$(echo "$container" | jq -r '.name // "unnamed"')
  policy=$(echo "$container" | jq -r '.terminationMessagePolicy // empty')

  if [[ -z "$policy" ]]; then
    echo "  ❌ [$kind/$resource_name][$cname] missing terminationMessagePolicy"
  elif [[ "$policy" != "FallbackToLogsOnError" ]]; then
    echo "  ⚠️  [$kind/$resource_name][$cname] has terminationMessagePolicy='$policy'"
  fi
}

main "$@"
