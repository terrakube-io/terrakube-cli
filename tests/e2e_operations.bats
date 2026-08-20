#!/usr/bin/env bats

# ==============================================================================
# Terrakube CLI BATS Test Suite — End to End Operations
# ==============================================================================
# Prerequisites:
#   - TERRAKUBE_API_URL environment variable set (e.g. http://localhost:8080)
#   - TERRAKUBE_PAT environment variable set (Bearer token)
#   - TERRAKUBE_BIN optional (defaults to 'terrakube' or './terrakube')
# ==============================================================================

STATE_FILE="${BATS_FILE_TMPDIR:-/tmp}/terrakube_e2e_state.env"

setup_file() {
    TEST_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")" && pwd)"
    CLI_DIR="$(cd "$TEST_DIR/.." && pwd)"

    # 0. Build terrakube CLI binary
    run go build -o "$CLI_DIR/terrakube" "$CLI_DIR/main.go"
    if [ "$status" -ne 0 ]; then
        echo "Failed to build terrakube binary in $CLI_DIR: $output" >&2
        return 1
    fi

    TERRAKUBE_CMD="$CLI_DIR/terrakube"

    # 1. Login verification
    if [ -z "$TERRAKUBE_API_URL" ] || [ -z "$TERRAKUBE_PAT" ]; then
        echo "Error: TERRAKUBE_API_URL and TERRAKUBE_PAT environment variables must be set." >&2
        return 1
    fi

    run "$TERRAKUBE_CMD" login -a "$TERRAKUBE_API_URL" -t "$TERRAKUBE_PAT"
    if [ "$status" -ne 0 ]; then
        echo "Login failed: $output" >&2
        return 1
    fi
}

setup() {
    TEST_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")" && pwd)"
    CLI_DIR="$(cd "$TEST_DIR/.." && pwd)"

    TERRAKUBE_CMD="${TERRAKUBE_BIN:-$CLI_DIR/terrakube}"
    if [ ! -x "$TERRAKUBE_CMD" ] && [ -x "./terrakube" ]; then
        TERRAKUBE_CMD="./terrakube"
    fi

    STATE_FILE="${BATS_FILE_TMPDIR:-/tmp}/terrakube_e2e_state.env"
    if [ -f "$STATE_FILE" ]; then
        # shellcheck disable=SC1090
        source "$STATE_FILE"
    fi
}

assert_json_array() {
    [ "$status" -eq 0 ]
    echo "$output" | jq -e 'type == "array"' >/dev/null
}

assert_success() {
    if [ "$status" -ne 0 ]; then
        echo "Command failed with status $status. Output: $output" >&2
        return 1
    fi
}

# ==============================================================================
# Step 1: Organization Creation & Update Verification
# ==============================================================================

@test "1. Ensure organization 'bats' exists and update description" {
    # Check if 'bats' organization exists
    run "$TERRAKUBE_CMD" organization list --filter "name==bats" --output json
    assert_success

    ORG_ID=""
    if [ "$status" -eq 0 ]; then
        ORG_ID=$(echo "$output" | jq -r '.[0].id // empty' 2>/dev/null || true)
    fi

    # Create organization if it does not exist
    if [ -z "$ORG_ID" ] || [ "$ORG_ID" = "null" ]; then
        run "$TERRAKUBE_CMD" organization create --name "bats" --description "Initial BATS Org" --execution-mode "remote" --output json
        assert_success
        ORG_ID=$(echo "$output" | jq -r '.id // empty' 2>/dev/null || true)
    fi

    [ -n "$ORG_ID" ] && [ "$ORG_ID" != "null" ]

    # Update organization description and execution mode
    UPDATED_DESC="Updated BATS Org Description"
    run "$TERRAKUBE_CMD" organization update --id "$ORG_ID" --name "bats" --description "$UPDATED_DESC" --execution-mode "remote" --output json
    assert_success

    # Validate JSON output format
    run "$TERRAKUBE_CMD" organization get --id "$ORG_ID" --output json
    assert_success
    GET_DESC=$(echo "$output" | jq -r '.attributes.description // .description // empty' 2>/dev/null || true)
    GET_MODE=$(echo "$output" | jq -r '.attributes.executionMode // .executionMode // empty' 2>/dev/null || true)
    if [ "$GET_DESC" != "$UPDATED_DESC" ]; then
        echo "Mismatch: GET_DESC='$GET_DESC', expected='$UPDATED_DESC', raw output: $output" >&2
        return 1
    fi
    [ "$GET_MODE" = "remote" ]

    # Validate Table output format
    run "$TERRAKUBE_CMD" organization get --id "$ORG_ID" --output table
    assert_success

    echo "export TERRAKUBE_TEST_E2E_ORG_ID=\"$ORG_ID\"" > "$STATE_FILE"
}

# ==============================================================================
# Step 2: Team Creation & Role Permission Verification
# ==============================================================================

@test "2. Create TERRAKUBE_ADMIN team and verify permissions" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"

    run "$TERRAKUBE_CMD" team create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --name "TERRAKUBE_ADMIN" \
        --role "Admin" \
        --manage-workspace \
        --manage-module \
        --manage-provider \
        --manage-state \
        --manage-collection \
        --manage-vcs \
        --manage-template \
        --manage-job \
        --plan-job \
        --approve-job \
        --output json
    assert_success

    TEAM_ID=$(echo "$output" | jq -r '.id // empty' 2>/dev/null || true)
    [ -n "$TEAM_ID" ] && [ "$TEAM_ID" != "null" ]

    # Read team values and validate (JSON format)
    run "$TERRAKUBE_CMD" team get -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$TEAM_ID" --output json
    assert_success
    TEAM_NAME=$(echo "$output" | jq -r '.attributes.name // .name // empty' 2>/dev/null || true)
    TEAM_ROLE=$(echo "$output" | jq -r '.attributes.role // .role // empty' 2>/dev/null || true)
    [ "$TEAM_NAME" = "TERRAKUBE_ADMIN" ]
    [ "$TEAM_ROLE" = "Admin" ]

    # Read team values and validate (Table format)
    run "$TERRAKUBE_CMD" team get -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$TEAM_ID" --output table
    assert_success

    run "$TERRAKUBE_CMD" team list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --output table
    assert_success

    echo "export TERRAKUBE_TEST_E2E_TEAM_ID=\"$TEAM_ID\"" >> "$STATE_FILE"
}

# ==============================================================================
# Step 3: Workspace Creation & Description Update
# ==============================================================================

@test "3. Create workspace with OpenTofu and update description" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"

    RAND_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    WS_NAME="test${RAND_SUFFIX}"

    run "$TERRAKUBE_CMD" workspace create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --name "$WS_NAME" \
        --source "https://github.com/terrakube-io/terrakube-docker-compose" \
        --branch "main" \
        --folder "/" \
        --iac-type "tofu" \
        --iac-version "1.12.5" \
        --execution-mode "remote" \
        --output json
    assert_success

    WS_ID=$(echo "$output" | jq -r '.id // empty' 2>/dev/null || true)
    [ -n "$WS_ID" ] && [ "$WS_ID" != "null" ]

    # Update description to a 6-character random alphanumeric string
    RAND_DESC=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 6)
    run "$TERRAKUBE_CMD" workspace update -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$WS_ID" \
        --name "$WS_NAME" \
        --description "$RAND_DESC" \
        --source "https://github.com/terrakube-io/terrakube-docker-compose" \
        --branch "main" \
        --folder "/" \
        --iac-type "tofu" \
        --iac-version "1.12.5" \
        --execution-mode "remote" \
        --output json
    assert_success

    # List workspace and validate JSON format
    run "$TERRAKUBE_CMD" workspace list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --filter "name==$WS_NAME" --output json
    assert_json_array
    FETCHED_DESC=$(echo "$output" | jq -r '.[0].attributes.description // .[0].description // empty' 2>/dev/null || true)
    [ "$FETCHED_DESC" = "$RAND_DESC" ]

    # List workspace and validate Table format
    run "$TERRAKUBE_CMD" workspace list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --output table
    assert_success

    echo "export TERRAKUBE_TEST_E2E_WS_ID=\"$WS_ID\"" >> "$STATE_FILE"
    echo "export TERRAKUBE_TEST_E2E_WS_NAME=\"$WS_NAME\"" >> "$STATE_FILE"
}

# ==============================================================================
# Step 4: Environment Variables CRUD & Validation
# ==============================================================================

@test "4. Add, update, and validate environment variables" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"
    [ -n "$TERRAKUBE_TEST_E2E_WS_ID" ] || skip "Workspace ID not available"

    # Add 3 dummy environment variables
    run "$TERRAKUBE_CMD" variable create -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --key "test1" --value "1" --category "ENV" --output json
    assert_success
    VAR1_ID=$(echo "$output" | jq -r '.id // empty' 2>/dev/null || true)

    run "$TERRAKUBE_CMD" variable create -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --key "test2" --value "2" --category "ENV" --output json
    assert_success
    VAR2_ID=$(echo "$output" | jq -r '.id // empty' 2>/dev/null || true)

    run "$TERRAKUBE_CMD" variable create -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --key "test3" --value "3" --category "ENV" --output json
    assert_success
    VAR3_ID=$(echo "$output" | jq -r '.id // empty' 2>/dev/null || true)

    [ -n "$VAR1_ID" ] && [ -n "$VAR2_ID" ] && [ -n "$VAR3_ID" ]

    # Update values to test1=a, test2=b, test3=c
    run "$TERRAKUBE_CMD" variable update -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --id "$VAR1_ID" --key "test1" --value "a" --category "ENV" --output json
    assert_success

    run "$TERRAKUBE_CMD" variable update -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --id "$VAR2_ID" --key "test2" --value "b" --category "ENV" --output json
    assert_success

    run "$TERRAKUBE_CMD" variable update -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --id "$VAR3_ID" --key "test3" --value "c" --category "ENV" --output json
    assert_success

    # Fetch new values and validate (JSON format)
    run "$TERRAKUBE_CMD" variable list -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --output json
    assert_json_array

    VAL1=$(echo "$output" | jq -r '.[] | select((.attributes.key // .key)=="test1") | (.attributes.value // .value)' 2>/dev/null || true)
    VAL2=$(echo "$output" | jq -r '.[] | select((.attributes.key // .key)=="test2") | (.attributes.value // .value)' 2>/dev/null || true)
    VAL3=$(echo "$output" | jq -r '.[] | select((.attributes.key // .key)=="test3") | (.attributes.value // .value)' 2>/dev/null || true)

    [ "$VAL1" = "a" ]
    [ "$VAL2" = "b" ]
    [ "$VAL3" = "c" ]

    # Validate Table output format
    run "$TERRAKUBE_CMD" variable list -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --output table
    assert_success

    echo "export TERRAKUBE_TEST_E2E_VAR1_ID=\"$VAR1_ID\"" >> "$STATE_FILE"
    echo "export TERRAKUBE_TEST_E2E_VAR2_ID=\"$VAR2_ID\"" >> "$STATE_FILE"
    echo "export TERRAKUBE_TEST_E2E_VAR3_ID=\"$VAR3_ID\"" >> "$STATE_FILE"
}

# ==============================================================================
# Step 5: Workspace Tags CRUD & Validation
# ==============================================================================

@test "5. Add and verify workspace tags" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"
    [ -n "$TERRAKUBE_TEST_E2E_WS_ID" ] || skip "Workspace ID not available"

    # Helper function to find or create an org tag
    get_or_create_tag() {
        local tag_name="$1"
        run "$TERRAKUBE_CMD" tag list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --filter "name==$tag_name" --output json
        local tag_id=""
        if [ "$status" -eq 0 ]; then
            tag_id=$(echo "$output" | jq -r '.[0].id // empty' 2>/dev/null || true)
        fi
        if [ -z "$tag_id" ] || [ "$tag_id" = "null" ]; then
            run "$TERRAKUBE_CMD" tag create -o "$TERRAKUBE_TEST_E2E_ORG_ID" --name "$tag_name" --output json
            if [ "$status" -eq 0 ]; then
                tag_id=$(echo "$output" | jq -r '.id // empty' 2>/dev/null || true)
            fi
        fi
        echo "$tag_id"
    }

    TAG1_ID=$(get_or_create_tag "networking")
    TAG2_ID=$(get_or_create_tag "iac")
    TAG3_ID=$(get_or_create_tag "development")

    [ -n "$TAG1_ID" ] && [ -n "$TAG2_ID" ] && [ -n "$TAG3_ID" ]

    # Associate 3 workspace tags
    run "$TERRAKUBE_CMD" workspace-tag create -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --tag-id "$TAG1_ID" --output json
    assert_success
    WSTAG1_ID=$(echo "$output" | jq -r '.id // empty' 2>/dev/null || true)

    run "$TERRAKUBE_CMD" workspace-tag create -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --tag-id "$TAG2_ID" --output json
    assert_success
    WSTAG2_ID=$(echo "$output" | jq -r '.id // empty' 2>/dev/null || true)

    run "$TERRAKUBE_CMD" workspace-tag create -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --tag-id "$TAG3_ID" --output json
    assert_success
    WSTAG3_ID=$(echo "$output" | jq -r '.id // empty' 2>/dev/null || true)

    [ -n "$WSTAG1_ID" ] && [ -n "$WSTAG2_ID" ] && [ -n "$WSTAG3_ID" ]

    # Read workspace tags values and confirm (JSON format)
    run "$TERRAKUBE_CMD" workspace-tag list -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --output json
    assert_json_array
    [ $(echo "$output" | jq 'length') -ge 3 ]

    # Read workspace tags values (Table format)
    run "$TERRAKUBE_CMD" workspace-tag list -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --output table
    assert_success

    echo "export TERRAKUBE_TEST_E2E_WSTAG1_ID=\"$WSTAG1_ID\"" >> "$STATE_FILE"
    echo "export TERRAKUBE_TEST_E2E_WSTAG2_ID=\"$WSTAG2_ID\"" >> "$STATE_FILE"
    echo "export TERRAKUBE_TEST_E2E_WSTAG3_ID=\"$WSTAG3_ID\"" >> "$STATE_FILE"
}

# ==============================================================================
# Step 6: Provider, Version & Implementation CRUD & Verification
# ==============================================================================

@test "6. Create, verify, and delete provider, version, and implementation" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"

    # 1. Create provider (name: random + 4 alphanumeric)
    RAND_PROV_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    PROV_NAME="random${RAND_PROV_SUFFIX}"
    run "$TERRAKUBE_CMD" provider create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --name "$PROV_NAME" \
        --description "random provider" \
        --output json
    assert_success
    PROV_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$PROV_ID" ] && [ "$PROV_ID" != "null" ]

    # 2. Create provider version 3.0.1
    run "$TERRAKUBE_CMD" provider-version create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --provider "$PROV_ID" \
        --version-number "3.0.1" \
        --protocols "5.0" \
        --output json
    assert_success
    VER_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$VER_ID" ] && [ "$VER_ID" != "null" ]

    # 3. Create provider implementation
    PGP_ARMOR="-----BEGIN PGP PUBLIC KEY BLOCK-----\n\ntest value-----END PGP PUBLIC KEY BLOCK-----"
    run "$TERRAKUBE_CMD" implementation create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --provider "$PROV_ID" \
        --provider-version "$VER_ID" \
        --os "linux" \
        --arch "amd64" \
        --filename "terraform-provider-random_3.0.1_linux_amd64.zip" \
        --download-url "https://releases.hashicorp.com/terraform-provider-random/3.0.1/terraform-provider-random_3.0.1_linux_amd64.zip" \
        --shasums-url "https://releases.hashicorp.com/terraform-provider-random/3.0.1/terraform-provider-random_3.0.1_SHA256SUMS" \
        --shasums-signature-url "https://releases.hashicorp.com/terraform-provider-random/3.0.1/terraform-provider-random_3.0.1_SHA256SUMS.72D7468F.sig" \
        --shasum "e385e00e7425dda9d30b74ab4ffa4636f4b8eb23918c0b763f0ffab84ece0c5c" \
        --key-id "34365D9472D7468F" \
        --ascii-armor "$PGP_ARMOR" \
        --trust-signature "5.0" \
        --source "HashiCorp" \
        --source-url "https://www.hashicorp.com/security.html" \
        --output json
    assert_success
    IMPL_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$IMPL_ID" ] && [ "$IMPL_ID" != "null" ]

    # 4. Check correct values exist using CLI (JSON format)
    run "$TERRAKUBE_CMD" provider get -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$PROV_ID" --output json
    assert_success
    FETCH_PROV_NAME=$(echo "$output" | jq -r '.attributes.name // .name // empty' 2>/dev/null || true)
    [ "$FETCH_PROV_NAME" = "$PROV_NAME" ]

    run "$TERRAKUBE_CMD" provider-version get -o "$TERRAKUBE_TEST_E2E_ORG_ID" --provider "$PROV_ID" --id "$VER_ID" --output json
    assert_success
    FETCH_VER_NUM=$(echo "$output" | jq -r '.attributes.versionNumber // .versionNumber // empty' 2>/dev/null || true)
    [ "$FETCH_VER_NUM" = "3.0.1" ]

    run "$TERRAKUBE_CMD" implementation get -o "$TERRAKUBE_TEST_E2E_ORG_ID" --provider "$PROV_ID" --provider-version "$VER_ID" --id "$IMPL_ID" --output json
    assert_success
    FETCH_IMPL_OS=$(echo "$output" | jq -r '.attributes.os // .os // empty' 2>/dev/null || true)
    FETCH_IMPL_ARCH=$(echo "$output" | jq -r '.attributes.arch // .arch // empty' 2>/dev/null || true)
    [ "$FETCH_IMPL_OS" = "linux" ]
    [ "$FETCH_IMPL_ARCH" = "amd64" ]

    # Check values using CLI (Table format)
    run "$TERRAKUBE_CMD" provider list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --output table
    assert_success

    run "$TERRAKUBE_CMD" provider-version list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --provider "$PROV_ID" --output table
    assert_success

    run "$TERRAKUBE_CMD" implementation list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --provider "$PROV_ID" --provider-version "$VER_ID" --output table
    assert_success

    # 5. Delete values
    run "$TERRAKUBE_CMD" implementation delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" --provider "$PROV_ID" --provider-version "$VER_ID" --id "$IMPL_ID"
    assert_success

    run "$TERRAKUBE_CMD" provider-version delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" --provider "$PROV_ID" --id "$VER_ID"
    assert_success

    run "$TERRAKUBE_CMD" provider delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$PROV_ID"
    assert_success
}

# ==============================================================================
# Step 7: Organization Module CRUD & Verification
# ==============================================================================

@test "7. Create, verify, and delete organization module" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"

    # 1. Create module (name: iam + 4 alphanumeric)
    RAND_MOD_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    MOD_NAME="iam${RAND_MOD_SUFFIX}"
    MOD_SOURCE="https://github.com/terraform-google-modules/terraform-google-iam"

    run "$TERRAKUBE_CMD" module create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --name "$MOD_NAME" \
        --provider "google" \
        --source "$MOD_SOURCE" \
        --description "Initial IAM module" \
        --output json
    assert_success

    MOD_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$MOD_ID" ] && [ "$MOD_ID" != "null" ]

    # 2. Update description using a 4-character random alphanumeric string
    RAND_MOD_DESC=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    run "$TERRAKUBE_CMD" module update -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$MOD_ID" \
        --name "$MOD_NAME" \
        --provider "google" \
        --source "$MOD_SOURCE" \
        --description "$RAND_MOD_DESC" \
        --output json
    assert_success

    # 3. Read values and validate (JSON format)
    run "$TERRAKUBE_CMD" module get -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$MOD_ID" --output json
    assert_success
    FETCHED_MOD_NAME=$(echo "$output" | jq -r '.attributes.name // .name // empty' 2>/dev/null || true)
    FETCHED_MOD_DESC=$(echo "$output" | jq -r '.attributes.description // .description // empty' 2>/dev/null || true)
    [ "$FETCHED_MOD_NAME" = "$MOD_NAME" ]
    [ "$FETCHED_MOD_DESC" = "$RAND_MOD_DESC" ]

    # Read values and validate (Table format)
    run "$TERRAKUBE_CMD" module list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --output table
    assert_success

    # 4. Remove module
    run "$TERRAKUBE_CMD" module delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$MOD_ID"
    assert_success
}

# ==============================================================================
# Step 8: Organization Global Variables CRUD & Verification
# ==============================================================================

@test "8. Add, update, validate, and delete organization global variables" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"

    # 1. Create global variable (key/val: dummy + 4 alphanumeric)
    RAND_VAR_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    VAR_KEY="dummy${RAND_VAR_SUFFIX}"
    VAR_VAL="dummy${RAND_VAR_SUFFIX}"

    run "$TERRAKUBE_CMD" organization-variable create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --key "$VAR_KEY" \
        --value "$VAR_VAL" \
        --category "ENV" \
        --description "Initial global variable" \
        --output json
    assert_success

    GVAR_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$GVAR_ID" ] && [ "$GVAR_ID" != "null" ]

    # 2. Update key and value to new dummy+4alphanumeric strings
    RAND_VAR_SUFFIX2=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    NEW_VAR_KEY="dummy${RAND_VAR_SUFFIX2}"
    NEW_VAR_VAL="dummy${RAND_VAR_SUFFIX2}"

    run "$TERRAKUBE_CMD" organization-variable update -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$GVAR_ID" \
        --key "$NEW_VAR_KEY" \
        --value "$NEW_VAR_VAL" \
        --category "ENV" \
        --description "Updated global variable" \
        --output json
    assert_success

    # 3. Read values and validate (JSON format)
    run "$TERRAKUBE_CMD" organization-variable get -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$GVAR_ID" --output json
    assert_success
    FETCHED_KEY=$(echo "$output" | jq -r '.attributes.key // .key // empty' 2>/dev/null || true)
    FETCHED_VAL=$(echo "$output" | jq -r '.attributes.value // .value // empty' 2>/dev/null || true)
    [ "$FETCHED_KEY" = "$NEW_VAR_KEY" ]
    [ "$FETCHED_VAL" = "$NEW_VAR_VAL" ]

    # Read values and validate (Table format)
    run "$TERRAKUBE_CMD" organization-variable list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --output table
    assert_success

    # 4. Delete global variable
    run "$TERRAKUBE_CMD" organization-variable delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$GVAR_ID"
    assert_success
}

# ==============================================================================
# Step 9: Organization Agent CRUD & Verification
# ==============================================================================

@test "9. Add, update, validate, and delete organization agents" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"

    # 1. Create agent (name: agent + 4 alphanumeric, URL: https://localhost:8080)
    RAND_AGT_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    AGT_NAME="agent${RAND_AGT_SUFFIX}"

    run "$TERRAKUBE_CMD" agent create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --name "$AGT_NAME" \
        --url "https://localhost:8080" \
        --description "Initial agent" \
        --output json
    assert_success

    AGT_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$AGT_ID" ] && [ "$AGT_ID" != "null" ]

    # 2. Update agent URL to https://localhost:8181
    run "$TERRAKUBE_CMD" agent update -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$AGT_ID" \
        --name "$AGT_NAME" \
        --url "https://localhost:8181" \
        --description "Updated agent" \
        --output json
    assert_success

    # 3. Read values and validate (JSON format)
    run "$TERRAKUBE_CMD" agent get -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$AGT_ID" --output json
    assert_success
    FETCHED_AGT_NAME=$(echo "$output" | jq -r '.attributes.name // .name // empty' 2>/dev/null || true)
    FETCHED_AGT_URL=$(echo "$output" | jq -r '.attributes.url // .url // empty' 2>/dev/null || true)
    [ "$FETCHED_AGT_NAME" = "$AGT_NAME" ]
    [ "$FETCHED_AGT_URL" = "https://localhost:8181" ]

    # Read values and validate (Table format)
    run "$TERRAKUBE_CMD" agent list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --output table
    assert_success

    # 4. Delete agent
    run "$TERRAKUBE_CMD" agent delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$AGT_ID"
    assert_success
}

# ==============================================================================
# Step 10: Organization SSH Key CRUD & Verification
# ==============================================================================

@test "10. Add, update, validate, and delete organization SSH keys" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"

    # 1. Create SSH key (name: ssh + 4 alphanumeric)
    RAND_SSH_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    SSH_NAME="ssh${RAND_SSH_SUFFIX}"
    DUMMY_KEY="-----BEGIN RSA PRIVATE KEY-----\ndummykeydata\n-----END RSA PRIVATE KEY-----\n"

    run "$TERRAKUBE_CMD" ssh create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --name "$SSH_NAME" \
        --ssh-type "rsa" \
        --private-key "$DUMMY_KEY" \
        --description "Initial SSH key" \
        --output json
    assert_success

    SSH_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$SSH_ID" ] && [ "$SSH_ID" != "null" ]

    # 2. Update description
    run "$TERRAKUBE_CMD" ssh update -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$SSH_ID" \
        --name "$SSH_NAME" \
        --ssh-type "rsa" \
        --private-key "$DUMMY_KEY" \
        --description "Updated SSH key" \
        --output json
    assert_success

    # 3. Read values and validate (JSON format)
    run "$TERRAKUBE_CMD" ssh get -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$SSH_ID" --output json
    assert_success
    FETCHED_SSH_NAME=$(echo "$output" | jq -r '.attributes.name // .name // empty' 2>/dev/null || true)
    FETCHED_SSH_DESC=$(echo "$output" | jq -r '.attributes.description // .description // empty' 2>/dev/null || true)
    [ "$FETCHED_SSH_NAME" = "$SSH_NAME" ]
    [ "$FETCHED_SSH_DESC" = "Updated SSH key" ]

    # Read values and validate (Table format)
    run "$TERRAKUBE_CMD" ssh list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --output table
    assert_success

    # 4. Delete SSH key
    run "$TERRAKUBE_CMD" ssh delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$SSH_ID"
    assert_success
}

# ==============================================================================
# Step 11: Workspace Access CRUD & Verification (Dedicated Workspace)
# ==============================================================================

@test "11. Create, verify, and delete workspace access with dedicated workspace" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"

    # 1. Create a new dedicated workspace for workspace-access testing
    RAND_WSACC_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    DEDICATED_WS_NAME="wsacc${RAND_WSACC_SUFFIX}"

    run "$TERRAKUBE_CMD" workspace create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --name "$DEDICATED_WS_NAME" \
        --source "https://github.com/terrakube-io/terrakube-docker-compose" \
        --branch "main" \
        --folder "/" \
        --iac-type "tofu" \
        --iac-version "1.12.5" \
        --execution-mode "remote" \
        --output json
    assert_success

    DEDICATED_WS_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$DEDICATED_WS_ID" ] && [ "$DEDICATED_WS_ID" != "null" ]

    # 2. Add workspace access entry for team "TERRAKUBE_ADMIN"
    run "$TERRAKUBE_CMD" workspace-access create -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$DEDICATED_WS_ID" \
        --name "TERRAKUBE_ADMIN" \
        --manage-state \
        --manage-workspace \
        --manage-job \
        --output json
    assert_success

    ACCESS_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$ACCESS_ID" ] && [ "$ACCESS_ID" != "null" ]

    # 3. Update workspace access permissions
    run "$TERRAKUBE_CMD" workspace-access update -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$DEDICATED_WS_ID" \
        --id "$ACCESS_ID" \
        --name "TERRAKUBE_ADMIN" \
        --manage-state=false \
        --manage-workspace \
        --manage-job \
        --output json
    assert_success

    # 4. Read values and validate (JSON format)
    run "$TERRAKUBE_CMD" workspace-access get -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$DEDICATED_WS_ID" --id "$ACCESS_ID" --output json
    assert_success
    FETCHED_ACC_NAME=$(echo "$output" | jq -r '.attributes.name // .name // empty' 2>/dev/null || true)
    [ "$FETCHED_ACC_NAME" = "TERRAKUBE_ADMIN" ]

    # Read values and validate (Table format)
    run "$TERRAKUBE_CMD" workspace-access list -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$DEDICATED_WS_ID" --output table
    assert_success

    # 5. Delete workspace access entry
    run "$TERRAKUBE_CMD" workspace-access delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$DEDICATED_WS_ID" --id "$ACCESS_ID"
    assert_success

    # 6. Soft-delete and cleanup the dedicated workspace
    NEW_DEDICATED_WS_NAME=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 6)
    run "$TERRAKUBE_CMD" workspace update -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$DEDICATED_WS_ID" \
        --name "$NEW_DEDICATED_WS_NAME" \
        --deleted \
        --source "https://github.com/terrakube-io/terrakube-docker-compose" \
        --branch "main" \
        --folder "/" \
        --iac-type "tofu" \
        --iac-version "1.12.5" \
        --execution-mode "remote"
    assert_success
}

# ==============================================================================
# Step 12: Organization Template CRUD & Verification
# ==============================================================================

@test "12. Create, verify, and delete organization template" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"

    # 1. Base64 encode TCL YAML flow
    TCL_RAW="flow:
  - type: \"terraformPlan\"
    step: 100
  - type: \"terraformApply\"
    step: 200"
    TCL_B64=$(echo "$TCL_RAW" | base64 -w 0)

    # 2. Create template (name: tmpl + 4 alphanumeric, description: desc + 4 alphanumeric, version: 1.0.0)
    RAND_TPL_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    TPL_NAME="tmpl${RAND_TPL_SUFFIX}"
    TPL_DESC="desc${RAND_TPL_SUFFIX}"

    run "$TERRAKUBE_CMD" template create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --name "$TPL_NAME" \
        --description "$TPL_DESC" \
        --version "1.0.0" \
        --content "$TCL_B64" \
        --output json
    assert_success

    TPL_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$TPL_ID" ] && [ "$TPL_ID" != "null" ]

    # 3. Update description with a new 4-character random string
    RAND_TPL_SUFFIX2=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    NEW_TPL_DESC="desc${RAND_TPL_SUFFIX2}"

    run "$TERRAKUBE_CMD" template update -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$TPL_ID" \
        --name "$TPL_NAME" \
        --description "$NEW_TPL_DESC" \
        --version "1.0.0" \
        --content "$TCL_B64" \
        --output json
    assert_success

    # 4. Read values and validate (JSON format)
    run "$TERRAKUBE_CMD" template get -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$TPL_ID" --output json
    assert_success
    FETCHED_TPL_NAME=$(echo "$output" | jq -r '.attributes.name // .name // empty' 2>/dev/null || true)
    FETCHED_TPL_DESC=$(echo "$output" | jq -r '.attributes.description // .description // empty' 2>/dev/null || true)
    FETCHED_TPL_VER=$(echo "$output" | jq -r '.attributes.version // .version // empty' 2>/dev/null || true)
    [ "$FETCHED_TPL_NAME" = "$TPL_NAME" ]
    [ "$FETCHED_TPL_DESC" = "$NEW_TPL_DESC" ]
    [ "$FETCHED_TPL_VER" = "1.0.0" ]

    # Read values and validate (Table format)
    run "$TERRAKUBE_CMD" template list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --output table
    assert_success

    # 5. Delete template
    run "$TERRAKUBE_CMD" template delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$TPL_ID"
    assert_success
}

# ==============================================================================
# Step 13: Collection, Collection Item, and Collection Reference CRUD
# ==============================================================================

@test "13. Create, verify, and delete collection, item, and reference with temporary workspace" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"

    # 1. Create temporary workspace for collection-reference testing
    RAND_COLWS_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    COL_WS_NAME="colws${RAND_COLWS_SUFFIX}"

    run "$TERRAKUBE_CMD" workspace create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --name "$COL_WS_NAME" \
        --source "https://github.com/terrakube-io/terrakube-docker-compose" \
        --branch "main" \
        --folder "/" \
        --iac-type "tofu" \
        --iac-version "1.12.5" \
        --execution-mode "remote" \
        --output json
    assert_success

    COL_WS_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$COL_WS_ID" ] && [ "$COL_WS_ID" != "null" ]

    # 2. Create collection
    RAND_COL_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    COL_NAME="col${RAND_COL_SUFFIX}"

    run "$TERRAKUBE_CMD" collection create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --name "$COL_NAME" \
        --description "Initial collection" \
        --priority 10 \
        --output json
    assert_success

    COL_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$COL_ID" ] && [ "$COL_ID" != "null" ]

    # Update collection description
    RAND_COL_DESC2="desc$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)"
    run "$TERRAKUBE_CMD" collection update -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$COL_ID" \
        --name "$COL_NAME" \
        --description "$RAND_COL_DESC2" \
        --priority 20 \
        --output json
    assert_success

    # Validate collection (JSON & Table)
    run "$TERRAKUBE_CMD" collection get -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$COL_ID" --output json
    assert_success
    FETCHED_COL_NAME=$(echo "$output" | jq -r '.attributes.name // .name // empty' 2>/dev/null || true)
    [ "$FETCHED_COL_NAME" = "$COL_NAME" ]

    run "$TERRAKUBE_CMD" collection list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --output table
    assert_success

    # 3. Create collection item (category: ENV)
    RAND_ITEM_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    ITEM_KEY="dummy${RAND_ITEM_SUFFIX}"
    ITEM_VAL="dummy${RAND_ITEM_SUFFIX}"

    run "$TERRAKUBE_CMD" collection-item create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --collection "$COL_ID" \
        --key "$ITEM_KEY" \
        --value "$ITEM_VAL" \
        --category "ENV" \
        --description "Collection item description" \
        --output json
    assert_success

    ITEM_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$ITEM_ID" ] && [ "$ITEM_ID" != "null" ]

    # Validate collection item (JSON & Table)
    run "$TERRAKUBE_CMD" collection-item get -o "$TERRAKUBE_TEST_E2E_ORG_ID" --collection "$COL_ID" --id "$ITEM_ID" --output json
    assert_success
    FETCHED_ITEM_KEY=$(echo "$output" | jq -r '.attributes.key // .key // empty' 2>/dev/null || true)
    [ "$FETCHED_ITEM_KEY" = "$ITEM_KEY" ]

    run "$TERRAKUBE_CMD" collection-item list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --collection "$COL_ID" --output table
    assert_success

    # 4. Create collection reference
    run "$TERRAKUBE_CMD" collection-reference create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --collection "$COL_ID" \
        -w "$COL_WS_ID" \
        --description "Collection reference to workspace" \
        --output json
    assert_success

    REF_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$REF_ID" ] && [ "$REF_ID" != "null" ]

    # Validate collection reference (JSON & Table)
    run "$TERRAKUBE_CMD" collection-reference get -o "$TERRAKUBE_TEST_E2E_ORG_ID" --collection "$COL_ID" -w "$COL_WS_ID" --id "$REF_ID" --output json
    assert_success

    run "$TERRAKUBE_CMD" collection-reference list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --collection "$COL_ID" -w "$COL_WS_ID" --output table
    assert_success

    # 5. Cleanup collection reference, collection item, collection, and temporary workspace
    run "$TERRAKUBE_CMD" collection-reference delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" --collection "$COL_ID" -w "$COL_WS_ID" --id "$REF_ID"
    assert_success

    run "$TERRAKUBE_CMD" collection-item delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" --collection "$COL_ID" --id "$ITEM_ID"
    assert_success

    run "$TERRAKUBE_CMD" collection delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$COL_ID"
    assert_success

    NEW_COL_WS_NAME=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 6)
    run "$TERRAKUBE_CMD" workspace update -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$COL_WS_ID" \
        --name "$NEW_COL_WS_NAME" \
        --deleted \
        --source "https://github.com/terrakube-io/terrakube-docker-compose" \
        --branch "main" \
        --folder "/" \
        --iac-type "tofu" \
        --iac-version "1.12.5" \
        --execution-mode "remote"
    assert_success
}

# ==============================================================================
# Step 14: Project & Project Access CRUD & Verification
# ==============================================================================

@test "14. Create, verify, and delete project and project access with temporary workspace" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"

    # 1. Create temporary workspace for project testing
    RAND_PRJWS_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    PRJ_WS_NAME="prjws${RAND_PRJWS_SUFFIX}"

    run "$TERRAKUBE_CMD" workspace create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --name "$PRJ_WS_NAME" \
        --source "https://github.com/terrakube-io/terrakube-docker-compose" \
        --branch "main" \
        --folder "/" \
        --iac-type "tofu" \
        --iac-version "1.12.5" \
        --execution-mode "remote" \
        --output json
    assert_success

    PRJ_WS_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$PRJ_WS_ID" ] && [ "$PRJ_WS_ID" != "null" ]

    # 2. Create project
    RAND_PRJ_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    PRJ_NAME="proj${RAND_PRJ_SUFFIX}"

    run "$TERRAKUBE_CMD" project create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --name "$PRJ_NAME" \
        --description "Initial project" \
        --output json
    assert_success

    PRJ_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$PRJ_ID" ] && [ "$PRJ_ID" != "null" ]

    # Update project description
    NEW_PRJ_DESC="desc$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)"
    run "$TERRAKUBE_CMD" project update -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$PRJ_ID" \
        --name "$PRJ_NAME" \
        --description "$NEW_PRJ_DESC" \
        --output json
    assert_success

    # Validate project (JSON & Table)
    run "$TERRAKUBE_CMD" project get -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$PRJ_ID" --output json
    assert_success
    FETCHED_PRJ_NAME=$(echo "$output" | jq -r '.attributes.name // .name // empty' 2>/dev/null || true)
    [ "$FETCHED_PRJ_NAME" = "$PRJ_NAME" ]

    run "$TERRAKUBE_CMD" project list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --output table
    assert_success

    # 3. Create project access rule for team TERRAKUBE_ADMIN
    run "$TERRAKUBE_CMD" project-access create -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --project "$PRJ_ID" \
        --name "TERRAKUBE_ADMIN" \
        --manage-state \
        --manage-workspace \
        --manage-job \
        --output json
    assert_success

    PRJ_ACC_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$PRJ_ACC_ID" ] && [ "$PRJ_ACC_ID" != "null" ]

    # Update project access rule
    run "$TERRAKUBE_CMD" project-access update -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --project "$PRJ_ID" \
        --id "$PRJ_ACC_ID" \
        --name "TERRAKUBE_ADMIN" \
        --manage-state=false \
        --manage-workspace \
        --manage-job \
        --output json
    assert_success

    # Validate project access (JSON & Table)
    run "$TERRAKUBE_CMD" project-access get -o "$TERRAKUBE_TEST_E2E_ORG_ID" --project "$PRJ_ID" --id "$PRJ_ACC_ID" --output json
    assert_success
    FETCHED_PRJ_ACC_NAME=$(echo "$output" | jq -r '.attributes.name // .name // empty' 2>/dev/null || true)
    [ "$FETCHED_PRJ_ACC_NAME" = "TERRAKUBE_ADMIN" ]

    run "$TERRAKUBE_CMD" project-access list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --project "$PRJ_ID" --output table
    assert_success

    # 4. Cleanup project access, project, and temporary workspace
    run "$TERRAKUBE_CMD" project-access delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" --project "$PRJ_ID" --id "$PRJ_ACC_ID"
    assert_success

    run "$TERRAKUBE_CMD" project delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$PRJ_ID"
    assert_success

    NEW_PRJ_WS_NAME=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 6)
    run "$TERRAKUBE_CMD" workspace update -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$PRJ_WS_ID" \
        --name "$NEW_PRJ_WS_NAME" \
        --deleted \
        --source "https://github.com/terrakube-io/terrakube-docker-compose" \
        --branch "main" \
        --folder "/" \
        --iac-type "tofu" \
        --iac-version "1.12.5" \
        --execution-mode "remote"
    assert_success
}

# ==============================================================================
# Step 15: Organization Federated Credentials & Claims CRUD & Verification
# ==============================================================================

@test "15. Create, verify, and delete federated identity and claims" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"

    # 1. Create federated identity credential (name: fed + 4 alphanumeric)
    RAND_FED_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    FED_NAME="fed${RAND_FED_SUFFIX}"

    run "$TERRAKUBE_CMD" federated create \
        --name "$FED_NAME" \
        --issuer-url "https://token.actions.githubusercontent.com" \
        --audience "api://Terrakube" \
        --output json
    assert_success

    FED_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$FED_ID" ] && [ "$FED_ID" != "null" ]

    # Update audience
    run "$TERRAKUBE_CMD" federated update \
        --id "$FED_ID" \
        --name "$FED_NAME" \
        --issuer-url "https://token.actions.githubusercontent.com" \
        --audience "api://TerrakubeV2" \
        --output json
    assert_success

    # Validate federated credential (JSON & Table)
    run "$TERRAKUBE_CMD" federated get --id "$FED_ID" --output json
    assert_success
    FETCHED_FED_NAME=$(echo "$output" | jq -r '.attributes.name // .name // empty' 2>/dev/null || true)
    FETCHED_FED_AUD=$(echo "$output" | jq -r '.attributes.audience // .audience // empty' 2>/dev/null || true)
    [ "$FETCHED_FED_NAME" = "$FED_NAME" ]
    [ "$FETCHED_FED_AUD" = "api://TerrakubeV2" ]

    run "$TERRAKUBE_CMD" federated list --output table
    assert_success

    # 2. Create federated claim (claim-key: repository_owner, claim-value: dummy + 4 alphanumeric)
    RAND_CLAIM_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    CLAIM_VAL="dummy${RAND_CLAIM_SUFFIX}"

    run "$TERRAKUBE_CMD" federated-claim create \
        --federated "$FED_ID" \
        --claim-key "repository_owner" \
        --claim-value "$CLAIM_VAL" \
        --output json
    assert_success

    CLAIM_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$CLAIM_ID" ] && [ "$CLAIM_ID" != "null" ]

    # Update claim value
    RAND_CLAIM_SUFFIX2=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    NEW_CLAIM_VAL="dummy${RAND_CLAIM_SUFFIX2}"
    run "$TERRAKUBE_CMD" federated-claim update \
        --federated "$FED_ID" \
        --id "$CLAIM_ID" \
        --claim-key "repository_owner" \
        --claim-value "$NEW_CLAIM_VAL" \
        --output json
    assert_success

    # Validate federated claim (JSON & Table)
    run "$TERRAKUBE_CMD" federated-claim get --federated "$FED_ID" --id "$CLAIM_ID" --output json
    assert_success
    FETCHED_CLAIM_VAL=$(echo "$output" | jq -r '.attributes.claimValue // .claimValue // empty' 2>/dev/null || true)
    [ "$FETCHED_CLAIM_VAL" = "$NEW_CLAIM_VAL" ]

    run "$TERRAKUBE_CMD" federated-claim list --federated "$FED_ID" --output table
    assert_success

    # 3. Cleanup federated claim and federated identity credential
    run "$TERRAKUBE_CMD" federated-claim delete --federated "$FED_ID" --id "$CLAIM_ID"
    assert_success

    run "$TERRAKUBE_CMD" federated delete --id "$FED_ID"
    assert_success
}

# ==============================================================================
# Step 16: Notification Configuration & Trigger CRUD & Verification
# ==============================================================================

@test "16. Create, verify, and delete notification configuration and triggers" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"

    # 1. Create organization notification configuration (Slack)
    RAND_NOTIF_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    NC_NAME="notif${RAND_NOTIF_SUFFIX}"
    NC_URL="https://hooks.slack.com/services/T00/B00/${RAND_NOTIF_SUFFIX}"

    run "$TERRAKUBE_CMD" notification-configuration create \
        -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --name "$NC_NAME" \
        --description "Initial notification config" \
        --channel-type "SLACK" \
        --destination-url "$NC_URL" \
        --message-style "DETAILED" \
        --active \
        --output json
    assert_success

    NC_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$NC_ID" ] && [ "$NC_ID" != "null" ]

    # 2. Update description and message style
    RAND_NC_DESC=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 6)
    run "$TERRAKUBE_CMD" notification-configuration update \
        -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$NC_ID" \
        --description "$RAND_NC_DESC" \
        --message-style "SIMPLE" \
        --output json
    assert_success

    # 3. Read values and validate (JSON & Table formats)
    run "$TERRAKUBE_CMD" notification-configuration get \
        -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$NC_ID" \
        --output json
    assert_success
    FETCHED_NC_NAME=$(echo "$output" | jq -r '.attributes.name // .name // empty' 2>/dev/null || true)
    FETCHED_NC_DESC=$(echo "$output" | jq -r '.attributes.description // .description // empty' 2>/dev/null || true)
    FETCHED_NC_STYLE=$(echo "$output" | jq -r '.attributes.messageStyle // .messageStyle // empty' 2>/dev/null || true)
    [ "$FETCHED_NC_NAME" = "$NC_NAME" ]
    [ "$FETCHED_NC_DESC" = "$RAND_NC_DESC" ]
    [ "$FETCHED_NC_STYLE" = "SIMPLE" ]

    run "$TERRAKUBE_CMD" notification-configuration list \
        -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --output table
    assert_success

    # 4. Create notification trigger for completed jobs
    run "$TERRAKUBE_CMD" notification-trigger create \
        -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --notification-configuration "$NC_ID" \
        --job-status "completed" \
        --output json
    assert_success

    NT_ID=$(echo "$output" | jq -r '.id // .attributes.id // empty' 2>/dev/null || true)
    [ -n "$NT_ID" ] && [ "$NT_ID" != "null" ]

    # 5. Read notification trigger and validate (JSON & Table formats)
    run "$TERRAKUBE_CMD" notification-trigger get \
        -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --notification-configuration "$NC_ID" \
        --id "$NT_ID" \
        --output json
    assert_success
    FETCHED_JOB_STATUS=$(echo "$output" | jq -r '.attributes.jobStatus // .jobStatus // empty' 2>/dev/null || true)
    [ "$FETCHED_JOB_STATUS" = "completed" ]

    run "$TERRAKUBE_CMD" notification-trigger list \
        -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --notification-configuration "$NC_ID" \
        --output table
    assert_success

    # 6. Delete notification trigger
    run "$TERRAKUBE_CMD" notification-trigger delete \
        -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --notification-configuration "$NC_ID" \
        --id "$NT_ID"
    assert_success

    # 7. Workspace-level notification configuration with dedicated temporary workspace
    RAND_WS_SUFFIX=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 4)
    TEMP_WS_NAME="wsnotif${RAND_WS_SUFFIX}"
    run "$TERRAKUBE_CMD" workspace create \
        -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --name "$TEMP_WS_NAME" \
        --source "https://github.com/terrakube-io/terrakube-docker-compose" \
        --branch "main" \
        --folder "/" \
        --iac-type "tofu" \
        --iac-version "1.12.5" \
        --execution-mode "remote" \
        --output json
    assert_success
    TEMP_WS_ID=$(echo "$output" | jq -r '.id // empty' 2>/dev/null || true)
    [ -n "$TEMP_WS_ID" ]

    WS_NC_NAME="wsnc${RAND_WS_SUFFIX}"
    WS_NC_URL="https://outlook.office.com/webhook/${RAND_WS_SUFFIX}"
    run "$TERRAKUBE_CMD" notification-configuration create \
        -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        -w "$TEMP_WS_ID" \
        --name "$WS_NC_NAME" \
        --channel-type "TEAMS" \
        --destination-url "$WS_NC_URL" \
        --active \
        --output json
    assert_success
    WS_NC_ID=$(echo "$output" | jq -r '.id // empty' 2>/dev/null || true)
    [ -n "$WS_NC_ID" ]

    run "$TERRAKUBE_CMD" notification-configuration list \
        -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        -w "$TEMP_WS_ID" \
        --output json
    assert_success

    run "$TERRAKUBE_CMD" notification-configuration list \
        -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        -w "$TEMP_WS_ID" \
        --output table
    assert_success

    run "$TERRAKUBE_CMD" notification-configuration delete \
        -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$WS_NC_ID"
    assert_success

    # Soft-delete temporary workspace
    run "$TERRAKUBE_CMD" workspace update \
        -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$TEMP_WS_ID" \
        --deleted
    assert_success

    # 8. Delete organization notification configuration
    run "$TERRAKUBE_CMD" notification-configuration delete \
        -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$NC_ID"
    assert_success
}

# ==============================================================================
# Step 17: Workspace Tag, Variable & Soft-Delete Cleanup
# ==============================================================================

@test "17. Delete tags, variables, and soft-delete workspace" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"
    [ -n "$TERRAKUBE_TEST_E2E_WS_ID" ] || skip "Workspace ID not available"

    # Delete each workspace tag
    if [ -n "$TERRAKUBE_TEST_E2E_WSTAG1_ID" ]; then
        run "$TERRAKUBE_CMD" workspace-tag delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --id "$TERRAKUBE_TEST_E2E_WSTAG1_ID"
        assert_success
    fi
    if [ -n "$TERRAKUBE_TEST_E2E_WSTAG2_ID" ]; then
        run "$TERRAKUBE_CMD" workspace-tag delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --id "$TERRAKUBE_TEST_E2E_WSTAG2_ID"
        assert_success
    fi
    if [ -n "$TERRAKUBE_TEST_E2E_WSTAG3_ID" ]; then
        run "$TERRAKUBE_CMD" workspace-tag delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --id "$TERRAKUBE_TEST_E2E_WSTAG3_ID"
        assert_success
    fi

    # Delete each environment variable
    if [ -n "$TERRAKUBE_TEST_E2E_VAR1_ID" ]; then
        run "$TERRAKUBE_CMD" variable delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --id "$TERRAKUBE_TEST_E2E_VAR1_ID"
        assert_success
    fi
    if [ -n "$TERRAKUBE_TEST_E2E_VAR2_ID" ]; then
        run "$TERRAKUBE_CMD" variable delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --id "$TERRAKUBE_TEST_E2E_VAR2_ID"
        assert_success
    fi
    if [ -n "$TERRAKUBE_TEST_E2E_VAR3_ID" ]; then
        run "$TERRAKUBE_CMD" variable delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" -w "$TERRAKUBE_TEST_E2E_WS_ID" --id "$TERRAKUBE_TEST_E2E_VAR3_ID"
        assert_success
    fi

    # Update workspace name to 6-character alphanumeric string and set deleted=true
    NEW_WS_NAME=$(head /dev/urandom | tr -dc 'a-z0-9' | head -c 6)
    run "$TERRAKUBE_CMD" workspace update -o "$TERRAKUBE_TEST_E2E_ORG_ID" \
        --id "$TERRAKUBE_TEST_E2E_WS_ID" \
        --name "$NEW_WS_NAME" \
        --deleted \
        --source "https://github.com/terrakube-io/terrakube-docker-compose" \
        --branch "main" \
        --folder "/" \
        --iac-type "tofu" \
        --iac-version "1.12.5" \
        --execution-mode "remote"
    assert_success

    # List all workspaces inside organization to validate the workspace does not exist
    run "$TERRAKUBE_CMD" workspace list -o "$TERRAKUBE_TEST_E2E_ORG_ID" --output json
    assert_success
    # Soft deleted workspaces should not be returned in standard active workspace list or deleted==true
    MATCH_COUNT=$(echo "$output" | jq "if type==\"array\" then [.[] | select((.id == \"$TERRAKUBE_TEST_E2E_WS_ID\") and ((.attributes.deleted // .deleted) != true))] else [] end | length")
    [ "$MATCH_COUNT" -eq 0 ]
}

# ==============================================================================
# Step 18: Final Team Cleanup (Always Executed Last)
# ==============================================================================

@test "18. Delete TERRAKUBE_ADMIN team" {
    [ -n "$TERRAKUBE_TEST_E2E_ORG_ID" ] || skip "Organization ID not available"
    [ -n "$TERRAKUBE_TEST_E2E_TEAM_ID" ] || skip "Team ID not available"

    run "$TERRAKUBE_CMD" team delete -o "$TERRAKUBE_TEST_E2E_ORG_ID" --id "$TERRAKUBE_TEST_E2E_TEAM_ID"
    assert_success

    # Organization "bats" is NOT deleted and kept for future executions.
}
