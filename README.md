# ChaosCI Stats

ChaosCI Stats is a Chaos Engineering Continuous Integration platform. It integrates with GitHub Actions to run chaos experiments (via [ChaosMesh](https://chaos-mesh.org/) or [LitmusChaos](https://litmuschaos.io/)) against target environments whenever a Pull Request is raised.

## Architecture

The backend consists of three Go binaries sharing a single repository, communicating via an in-process job queue and a PostgreSQL database.

- **Webhook Server** (`cmd/webhook`): Receives GitHub `pull_request` webhooks, validates HMAC, and creates runs in the database.
- **Worker** (`cmd/worker`): Polls the database for pending jobs, parses `chaos.yaml`, and executes experiments against a Kubernetes cluster.
- **API Server** (`cmd/api`): Serves the dashboard API and Server-Sent Events (SSE) for real-time experiment updates.

## Tech Stack
- **Language**: Go 1.22+
- **Database**: PostgreSQL 16
- **Database Access**: `sqlc`
- **Migrations**: `golang-migrate`

## Local Development

### Prerequisites
- Go 1.22+
- Docker Desktop
- `sqlc` (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)
- `golang-migrate` (`go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`)

### Setup
1. Start the local PostgreSQL database:
   ```bash
   make db-up
   ```
2. Run database migrations:
   ```bash
   make migrate-up
   ```
3. Generate `sqlc` queries (if you modified `queries.sql`):
   ```bash
   make generate
   ```
4. Start the backend services (webhook, worker, api):
   ```bash
   make dev
   ```

### Environment Variables
Local development uses a `.env` file for secrets. Copy `.env.example` to `.env` to override the defaults. The `.env` file is git-ignored.

## Dashboard (Frontend)

The frontend is a SvelteKit application located in the `dashboard/` directory.

### Setup
1. Navigate to the dashboard directory:
   ```bash
   cd dashboard
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Start the development server:
   ```bash
   npm run dev
   ```
   The dashboard will be available at `http://localhost:5173`.

## Testing

To run the full stack and test the flow:
1. Ensure your Postgres database is running (`make db-up`).
2. Start the backend services in one terminal (`make dev`).
3. Start the dashboard in another terminal (`cd dashboard && npm run dev`).
4. Trigger a webhook by sending a POST request to `http://localhost:8080/webhook` (you can simulate a GitHub payload).
5. Open the dashboard at `http://localhost:5173`, enter the returned Run ID, and watch the live SSE logs!
