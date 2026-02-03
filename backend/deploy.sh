#!/bin/bash
# Deploy SIP Allocator to Cloud Run Jobs

set -e

PROJECT_ID="${GCP_PROJECT_ID:-your-project-id}"
REGION="${GCP_REGION:-asia-south1}"
JOB_NAME="sip-allocator"
IMAGE="gcr.io/${PROJECT_ID}/${JOB_NAME}"

echo "=== Building and Deploying SIP Allocator ==="

# Build and push Docker image
echo "Building Docker image..."
docker build -t ${IMAGE} .

echo "Pushing to Container Registry..."
docker push ${IMAGE}

# Create/update Cloud Run Job
echo "Deploying Cloud Run Job..."
gcloud run jobs deploy ${JOB_NAME} \
  --image ${IMAGE} \
  --region ${REGION} \
  --set-env-vars "TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}" \
  --set-env-vars "TELEGRAM_CHAT_ID=${TELEGRAM_CHAT_ID}" \
  --set-env-vars "MONTHLY_SIP_AMOUNT=${MONTHLY_SIP_AMOUNT:-35000}" \
  --set-env-vars "DEBT_RESERVE=${DEBT_RESERVE:-0}" \
  --max-retries 1 \
  --task-timeout 60s \
  --memory 256Mi

# Create Cloud Scheduler trigger (runs at 12:00 PM IST daily)
echo "Creating Cloud Scheduler trigger..."
gcloud scheduler jobs create http ${JOB_NAME}-trigger \
  --location ${REGION} \
  --schedule "0 12 * * *" \
  --time-zone "Asia/Kolkata" \
  --uri "https://${REGION}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${PROJECT_ID}/jobs/${JOB_NAME}:run" \
  --http-method POST \
  --oauth-service-account-email "${PROJECT_ID}@appspot.gserviceaccount.com" \
  2>/dev/null || \
gcloud scheduler jobs update http ${JOB_NAME}-trigger \
  --location ${REGION} \
  --schedule "0 12 * * *" \
  --time-zone "Asia/Kolkata"

echo "=== Deployment Complete ==="
echo "Job: ${JOB_NAME}"
echo "Trigger: Daily at 12:00 PM IST"
echo ""
echo "To run manually: gcloud run jobs execute ${JOB_NAME} --region ${REGION}"
