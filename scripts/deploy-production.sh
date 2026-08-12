#!/usr/bin/env bash
set -Eeuo pipefail

readonly COMPOSE_FILE="/Project/compose.yml"
readonly DEPLOY_ENV="/Project/deployments/dujiao/deploy.env"
readonly REGISTRY="ghcr.io"
readonly IMAGE_REPOSITORY="ghcr.io/dovelora/dujiao-next"
readonly LEGACY_IMAGE_REPOSITORY="dovelora/dujiao-next"

if [[ "$#" -ne 3 ]]; then
  echo "Expected a GHCR image reference, digest, and registry username." >&2
  exit 2
fi

candidate_image="$1"
expected_digest="$2"
registry_username="$3"
if [[ ! "${candidate_image}" =~ ^ghcr\.io/dovelora/dujiao-next:gh-[0-9a-f]{12}-[0-9]+-[0-9]+$ ]]; then
  echo "Expected a uniquely tagged dovelora/dujiao-next GHCR image." >&2
  exit 2
fi
if [[ ! "${expected_digest}" =~ ^sha256:[0-9a-f]{64}$ ]]; then
  echo "Expected a sha256 image digest." >&2
  exit 2
fi
if [[ -z "${registry_username}" || "${registry_username}" =~ [[:space:]] ]]; then
  echo "Expected a non-empty registry username without whitespace." >&2
  exit 2
fi

docker_config_dir="$(mktemp -d /tmp/dovelora-docker-config.XXXXXX)"

cleanup() {
  rm -rf "${docker_config_dir}"
}
trap cleanup EXIT

read_image_pointer() {
  local key="$1"
  awk -F= -v key="${key}" '$1 == key { print substr($0, index($0, "=") + 1); exit }' "${DEPLOY_ENV}"
}

write_image_pointers() {
  local current_image="$1"
  local rollback_image="${2:-}"
  local pointer_dir
  local temporary_pointer
  pointer_dir="$(dirname "${DEPLOY_ENV}")"
  mkdir -p "${pointer_dir}"
  temporary_pointer="$(mktemp "${pointer_dir}/.deploy.env.XXXXXX")"
  {
    printf 'DUJIAO_IMAGE=%s\n' "${current_image}"
    if [[ -n "${rollback_image}" ]]; then
      printf 'DUJIAO_PREVIOUS_IMAGE=%s\n' "${rollback_image}"
    fi
  } > "${temporary_pointer}"
  chmod 640 "${temporary_pointer}"
  mv "${temporary_pointer}" "${DEPLOY_ENV}"
}

recreate_storefront() {
  docker compose \
    --env-file "${DEPLOY_ENV}" \
    -f "${COMPOSE_FILE}" \
    up -d --no-deps --force-recreate dujiao
}

wait_for_local_health() {
  local attempt
  for attempt in $(seq 1 30); do
    if curl --fail --silent --show-error --max-time 5 \
      http://127.0.0.1:18081/api/v1/public/config >/dev/null; then
      return 0
    fi
    sleep 2
  done
  return 1
}

wait_for_public_health() {
  local attempt
  for attempt in $(seq 1 8); do
    if curl --fail --silent --show-error --max-time 10 \
      https://www.dovelora.com/ >/dev/null; then
      return 0
    fi
    sleep 3
  done
  return 1
}

rollback() {
  local previous_image="$1"
  local previous_rollback_image="${2:-}"
  echo "Deployment failed; restoring ${previous_image}." >&2
  write_image_pointers "${previous_image}" "${previous_rollback_image}"
  recreate_storefront
  wait_for_local_health
}

cleanup_failed_candidate() {
  local candidate_image="$1"

  if ! docker image rm "${candidate_image}" >/dev/null 2>&1; then
    echo "Warning: failed to remove unsuccessful candidate image ${candidate_image}." >&2
  fi
}

cleanup_old_application_images() {
  local current_image="$1"
  local rollback_image="$2"
  local image
  local repository

  for repository in "${IMAGE_REPOSITORY}" "${LEGACY_IMAGE_REPOSITORY}"; do
    while IFS= read -r image; do
      if [[ -z "${image}" || "${image}" == "${current_image}" || "${image}" == "${rollback_image}" ]]; then
        continue
      fi

      if docker image rm "${image}" >/dev/null; then
        echo "Removed superseded application image ${image}."
      else
        echo "Warning: failed to remove superseded application image ${image}." >&2
      fi
    done < <(docker image ls "${repository}" --format '{{.Repository}}:{{.Tag}}' | sort -u)
  done
}

if [[ ! -f "${DEPLOY_ENV}" ]]; then
  echo "Missing deployment pointer: ${DEPLOY_ENV}" >&2
  exit 1
fi

previous_image="$(read_image_pointer DUJIAO_IMAGE)"
if [[ -z "${previous_image}" ]]; then
  echo "Deployment pointer does not contain DUJIAO_IMAGE." >&2
  exit 1
fi
previous_rollback_image="$(read_image_pointer DUJIAO_PREVIOUS_IMAGE)"

if [[ "${candidate_image}" == "${previous_image}" ]]; then
  rollback_image="${previous_rollback_image}"
else
  rollback_image="${previous_image}"
fi

payment_before="$(docker inspect -f '{{.Id}}|{{.State.StartedAt}}' epusdt)"

registry_token="$(cat)"
if [[ -z "${registry_token}" ]]; then
  echo "Expected a GHCR token on standard input." >&2
  exit 2
fi
printf '%s' "${registry_token}" \
  | docker --config "${docker_config_dir}" login "${REGISTRY}" \
      --username "${registry_username}" --password-stdin >/dev/null
unset registry_token
pinned_image="${candidate_image}@${expected_digest}"
docker --config "${docker_config_dir}" pull "${pinned_image}"
pulled_image_id="$(docker image inspect -f '{{.Id}}' "${pinned_image}")"
docker image tag "${pulled_image_id}" "${candidate_image}"

write_image_pointers "${candidate_image}" "${rollback_image}"
if ! recreate_storefront || ! wait_for_local_health || ! wait_for_public_health; then
  rollback "${previous_image}" "${previous_rollback_image}"
  if [[ "${candidate_image}" != "${previous_image}" ]]; then
    cleanup_failed_candidate "${candidate_image}"
  fi
  exit 1
fi

payment_after="$(docker inspect -f '{{.Id}}|{{.State.StartedAt}}' epusdt)"
if [[ "${payment_before}" != "${payment_after}" ]]; then
  echo "Payment container identity changed outside the deployment action." >&2
  rollback "${previous_image}" "${previous_rollback_image}"
  if [[ "${candidate_image}" != "${previous_image}" ]]; then
    cleanup_failed_candidate "${candidate_image}"
  fi
  exit 1
fi

deployed_image="$(docker inspect -f '{{.Config.Image}}' dujiao)"
if [[ "${deployed_image}" != "${candidate_image}" ]]; then
  echo "Storefront is running ${deployed_image}, expected ${candidate_image}." >&2
  rollback "${previous_image}" "${previous_rollback_image}"
  if [[ "${candidate_image}" != "${previous_image}" ]]; then
    cleanup_failed_candidate "${candidate_image}"
  fi
  exit 1
fi

cleanup_old_application_images "${candidate_image}" "${rollback_image}"

echo "Deployed ${candidate_image}."
if [[ -n "${rollback_image}" ]]; then
  echo "Retained ${rollback_image} for rollback."
fi
echo "Payment container remained ${payment_after}."
