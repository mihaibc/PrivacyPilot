# 🛡️ PrivacyPilot - Advanced Backend & AI Showcase

**PrivacyPilot** is an open-source, privacy-focused backend platform designed to automatically detect, moderate, and anonymize sensitive personal information using advanced AI models.

**Beyond its core functionality, this project serves as a comprehensive showcase** demonstrating proficiency in:
*   Modern backend development with a polyglot approach (**Go, Node.js, Perl, Python**) using idiomatic practices (including local Go module management).
*   Microservice architecture and distributed systems design.
*   Integration with various AI services (**Ollama** via its official Go library, Azure AI planned).
*   End-to-end DevOps practices (Containerization, IaC, Orchestration, CI/CD).
*   Robust observability and monitoring.

It ensures data privacy, security, and compliance (e.g., GDPR principles), making it a blueprint for building sophisticated, real-world applications.

---

## 🌟 Key Features

- ✅ **Real-time Data Anonymization**: Protect user identities by anonymizing sensitive textual data instantly via dedicated microservices.
- ✅ **Automated Content Moderation**: Placeholder for AI-driven moderation of harmful or inappropriate content (Azure AI integration planned).
- ✅ **Flexible AI Integration**: Pluggable AI architecture via an **AI Coordinator**. Currently supports **Ollama** (using official Go client), allowing dynamic model selection per request (e.g., Gemma, Mistral, Llama). Azure AI/Stable Diffusion planned.
- ✅ **Scalable Microservice Architecture**: Efficient, reliable microservices built with **Go**, **Node.js**, **Perl**, and **Python**, communicating via **REST APIs** and potentially a **RabbitMQ** message queue (planned).
- ✅ **Robust Infrastructure & DevOps**:
    - Containerized with **Docker**.
    - Local development via **Docker Compose**.
    - Production-ready orchestration with **Kubernetes** (managed by **Helm**) (planned).
    - Infrastructure provisioned using **Terraform** (planned).
    - Automated CI/CD pipelines via **GitHub Actions** (basic setup exists).
- ✅ **Privacy and Security Compliance**: GDPR-aware design principles, **OAuth2/OIDC** secured APIs (planned), secure data handling practices.
- ✅ **Comprehensive Observability**: Basic setup for Prometheus, Grafana, Jaeger via Docker Compose (instrumentation needed). Standardized **JSON logging**.
- ✅ **Data Persistence**: Utilizes **MongoDB** and **Redis** via Docker Compose.
- ✅ **Formal API Contracts**: APIs defined using **OpenAPI 3.0** (planned for `api-specs/`).

---

## 🎯 Showcase Goals

This project intentionally incorporates a diverse set of technologies and practices:

*   **Polyglot Microservices:** Demonstrates choosing the right tool (language/framework) for the job (Go for performance, Node.js for I/O & ecosystem, Python for AI, Perl for specific scripting) and managing a heterogeneous environment.
*   **Go Best Practices:** Uses idiomatic Go, including proper local module management for internal project dependencies.
*   **Cloud-Native Principles:** Leverages containers, local orchestration, service discovery, preparing for future K8s deployment and IaC.
*   **End-to-End DevOps:** Implements local development, build, and run lifecycle, preparing for automated testing, CI/CD.
*   **AI Abstraction:** Shows design patterns (Coordinator/Adapter) for integrating and managing multiple AI service providers flexibly.
*   **Observability Setup:** Includes the basic observability stack (Prometheus, Grafana, Jaeger) in local setup, ready for instrumentation.

---

## 🛠️ Tech Stack

| Category                 | Technologies Used                                                                          |
| :----------------------- | :----------------------------------------------------------------------------------------- |
| **Architecture**         | Microservices, REST APIs                                                                   |
| **Backend Languages**    | Go, Node.js, Perl (Planned), Python (Planned)                                             |
| **AI Services/Adapters** | AI Coordinator (Go), Ollama Adapter (Go, using `ollama/api`), Azure AI (Planned), SD (Planned) |
| **Databases**            | MongoDB (Document Store), Redis (Cache/KV Store)                                          |
| **Containerization**     | Docker                                                                                     |
| **Orchestration**        | Docker Compose (Local), Kubernetes/Helm (Planned)                                          |
| **Infrastructure (IaC)** | Terraform (Planned)                                                                        |
| **CI/CD**                | GitHub Actions                                                                             |
| **Observability**        | Prometheus, Grafana, Jaeger (Setup via Compose)                                             |
| **API Specification**    | OpenAPI 3.0 (Planned)                                                                      |
| **Security**             | OAuth 2.0 / OIDC (JWT) (Planned)                                                           |

---

## 🚀 Getting Started (Local Development)

Follow these instructions precisely to set up and run the PrivacyPilot stack locally using Docker Compose.

### 📋 Prerequisites

1.  **Git:** [Install Git](https://git-scm.com/downloads).
2.  **Docker:** [Install Docker Desktop](https://docs.docker.com/get-docker/) (Mac/Windows) or Docker Engine (Linux). Ensure Docker Compose V2 is included or installed separately. Docker daemon must be running.
3.  **Ollama:** [Install Ollama](https://ollama.com/) on your host machine. Ensure the Ollama application/server is running.
4.  **Pull an Ollama Model:** Download a model for testing (e.g., Gemma 2B). Open your terminal and run:
    ```bash
    ollama pull gemma:2b
    # You can also pull others like mistral:7b, llama3:8b etc.
    ```
5.  **(Recommended)** `jq`: A command-line JSON processor, useful for viewing API responses. [Install jq](https://jqlang.github.io/jq/download/).

### ⚙️ Installation & Setup

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/<your-username>/PrivacyPilot.git
    cd PrivacyPilot
    ```

2.  **Initialize Go Modules:**
    Run the provided script to correctly set up local Go modules for all Go services. This step is crucial for internal imports to work correctly.
    ```bash
    # Ensure the script is executable (run once)
    chmod +x ./scripts/reinit_go_mods.sh

    # Run the script from the project root
    ./scripts/reinit_go_mods.sh
    ```
    *(This script cleans old `go.mod`/`go.sum` files, runs `go mod init <module-name>`, `go get <deps>`, and `go mod tidy` in each Go service directory.)*

3.  **Configure Local Environment:**
    *   Navigate to the local DevOps directory:
        ```bash
        cd devops/local
        ```
    *   Create your local environment file from the example:
        ```bash
        cp .env.example .env
        ```
    *   **Edit the `.env` file** (or modify `docker-compose.yml` directly):
        *   **`OLLAMA_ANONYMIZE_MODEL`**: Set this to the default Ollama model you want the adapter to use if none is specified in the API request (e.g., `OLLAMA_ANONYMIZE_MODEL=gemma:2b`).
        *   **`OLLAMA_API_URL`**: Set this to the URL of your Ollama instance *as seen from within Docker containers*. **Use `http://host.docker.internal:11434`**. (Do *not* use `localhost`).
        *   Review other variables (like `GIN_MODE`, database URIs) - defaults should work initially.

### 🚀 Running the Stack

1.  **Build and Start Services:**
    Make sure you are still in the `devops/local` directory.
    ```bash
    docker-compose up --build -d
    ```
    *   `--build` forces Docker to rebuild images using the latest code.
    *   `-d` runs containers in the background.
    *   This command will:
        *   Build Docker images for all services.
        *   Start containers for: `api-gateway`, `anonymizer-service`, `moderation-service`, `ai-coordinator`, `ollama-adapter`, `mongo_db`, `redis_cache`, `prometheus`, `grafana`, `jaeger`.
        *   (It does *not* start the optional `ollama` service defined in the compose file, relying on your host Ollama instance via `host.docker.internal`).

2.  **Verify Services:**
    Check if all containers are running and healthy.
    ```bash
    docker-compose ps
    ```
    *(Look for `State: Up` or `Running`)*

3.  **Check Logs (Crucial for Debugging):**
    Monitor the logs, especially during the first startup, for any errors.
    ```bash
    # Follow logs from all services
    docker-compose logs -f

    # Check specific service logs if needed
    docker-compose logs ollama-adapter
    docker-compose logs ai-coordinator
    ```
    *(Look for connection messages, especially from `ollama-adapter` trying to reach `host.docker.internal:11434`).*

### 🧪 Testing the Installation

Use `curl` or an API client like Postman/Insomnia to interact with the API Gateway running on `http://localhost:8080`.

1.  **API Health Check:**
    ```bash
    curl http://localhost:8080/health | jq
    ```
    *   Expected: `200 OK` status and `{"service": "API Gateway", "status": "OK"}`.

2.  **Test Anonymization (Specify Model):**
    Replace `"gemma:2b"` if you pulled a different model tag.
    ```bash
    curl -X POST http://localhost:8080/api/v1/anonymize \
         -H "Content-Type: application/json" \
         -d '{
               "text": "My name is Agent Smith, contact me at smith@matrix.com or 1-800-MATRIX.",
               "config": {
                 "model": "gemma:2b"
               }
             }' | jq
    ```
    *   Expected: `200 OK` status, and a JSON response like:
        ```json
        {
          "success": true,
          "result": {
            "anonymized_text": "My name is [NAME], contact me at [EMAIL] or [PHONE].", // Example output
            "model_used": "gemma:2b"
          }
        }
        ```

3.  **Test Anonymization (Use Default Model):**
    This uses the model defined by `OLLAMA_ANONYMIZE_MODEL` in your `.env` file.
    ```bash
    curl -X POST http://localhost:8080/api/v1/anonymize \
         -H "Content-Type: application/json" \
         -d '{
               "text": "Send details to alice.wonder@example.org regarding order #987654."
             }' | jq
    ```
    *   Expected: `200 OK` and anonymized text, with `model_used` showing the default model.

4.  **Test Moderation (Expected Failure):**
    Moderation routing is set up, but no adapter is implemented yet.
    ```bash
    curl -X POST http://localhost:8080/api/v1/moderate \
         -H "Content-Type: application/json" \
         -d '{
               "text": "This is some text."
             }' | jq
    ```
    *   Expected: `500 Internal Server Error` because the AI Coordinator cannot fulfill the `moderate_text` task yet. Check `ai-coordinator` logs.

5.  **Access Observability Tools (Basic Setup):**
    *   **Grafana:** `http://localhost:3000` (Default user/pass: admin/admin)
    *   **Prometheus:** `http://localhost:9090`
    *   **Jaeger:** `http://localhost:16686`
    *(Note: Services need further instrumentation to send useful data to these tools).*

### 🛑 Stopping the Stack

```bash
# Navigate back to devops/local if you left it
cd devops/local

# Stop and remove containers, networks
docker-compose down

# To also remove volumes (database data, ollama models if using compose ollama):
# docker-compose down -v
```

---

## 📚 Project Documentation

Explore the following documents for comprehensive guidance:

- [📖 Contribution Guidelines](CONTRIBUTING.md)
- [🧑‍💻 Issue and PR Creation Guidelines](ISSUE_PR_GUIDELINES.md)
- [📜 Code of Conduct](CODE_OF_CONDUCT.md)
- [📝 Coding Style & Conventions](CODING_STYLE_AND_CONVENTIONS.md)
- [📄 License](LICENSE)
- `api-specs/` (Planned: OpenAPI definitions)

---

## 🤝 Contributing

Contributions to PrivacyPilot are greatly appreciated! Please follow the guidelines outlined in [CONTRIBUTING.md](CONTRIBUTING.md) and [ISSUE_PR_GUIDELINES.md](ISSUE_PR_GUIDELINES.md). Ensure PRs are linked to issues.

---

## 🏗️ Project Structure Overview

```text
PrivacyPilot/
├── services/           # Core backend microservices (Go, Node.js, Perl planned)
├── ai-adapters/        # Adapters for specific AI models (Go, Python planned)
├── tools/              # Standalone utility scripts (Perl planned)
├── devops/             # Docker Compose, K8s (Planned), Terraform (Planned)
├── scripts/            # Helper scripts (e.g., reinit_go_mods.sh)
├── observability/      # Prometheus, Grafana, Jaeger configurations
├── database/           # DB Migrations (Planned)
├── api-specs/          # OpenAPI definitions (Planned)
├── .github/workflows/  # CI/CD Pipelines
├── README.md           # This file
└── ...                 # Standard config and documentation files (LICENSE, .gitignore etc.)
```

---

## 📫 Contact & Support

For questions, suggestions, or to report issues, please open an issue on this repository:

- 🐛 **Report Issues:** [Open an issue](https://github.com/mihaibc/PrivacyPilot/issues)

---

## ⚖️ License

PrivacyPilot is released under the [MIT License](LICENSE).

---

### 🙌 Acknowledgments

- Inspired by privacy-focused tools and the need for robust backend showcases.
- Thanks to the open-source community for the amazing tools and frameworks used throughout this project.

---

Built with ❤️ for privacy and showcasing engineering excellence.