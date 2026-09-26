#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MAKEFILE="${ROOT_DIR}/Makefile"

teardown_block="$(sed -n '/^teardown:/,/^logs:/p' "$MAKEFILE")"
clean_local_block="$(sed -n '/^clean-local:/,/^clean-bootstrap-state:/p' "$MAKEFILE")"
uninstall_block="$(sed -n '/^uninstall-cli:/,/^test:/p' "$MAKEFILE")"

[[ "$teardown_block" == *'$(MAKE) bootstrap-destroy'* ]]
[[ "$teardown_block" == *'$(CLI_INSTALL) logout'* ]]
[[ "$teardown_block" == *'go run ./cmd/spx logout'* ]]
[[ "$teardown_block" == *'$(MAKE) clean-local'* ]]
[[ "$teardown_block" == *'$(MAKE) clean-bootstrap-state'* ]]

destroy_position="$(grep -nF '$(MAKE) bootstrap-destroy' <<<"$teardown_block" | head -n1 | cut -d: -f1)"
logout_position="$(grep -nF '$(CLI_INSTALL) logout' <<<"$teardown_block" | head -n1 | cut -d: -f1)"
clean_position="$(grep -nF '$(MAKE) clean-local' <<<"$teardown_block" | head -n1 | cut -d: -f1)"
state_position="$(grep -nF '$(MAKE) clean-bootstrap-state' <<<"$teardown_block" | head -n1 | cut -d: -f1)"
(( destroy_position < logout_position ))
(( logout_position < clean_position ))
(( clean_position < state_position ))

[[ "$uninstall_block" == *'$(CLI_INSTALL)'* ]]
[[ "$uninstall_block" != *'$(CLI_CONFIG_FILE)'* ]]

printf 'cleanup checks passed.\n'
