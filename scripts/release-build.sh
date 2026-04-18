#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  ./scripts/release-build.sh
  ./scripts/release-build.sh --all

Options:
  --all    Build all release targets used in CI:
           linux/amd64, linux/arm64, darwin/arm64
EOF
}

mode="current"
if [[ $# -gt 1 ]]; then
  usage
  exit 2
fi
if [[ $# -eq 1 ]]; then
  case "$1" in
    --all)
      mode="all"
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      usage
      exit 2
      ;;
  esac
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"

cd "${repo_root}"

version="$(tr -d '[:space:]' < VERSION)"
if [[ -z "${version}" ]]; then
  echo "VERSION is empty" >&2
  exit 1
fi

ldflags="-s -w -X github.com/fserg/md-to-html/internal/version.Version=${version}"
dist_dir="${repo_root}/dist"
mkdir -p "${dist_dir}" "${repo_root}/web/static/dist"

echo "==> Generating templ code"
go run github.com/a-h/templ/cmd/templ@v0.3.1001 generate ./...

echo "==> Building Tailwind bundle"
npx tailwindcss -c tailwind.config.js -i web/static/src/app.css -o web/static/dist/app.css --minify

echo "==> Running tests"
go test ./...

targets=()
if [[ "${mode}" == "all" ]]; then
  targets+=("linux amd64")
  targets+=("linux arm64")
  targets+=("darwin arm64")
else
  current_goos="$(go env GOOS)"
  current_goarch="$(go env GOARCH)"
  targets+=("${current_goos} ${current_goarch}")
fi

artifacts=()
for target in "${targets[@]}"; do
  read -r goos goarch <<<"${target}"
  output="${dist_dir}/md-to-html-${goos}-${goarch}"
  echo "==> Building ${goos}/${goarch}"
  CGO_ENABLED=0 GOOS="${goos}" GOARCH="${goarch}" \
    go build -trimpath -ldflags="${ldflags}" -o "${output}" ./cmd/md-to-html
  artifacts+=("${output}")
done

checksum_file="${dist_dir}/SHA256SUMS"
(
  cd "${dist_dir}"
  shasum -a 256 "${artifacts[@]##${dist_dir}/}" > "${checksum_file}"
)

echo
echo "Artifacts:"
for artifact in "${artifacts[@]}"; do
  echo "  ${artifact}"
done
echo "  ${checksum_file}"

if [[ "${mode}" == "current" ]]; then
  echo
  echo "Run to verify:"
  echo "  ${artifacts[0]} serve"
fi
