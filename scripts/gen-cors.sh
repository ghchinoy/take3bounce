#!/bin/bash
# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# gen-cors.sh — emit the GCS bucket CORS configuration to stdout.
#
# Single source of truth: the bucket's allowed origins are derived from
# $ALLOWED_ORIGINS (the same comma-separated list the API middleware uses), so
# the Download button (a cross-origin fetch() of the firebasestorage token URL)
# works for exactly the origins allowed to call the API. No wildcard is ever
# emitted. When $ALLOWED_ORIGINS is empty, the committed cors.json (dev default:
# http://localhost:5173) is emitted unchanged.
#
# Usage:
#   ALLOWED_ORIGINS="https://app.run.app,http://localhost:5173" scripts/gen-cors.sh > /tmp/cors.json
set -e

cd "$(dirname "$0")/.."

ORIGINS="${ALLOWED_ORIGINS:-}"

# Empty -> fall back to the committed dev-default cors.json (no wildcard).
if [ -z "${ORIGINS// /}" ]; then
    cat cors.json
    exit 0
fi

# Build a JSON array of exact origins from the comma-separated list, trimming
# whitespace and dropping empty entries.
json_origins=""
IFS=','
for o in $ORIGINS; do
    # Trim leading/trailing whitespace.
    o="${o#"${o%%[![:space:]]*}"}"
    o="${o%"${o##*[![:space:]]}"}"
    [ -z "$o" ] && continue
    if [ -z "$json_origins" ]; then
        json_origins="\"$o\""
    else
        json_origins="$json_origins, \"$o\""
    fi
done
unset IFS

# If the list contained only separators/whitespace, fall back to the committed file.
if [ -z "$json_origins" ]; then
    cat cors.json
    exit 0
fi

cat <<EOF
[
  {
    "origin": [$json_origins],
    "method": ["GET", "HEAD"],
    "responseHeader": ["Content-Type", "Content-Length"],
    "maxAgeSeconds": 3600
  }
]
EOF
