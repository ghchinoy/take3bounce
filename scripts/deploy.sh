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

# Deploy Three-Up TTS Generator
set -e

# Always run from the project root
cd "$(dirname "$0")/.."

# --- Load Configuration ---
if [ -f ".env.deploy" ]; then
    echo "Loading config from .env.deploy..."
    export $(grep -v '^#' .env.deploy | xargs)
elif [ -f "backend/.env" ]; then
    echo "Loading config from backend/.env..."
    export $(grep -v '^#' backend/.env | xargs)
else
    echo "Warning: No .env.deploy or backend/.env found."
fi

PROJECT_ID=${PROJECT_ID:-$(gcloud config get project)}
REGION=${REGION:-"us-central1"}
# SERVICE_NAME="three-up-generator"
SERVICE_NAME=${SERVICE_NAME:-"repotron-agent"} # reusing an existing cloud run service
SERVICE_ACCOUNT="threeup-sa@${PROJECT_ID}.iam.gserviceaccount.com"
IMAGE_TAG="gcr.io/${PROJECT_ID}/${SERVICE_NAME}"
USE_IAP=${USE_IAP:-"false"}
EAP_GROUP=${EAP_GROUP:-"aaie-musicbox@google.com"}

# Fallbacks
: "${GOOGLE_CLOUD_PROJECT:=${PROJECT_ID}}"
: "${GENMEDIA_BUCKET:=aaie-speech-arena}"

echo "====================================================="
echo " Deploying ${SERVICE_NAME} to Cloud Run in ${REGION}"
echo "====================================================="

echo "Checking Service Account..."
if ! gcloud iam service-accounts describe "${SERVICE_ACCOUNT}" --project "${PROJECT_ID}" >/dev/null 2>&1; then
    echo "Creating Service Account: ${SERVICE_ACCOUNT}"
    gcloud iam service-accounts create threeup-sa \
        --display-name="Three-Up Generator SA" \
        --project="${PROJECT_ID}"
fi

echo "Granting Vertex AI User role..."
gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/aiplatform.user" \
    --condition=None --quiet >/dev/null

echo "Granting Storage Object Admin role..."
gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/storage.objectAdmin" \
    --condition=None --quiet >/dev/null

# 1. Build image
echo "Building App Image..."
gcloud builds submit --tag ${IMAGE_TAG} .

# 2. Deploy to Cloud Run
echo "Deploying to Cloud Run..."
if [ "${USE_IAP}" = "true" ]; then
    echo "Using IAP for UX service..."
    IAP_FLAGS="--iap --no-allow-unauthenticated"
else
    echo "Deploying public UX service (Native Auth)..."
    IAP_FLAGS="--allow-unauthenticated"
fi

gcloud run deploy ${SERVICE_NAME} \
    --region ${REGION} \
    --project ${PROJECT_ID} \
    --service-account="${SERVICE_ACCOUNT}" \
    ${IAP_FLAGS} \
    --image ${IMAGE_TAG} \
    --port 8080 \
    --set-env-vars GOOGLE_CLOUD_PROJECT=${GOOGLE_CLOUD_PROJECT},GOOGLE_CLOUD_LOCATION=${GOOGLE_CLOUD_LOCATION},GENMEDIA_BUCKET=${GENMEDIA_BUCKET},GEMINI_MODEL=${GEMINI_MODEL},GEMINI_TTS_MODEL=${GEMINI_TTS_MODEL} \
    --quiet

echo "🔐 Configuring Cloud Run Invoker Access for Service Account..."
gcloud run services add-iam-policy-binding "${SERVICE_NAME}" \
    --region="${REGION}" \
    --project="${PROJECT_ID}" \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/run.invoker" \
    --quiet

if [ "${USE_IAP}" = "true" ]; then
    echo "🔐 Configuring IAP Access for Service Account..."
    gcloud iap web add-iam-policy-binding \
        --member="serviceAccount:${SERVICE_ACCOUNT}" \
        --role="roles/iap.httpsResourceAccessor" \
        --resource-type="cloud-run" \
        --region="$REGION" \
        --service="$SERVICE_NAME" \
        --project="$PROJECT_ID" \
        --condition=None \
        --quiet

    if [ -n "${EAP_GROUP}" ]; then
        echo "🔐 Configuring IAP Access for ${EAP_GROUP}..."
        gcloud iap web add-iam-policy-binding \
            --member="group:${EAP_GROUP}" \
            --role="roles/iap.httpsResourceAccessor" \
            --resource-type="cloud-run" \
            --region="$REGION" \
            --service="$SERVICE_NAME" \
            --project="$PROJECT_ID" \
            --condition=None \
            --quiet
    fi
fi

echo "Deployment completed successfully! Your app is live at: $(gcloud run services describe ${SERVICE_NAME} --project ${PROJECT_ID} --region ${REGION} --format=value\(status.url\))"