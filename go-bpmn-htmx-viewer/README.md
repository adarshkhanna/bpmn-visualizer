# Go BPMN Viewer with HTMX and GCS Integration

This web application allows users to render BPMN 2.0 diagrams stored in Google Cloud Storage (GCS). It uses a Go backend, HTMX for dynamic frontend updates, and the `bpmn-js` library for rendering diagrams in the browser.

## Features

-   Specify a GCS URI for a BPMN 2.0 XML file.
-   Backend fetches the file from GCS.
-   Diagram is rendered in the browser using `bpmn-js`.
-   Dynamic updates using HTMX without full page reloads.
-   Containerized for deployment using Docker.

## Core Technologies

-   **Backend:** Go (`net/http`, `html/template`)
-   **Frontend:** HTML5, HTMX, `bpmn-js`
-   **Cloud:** Google Cloud Storage (GCS), Google Cloud Run (for deployment)

## Project Structure

```
/go-bpmn-htmx-viewer
|-- /cmd/server                 # Main application (main.go)
|-- /internal                   # Internal packages
|   |-- /api_handler            # HTTP API handlers
|   |-- /gcs_handler            # GCS interaction logic
|-- /static                     # Static files (CSS)
|-- /templates                  # HTML templates (index.html, bpmn_fragment.html)
|-- go.mod                      # Go module definition
|-- go.sum                      # Go module checksums
|-- Dockerfile                  # For building the Docker image
|-- README.md                   # This file
```

## Prerequisites for Local Development

-   Go (version 1.21 or later recommended)
-   Access to a Google Cloud Storage bucket.
-   A BPMN 2.0 XML file uploaded to your GCS bucket.
-   Google Cloud SDK (`gcloud`) configured for Application Default Credentials (ADC) if you want to access GCS locally using your user credentials. Run `gcloud auth application-default login`.

## Build and Run Locally

1.  **Clone the repository (if applicable):**
    ```bash
    git clone <repository-url>
    cd go-bpmn-htmx-viewer
    ```

2.  **Set up Application Default Credentials (ADC) for GCS access:**
    If you haven't already, authenticate `gcloud` for ADC:
    ```bash
    gcloud auth application-default login
    ```
    This allows the Go GCS client library to automatically find your credentials for local development. Ensure the authenticated user has `Storage Object Viewer` permissions on the target GCS bucket.

3.  **Build the application:**
    ```bash
    go build -o go-bpmn-viewer cmd/server/main.go
    ```
    This will create an executable named `go-bpmn-viewer` in the project root.

4.  **Run the application:**
    ```bash
    ./go-bpmn-viewer
    ```
    The server will start, typically on port 8080 (or the port specified by the `PORT` environment variable). You should see a log message like:
    `YYYY/MM/DD HH:MM:SS Server starting on port 8080...`

5.  **Access the application:**
    Open your web browser and go to `http://localhost:8080`.

6.  **Test with a GCS URI:**
    Enter the GCS URI of your BPMN file (e.g., `gs://your-bucket-name/path/to/diagram.bpmn`) into the input field and click "Render Diagram".

## Build and Run with Docker (Alternative Local Test)

1.  **Build the Docker image:**
    ```bash
    docker build -t go-bpmn-viewer-app .
    ```

2.  **Run the Docker container:**
    To allow the container to use your local Google Cloud ADC for GCS access, you need to mount the ADC configuration file into the container.

    *   **Linux/macOS:**
        ```bash
        docker run -p 8080:8080           -v ~/.config/gcloud/application_default_credentials.json:/root/.config/gcloud/application_default_credentials.json:ro           --env GOOGLE_APPLICATION_CREDENTIALS=/root/.config/gcloud/application_default_credentials.json           go-bpmn-viewer-app
        ```
    *   **Windows (PowerShell - path might vary):**
        ```bash
        docker run -p 8080:8080 `
          -v $env:APPDATA\gcloud\application_default_credentials.json:/root/.config/gcloud/application_default_credentials.json:ro `
          --env GOOGLE_APPLICATION_CREDENTIALS=/root/.config/gcloud/application_default_credentials.json `
          go-bpmn-viewer-app
        ```
    *Note: The path to `application_default_credentials.json` might differ based on your OS and configuration. The `--env GOOGLE_APPLICATION_CREDENTIALS` tells the Go client library inside the container where to find these credentials.*

    The application inside the container will listen on port 8080, which is mapped to port 8080 on your host.

3.  **Access the application:**
    Open your web browser and go to `http://localhost:8080`.

## Deployment to Google Cloud Run

Follow these steps to deploy the application to Google Cloud Run.

**1. Prerequisites:**

*   **Google Cloud SDK (`gcloud`):** Installed and configured (`gcloud init`).
*   **Google Cloud Project:** A project created in the Google Cloud Console with billing enabled.
*   **APIs Enabled:** Ensure the following APIs are enabled for your project:
    *   Cloud Build API (for building the container)
    *   Artifact Registry API (or Google Container Registry API)
    *   Cloud Run API
    You can enable them via the console or `gcloud services enable <SERVICE_NAME>`.
    Example:
    ```bash
    gcloud services enable run.googleapis.com cloudbuild.googleapis.com artifactregistry.googleapis.com
    ```

**2. Service Account (Recommended):**

While Cloud Run services can use the default Compute Engine service account, it's best practice to create a dedicated service account with minimal permissions.

*   **Create a Service Account:**
    ```bash
    gcloud iam service-accounts create go-bpmn-viewer-sa         --description="Service account for Go BPMN Viewer on Cloud Run"         --display-name="Go BPMN Viewer SA"
    ```
    Replace `go-bpmn-viewer-sa` with your preferred service account name.

*   **Grant IAM Permissions:**
    The service account needs permission to read from GCS buckets.
    ```bash
    # Get your project ID
    PROJECT_ID=$(gcloud config get-value project)

    # Grant Storage Object Viewer role to the service account for a specific bucket
    # Replace YOUR_BUCKET_NAME with the actual bucket name
    gcloud storage buckets add-iam-policy-binding gs://YOUR_BUCKET_NAME         --member="serviceAccount:go-bpmn-viewer-sa@${PROJECT_ID}.iam.gserviceaccount.com"         --role="roles/storage.objectViewer"

    # If you want it to access *any* bucket (less secure, use with caution):
    # gcloud projects add-iam-policy-binding ${PROJECT_ID}     #     --member="serviceAccount:go-bpmn-viewer-sa@${PROJECT_ID}.iam.gserviceaccount.com"     #     --role="roles/storage.objectViewer"
    ```
    *Important: Grant permissions only to the specific bucket(s) the application needs to access.*

**3. Configure Artifact Registry (Optional, but Recommended over GCR):**

If you haven't already, create an Artifact Registry Docker repository:
```bash
gcloud artifacts repositories create go-bpmn-repo     --repository-format=docker     --location=YOUR_REGION     --description="Docker repository for Go BPMN Viewer"
# Replace YOUR_REGION with your preferred region, e.g., us-central1
```

**4. Building the Container Image (using Cloud Build):**

Cloud Build will use the `Dockerfile` in your project root.

*   **If using Artifact Registry:**
    ```bash
    gcloud builds submit --tag YOUR_REGION-docker.pkg.dev/YOUR_PROJECT_ID/go-bpmn-repo/go-bpmn-viewer:latest .
    ```
    Replace `YOUR_REGION`, `YOUR_PROJECT_ID`, and `go-bpmn-repo` accordingly.

*   **If using Google Container Registry (GCR):**
    ```bash
    gcloud builds submit --tag gcr.io/YOUR_PROJECT_ID/go-bpmn-viewer:latest .
    ```
    Replace `YOUR_PROJECT_ID`.

**5. Deploying to Cloud Run:**

*   **Deploy Command:**
    ```bash
    # Replace placeholders: YOUR_PROJECT_ID, YOUR_REGION, YOUR_SERVICE_ACCOUNT_EMAIL, IMAGE_URI
    PROJECT_ID=$(gcloud config get-value project)
    REGION="us-central1" # Choose your preferred region
    SERVICE_ACCOUNT_EMAIL="go-bpmn-viewer-sa@${PROJECT_ID}.iam.gserviceaccount.com"
    
    # For Artifact Registry:
    IMAGE_URI="${REGION}-docker.pkg.dev/${PROJECT_ID}/go-bpmn-repo/go-bpmn-viewer:latest"
    # For GCR:
    # IMAGE_URI="gcr.io/${PROJECT_ID}/go-bpmn-viewer:latest"

    gcloud run deploy go-bpmn-htmx-viewer         --image "${IMAGE_URI}"         --platform managed         --region "${REGION}"         --service-account "${SERVICE_ACCOUNT_EMAIL}"         --port 8080 \ # Port your application listens on (matches Dockerfile EXPOSE and app config)
        --allow-unauthenticated \ # To make the service publicly accessible. Remove for private services.
        --project "${PROJECT_ID}"
        # Add other flags as needed, e.g., --memory, --cpu, --max-instances
    ```

*   **Explanation of Key Flags:**
    *   `go-bpmn-htmx-viewer`: The name of your Cloud Run service.
    *   `--image`: The full path to your container image in Artifact Registry or GCR.
    *   `--platform managed`: Specifies the fully managed Cloud Run environment.
    *   `--region`: The region where the service will be deployed.
    *   `--service-account`: The email of the service account the Cloud Run revision will use. This grants the service its GCS access permissions.
    *   `--port 8080`: The port your container listens on. Cloud Run automatically sends requests to this port. The `PORT` environment variable is also set by Cloud Run for your application.
    *   `--allow-unauthenticated`: Makes the service publicly accessible. For restricted access, configure IAM (`roles/run.invoker`).
    *   `--project`: Specifies the GCP project ID.

**6. Accessing the Deployed Service:**

After successful deployment, Cloud Run will provide a URL for your service. You can also find it in the Google Cloud Console under Cloud Run.

**7. Environment Variables in Cloud Run:**

-   The `PORT` environment variable is automatically set by Cloud Run and your application should respect it (which it does if it reads `os.Getenv("PORT")`).
-   You can set other environment variables via the `--set-env-vars` flag during deployment or through the Cloud Console if needed.

## Code Review Aspects (Self-Check)

-   **Go Best Practices:**
    -   Error handling (`if err != nil`).
    -   Context handling for requests.
    -   Code readability and comments.
    -   Correct use of Go standard library and GCS client.
    -   Security: Input validation (GCS URI), proper escaping in templates (Go's `html/template` helps significantly).
-   **HTMX Integration:**
    -   Correct HTMX attributes for desired behavior.
    -   Efficient fragment updates.
-   **`bpmn-js` Integration:**
    -   Correct initialization and usage.
    -   Robust handling of XML import and errors.
-   **Concurrency:** Server handles multiple requests correctly (Go's `net/http` handles this well by default).
-   **Dependencies:** `go.mod` is clean.

This `README.md` provides a comprehensive guide for users and developers of the application.
```
