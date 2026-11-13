# 🛒 GoCommerce

A full-stack e-commerce platform. The backend is written in **Go using only the standard library (`net/http`, no framework)**. Data lives in **PostgreSQL**, product data is cached in **Redis**, and the frontend is **React (Context API, no Redux)**. Each app has its own **Dockerfile**, and it deploys to **Kubernetes** through **GitHub Actions + ArgoCD** (GitOps).

The goal is to be easy to read: every file has comments explaining *why*, which makes the project a good way to learn or revise Go backend concepts.

![Go](https://img.shields.io/badge/Go-1.27-00ADD8?logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=black)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-4169E1?logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7-DC382D?logo=redis&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-multi--stage-2496ED?logo=docker&logoColor=white)
![Kubernetes](https://img.shields.io/badge/Kubernetes-ready-326CE5?logo=kubernetes&logoColor=white)
![GitHub Actions](https://img.shields.io/badge/GitHub_Actions-CI%2FCD-2088FF?logo=githubactions&logoColor=white)
![ArgoCD](https://img.shields.io/badge/ArgoCD-GitOps-EF7B4D?logo=argo&logoColor=white)

![Shop home page](docs/screenshots/01-home.png)

---

## 📑 Table of contents

- [Features](#-features)
- [Screenshots](#-screenshots)
- [Tech stack](#-tech-stack)
- [Architecture](#-architecture)
- [Project structure](#-project-structure)
- [Run it locally](#-run-it-locally)
- [Docker images](#-docker-images)
- [Deployment: GitHub Actions + ArgoCD](#-deployment-github-actions--argocd)
- [Configuration](#-configuration)
- [API reference](#-api-reference)
- [Database schema](#-database-schema)
- [How the important parts work](#-how-the-important-parts-work)
- [Testing](#-testing)
- [Go concepts used (learning map)](#-go-concepts-used-learning-map)
- [Troubleshooting](#-troubleshooting)
- [Ideas for future work](#-ideas-for-future-work)

---

## ✨ Features

**Customers**
- Sign up / log in with JWT authentication (passwords hashed with bcrypt)
- Browse products with instant **search** (runs in the browser, no API call per keystroke), **category filter** and **pagination**
- Product detail page with live stock
- Shopping cart: add, change quantity, remove, clear
- Checkout that turns the cart into an order (**inside a database transaction**)
- Order history, and cancelling a pending order (the items go back into stock)

**Admin panel** (`/admin`)
- Dashboard: revenue, orders, pending orders, low-stock products, users
- Product **CRUD**: create, read, update, delete
- Manage orders: change status `pending → paid → shipped → delivered` or `cancelled`
- List all users

**Engineering**
- Stock is never oversold, even when many people buy the last item at the same moment (row locking with `SELECT ... FOR UPDATE`)
- Redis caching of product data with automatic invalidation (look for the `X-Cache: HIT/MISS` header)
- The app keeps working if Redis is down (it falls back to Postgres)
- Graceful shutdown, panic recovery, request logging, CORS
- Unit tests (Go) + end-to-end browser tests (Playwright)

---

## 📸 Screenshots

| Product page | Cart |
|---|---|
| ![Product detail](docs/screenshots/03-product-detail.png) | ![Cart](docs/screenshots/05-cart.png) |

| My orders | Admin dashboard |
|---|---|
| ![Orders](docs/screenshots/07-order-cancelled.png) | ![Admin dashboard](docs/screenshots/08-admin-dashboard.png) |

| Admin: products CRUD | Admin: orders |
|---|---|
| ![Admin products](docs/screenshots/10-admin-product-form.png) | ![Admin orders](docs/screenshots/11-admin-orders.png) |

| Admin: users | Mobile |
|---|---|
| ![Admin users](docs/screenshots/12-admin-users.png) | <img src="docs/screenshots/13-mobile-home.png" width="260" alt="Mobile view"> |

---

## 🧰 Tech stack

| Layer | Technology | Why |
|---|---|---|
| Backend language | **Go 1.27** | Fast, simple, great standard library |
| HTTP server / router | **`net/http`** (standard library) | Since Go 1.22, `ServeMux` supports `GET /products/{id}`, so no framework is needed |
| Database | **PostgreSQL 17** | Transactions, row locking, constraints |
| DB driver | `database/sql` + [`lib/pq`](https://github.com/lib/pq) | The standard Go way to talk to SQL databases |
| Cache | **Redis 7** + [`go-redis/v9`](https://github.com/redis/go-redis) | Caches product lists and details |
| Auth | [`golang-jwt/jwt/v5`](https://github.com/golang-jwt/jwt) + `bcrypt` | Stateless login tokens + safe password hashing |
| Frontend | **React 19** + **Vite** | Fast dev server and build |
| Routing (frontend) | React Router 7 | Pages and nested admin routes |
| State management | **Context API** + props | `AuthContext` and `CartContext`, no Redux |
| Web server (prod) | **nginx** | Serves the React build and proxies `/api` to Go |
| Containers | **Docker** (multi-stage builds) | Small images: ~20 MB backend, nginx frontend |
| Registry | **GitHub Container Registry** (ghcr.io) | Free, works with the built-in `GITHUB_TOKEN` |
| CI/CD | **GitHub Actions** | Test → build images → update image tags |
| Deployment (GitOps) | **ArgoCD** | Watches `k8s/` in git and syncs the cluster |
| Orchestration | **Kubernetes** | Simple Deployments + Services |
| Testing | `go test`, **Playwright** | Unit tests + real-browser end-to-end tests |

---

## 🏗 Architecture

```mermaid
flowchart LR
    B[Browser] -->|":30080"| N["nginx<br/>(React build)"]
    N -->|"/api/*"| G["Go API<br/>net/http :8080"]
    G -->|SQL| P[(PostgreSQL)]
    G -->|cache| R[(Redis)]
```

A request passes through these layers inside the Go server:

```
Logger → Recover → CORS → Router (ServeMux) → [Auth / Admin middleware] → Handler → Postgres / Redis
```

---

## 📁 Project structure

```
.
├── backend/                     Go API
│   ├── main.go                  startup + graceful shutdown
│   ├── routes.go                ⭐ every endpoint in one place - start reading here
│   ├── config/                  reads environment variables
│   ├── db/
│   │   ├── postgres.go          connect (with retry), migrate, seed demo data
│   │   └── schema.sql           tables (embedded into the binary with //go:embed)
│   ├── cache/redis.go           tiny Redis wrapper (JSON get/set, delete by pattern)
│   ├── auth/                    JWT + bcrypt (+ tests)
│   ├── middleware/              Logger, Recover, CORS, Auth, Admin
│   ├── models/                  structs sent as JSON (+ tests)
│   ├── utils/json.go            WriteJSON / WriteError / ReadJSON helpers
│   ├── handlers/
│   │   ├── auth.go              register, login, me
│   │   ├── products.go          list (search/filter/paginate/cache), get, categories
│   │   ├── cart.go              get, add, update, remove, clear
│   │   ├── orders.go            ⭐ checkout transaction, list, cancel + restock
│   │   └── admin.go             product CRUD, order status, users, stats
│   └── Dockerfile
├── frontend/                    React app
│   ├── src/
│   │   ├── api.js               every call to the backend
│   │   ├── context/             AuthContext, CartContext (global state)
│   │   ├── components/          Navbar, ProductCard, ProtectedRoute, Pagination...
│   │   ├── pages/               Home, ProductDetail, Login, Register, Cart, Orders
│   │   └── pages/admin/         Dashboard, AdminProducts, ProductForm, AdminOrders, AdminUsers
│   ├── public/images/           product images (work offline)
│   ├── e2e/shop.spec.js         Playwright browser tests
│   ├── nginx.conf
│   └── Dockerfile
├── k8s/                         Kubernetes: postgres, redis, backend, frontend
├── argocd/application.yaml      tells ArgoCD to deploy the k8s/ folder
├── .github/workflows/ci-cd.yml  GitHub Actions pipeline
└── docs/screenshots/
```

---

## 🚀 Run it locally

**Demo admin account** (created automatically on first start):

```
email:    admin@gocommerce.com
password: admin123
```

**Requires:** Go 1.27+, Node.js 20+, PostgreSQL, Redis (optional)

```bash
# macOS (Homebrew) - skip what you already have
brew install go node postgresql@17 redis
brew services start postgresql@17
brew services start redis
```

**1. Create the database**

```bash
createdb gocommerce
```

**2. Start the backend** (terminal 1)

```bash
cd backend
go mod download
DATABASE_URL="postgres://$(whoami)@localhost:5432/gocommerce?sslmode=disable" go run .
```

You should see:

```
connected to postgres
seeded 8 demo products
connected to redis at localhost:6379
GoCommerce API listening on http://localhost:8080
```

> On Linux/Windows the default Postgres user is usually `postgres`:
> `DATABASE_URL="postgres://postgres:yourpassword@localhost:5432/gocommerce?sslmode=disable"`

**3. Start the frontend** (terminal 2)

```bash
cd frontend
npm install
npm run dev
```

Open http://localhost:5173 (admin panel: http://localhost:5173/admin). Vite forwards every `/api` request to `localhost:8080`.

> **No Redis?** That's fine. The API logs a warning and reads from Postgres every time. Set `REDIS_ADDR=""` to turn caching off completely.

---

## 🐳 Docker images

There is one Dockerfile per app. Both are **multi-stage**: stage 1 builds, stage 2 copies only the result into a small image.

| Image | Dockerfile | Stage 1 | Stage 2 (final image) |
|---|---|---|---|
| backend | [`backend/Dockerfile`](backend/Dockerfile) | `golang` compiles a static binary | `alpine` + the binary (~20 MB), runs as non-root |
| frontend | [`frontend/Dockerfile`](frontend/Dockerfile) | `node` runs `npm run build` | `nginx` serves the files and proxies `/api` to `backend:8080` |

You normally don't build them yourself, because GitHub Actions does it. To try a build locally:

```bash
docker build -t gocommerce-backend ./backend
docker build -t gocommerce-frontend ./frontend
```

---

## 🚢 Deployment: GitHub Actions + ArgoCD

This project uses **GitOps**: git is the single source of truth for what runs in the cluster. GitHub Actions never talks to Kubernetes. It only builds images and writes the new image tag into `k8s/`, and **ArgoCD** notices the change and deploys it.

```mermaid
flowchart LR
    D[git push to main] --> T["GitHub Actions<br/>1. test"]
    T --> B["2. build & push images<br/>ghcr.io/…:commit-sha"]
    B --> U["3. update image tag<br/>in k8s/*.yaml + commit"]
    U --> A["ArgoCD sees the change<br/>in the k8s/ folder"]
    A --> K["Kubernetes<br/>rolls out new pods"]
```

### The pipeline ([`.github/workflows/ci-cd.yml`](.github/workflows/ci-cd.yml))

| Job | Runs on | What it does |
|---|---|---|
| `test` | every push and pull request | `go vet`, `go test ./...`, `npm run build` |
| `build` | push to `main` | builds both Dockerfiles in parallel, pushes `ghcr.io/<you>/gocommerce-backend:<sha>` and `…-frontend:<sha>` (plus `latest`) |
| `deploy` | after `build` | replaces the image tag in `k8s/backend.yaml` and `k8s/frontend.yaml` with the commit SHA, then commits with `[skip ci]` |

Each image is tagged with the **commit SHA**, so you always know exactly which code is running, and rolling back is just reverting a commit.

No secrets need to be added: the built-in `GITHUB_TOKEN` can push to ghcr.io and commit to the repo.

### Kubernetes manifests ([`k8s/`](k8s/))

Kept to the basics: one **Deployment** (runs the containers) and one **Service** (gives them a fixed network name) per app.

```
k8s/
├── postgres.yaml   Deployment + Service "postgres"  (port 5432)
├── redis.yaml      Deployment + Service "redis"     (port 6379)
├── backend.yaml    Deployment (2 replicas) + Service "backend" (port 8080)
└── frontend.yaml   Deployment + Service "frontend"  (NodePort 30080 → open in browser)
```

How the pieces find each other: the backend connects to `postgres:5432` and `redis:6379`, and nginx in the frontend sends `/api` to `backend:8080`. These are the **Service names**, and Kubernetes DNS resolves them.

> The backend may restart once or twice while Postgres is still starting. That's normal: Kubernetes restarts it until the database is ready.
> The database has no persistent volume (kept simple on purpose), so data is lost if the Postgres pod is deleted.

### One-time setup

**1. Push the code to GitHub**

```bash
git init && git add . && git commit -m "Initial commit"
git branch -M main
git remote add origin https://github.com/<you>/gocommerce.git
git push -u origin main
```

**2. Allow the workflow to commit.** Repo → *Settings* → *Actions* → *General* → *Workflow permissions* → select **Read and write permissions**.

**3. Replace the placeholders.** Put your GitHub username (lowercase) in:
- `k8s/backend.yaml` and `k8s/frontend.yaml` → `ghcr.io/YOUR_GITHUB_USERNAME/...` (CI also rewrites this after the first run)
- `argocd/application.yaml` → `repoURL`

**4. Make the images pullable.** After the first pipeline run, open your GitHub profile → *Packages* → `gocommerce-backend` → *Package settings* → *Change visibility* → **Public**. Do the same for `gocommerce-frontend`.
(If you want them private, create an `imagePullSecret` in the `gocommerce` namespace and reference it in the Deployments instead.)

**5. Install ArgoCD in your cluster** (skip if you already have it):

```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

**6. Register the app with ArgoCD** (once):

```bash
kubectl apply -f argocd/application.yaml
```

ArgoCD now deploys everything in `k8s/` into the `gocommerce` namespace, and redeploys after every pipeline run.

**7. Open it**

```bash
kubectl get pods -n gocommerce                 # wait until all are Running
# shop: http://<node-ip>:30080   (Docker Desktop / minikube tunnel: http://localhost:30080)
```

ArgoCD UI (optional):

```bash
kubectl port-forward svc/argocd-server -n argocd 8443:443
# https://localhost:8443  user: admin
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

### Day-to-day

```bash
git commit -am "Add product reviews" && git push    # that's it - CI/CD + ArgoCD do the rest
```

Useful commands:

```bash
kubectl get pods,services -n gocommerce
kubectl logs -n gocommerce deploy/backend
kubectl get application gocommerce -n argocd     # ArgoCD sync status
```

**Rollback:** `git revert <the "deploy: …" commit>` and push. ArgoCD switches back to the previous image.

---

## ⚙️ Configuration

The backend reads these environment variables. All have defaults, so nothing is required for local use.

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP port |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/gocommerce?sslmode=disable` | Postgres connection string |
| `REDIS_ADDR` | `localhost:6379` | Redis address. Empty string = caching disabled |
| `JWT_SECRET` | `dev-secret-change-me` | Secret used to sign login tokens. **Change in production!** |
| `ADMIN_EMAIL` | `admin@gocommerce.com` | Admin account created on startup |
| `ADMIN_PASSWORD` | `admin123` | Password for that admin account |

The frontend has one optional variable: `VITE_API_URL` (default `/api`).

---

## 📡 API reference

Base URL: `http://localhost:8080/api`

Send `Authorization: Bearer <token>` for 🔒 (logged-in) and 👑 (admin) routes.
Errors always look like `{"error": "message"}`. Money is always in **cents** (`price_cents: 1999` = $19.99).

### Auth

| Method | Endpoint | | Body | Returns |
|---|---|---|---|---|
| POST | `/auth/register` | | `{name, email, password}` | `{token, user}` |
| POST | `/auth/login` | | `{email, password}` | `{token, user}` |
| GET | `/auth/me` | 🔒 | | current user |

### Products (public)

| Method | Endpoint | Description |
|---|---|---|
| GET | `/products?search=&category=&page=1&limit=12` | Paginated list `{products, total, page, limit}` (cached). The React app loads all pages once and searches client-side; `search`/`category` params are still available for other clients |
| GET | `/products/{id}` | One product (cached) |
| GET | `/categories` | List of category names (cached) |

### Cart 🔒

| Method | Endpoint | Body | Description |
|---|---|---|---|
| GET | `/cart` | | `{items, total_cents}` |
| POST | `/cart` | `{product_id, quantity}` | Add (adds to the existing quantity) |
| PUT | `/cart/{productId}` | `{quantity}` | Set quantity |
| DELETE | `/cart/{productId}` | | Remove one item |
| DELETE | `/cart` | | Empty the cart |

### Orders 🔒

| Method | Endpoint | Description |
|---|---|---|
| POST | `/orders` | **Checkout**: cart → order. `409` if not enough stock, `400` if the cart is empty |
| GET | `/orders` | My orders (with items) |
| GET | `/orders/{id}` | One of my orders |
| POST | `/orders/{id}/cancel` | Cancel a `pending` order; stock is restored |

### Admin 👑

| Method | Endpoint | Body | Description |
|---|---|---|---|
| GET | `/admin/stats` | | Dashboard numbers |
| POST | `/admin/products` | product JSON | Create product |
| PUT | `/admin/products/{id}` | product JSON | Update product |
| DELETE | `/admin/products/{id}` | | Delete product (`204`) |
| GET | `/admin/orders?status=` | | All orders, optional status filter |
| PUT | `/admin/orders/{id}/status` | `{status}` | `pending` / `paid` / `shipped` / `delivered` / `cancelled` |
| GET | `/admin/users` | | All users |

Product JSON:

```json
{
  "name": "Notebook",
  "description": "A5 dotted notebook",
  "price_cents": 499,
  "stock": 50,
  "category": "Office",
  "image_url": "https://example.com/notebook.jpg"
}
```

### Try it with curl

```bash
# log in as admin and save the token
TOKEN=$(curl -s -X POST localhost:8080/api/auth/login \
  -d '{"email":"admin@gocommerce.com","password":"admin123"}' | jq -r .token)

# create a product
curl -X POST localhost:8080/api/admin/products \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Notebook","price_cents":499,"stock":50,"category":"Office"}'

# list products - run twice and watch X-Cache go from MISS to HIT
curl -i "localhost:8080/api/products?limit=2" | grep X-Cache
```

---

## 🗄 Database schema

```mermaid
erDiagram
    users ||--o{ cart_items : has
    users ||--o{ orders : places
    products ||--o{ cart_items : "in"
    orders ||--|{ order_items : contains
    products |o--o{ order_items : "was"

    users {
        int id PK
        text name
        text email UK
        text password_hash
        text role "customer | admin"
    }
    products {
        int id PK
        text name
        bigint price_cents "CHECK >= 0"
        int stock "CHECK >= 0"
        text category
        text image_url
    }
    cart_items {
        int user_id FK
        int product_id FK
        int quantity
    }
    orders {
        int id PK
        int user_id FK
        bigint total_cents
        text status
    }
    order_items {
        int order_id FK
        int product_id FK "NULL if product deleted"
        text product_name "copied at purchase"
        bigint price_cents "copied at purchase"
        int quantity
    }
```

Design notes:
- **Money is stored as integer cents**, never as a float (floats can't represent 0.1 exactly).
- `order_items` **copies** the product name and price, so order history stays correct after the admin edits or deletes a product.
- `CHECK (stock >= 0)` is a last safety net: the database itself refuses negative stock.
- Tables are created automatically on startup from [`backend/db/schema.sql`](backend/db/schema.sql).

---

## 🔍 How the important parts work

### 1. Checkout without overselling (transactions + row locks)

See `placeOrder` in [`backend/handlers/orders.go`](backend/handlers/orders.go). Everything happens in **one transaction**, so either all steps succeed or none do:

```
BEGIN
  1. SELECT products in my cart ... FOR UPDATE   ← locks those rows; other buyers WAIT here
  2. check stock for every item                   → not enough? ROLLBACK, return 409
  3. INSERT order + order_items
  4. UPDATE products SET stock = stock - quantity
  5. DELETE my cart items
COMMIT
then: clear product cache
```

**Why the lock matters:** two people buy the last lamp at the same moment. Without `FOR UPDATE`, both read `stock = 1`, both succeed, and stock becomes `-1`. With the lock, the second transaction waits, then sees `stock = 0` and gets a `409`.

*Tested:* 5 users bought an item with only 3 in stock at the same time. Exactly 3 got `201`, 2 got `409`, and stock ended at `0`.

Cancelling an order (`changeOrderStatus`) is also a transaction: it locks the order, puts every item back into stock, and updates the status.

### 2. Redis caching (cache-aside pattern)

See [`backend/handlers/products.go`](backend/handlers/products.go) and [`backend/cache/redis.go`](backend/cache/redis.go).

```
GET /api/products
   │
   ├─ key in Redis? ── yes ──► return cached JSON           (X-Cache: HIT)
   │
   └─ no ──► query Postgres ──► save to Redis for 5 min ──► return   (X-Cache: MISS)

Product created / updated / deleted, order placed / cancelled
   └──► delete all keys matching "gocommerce:products:*"
```

If Redis is unreachable, `Get` simply reports a miss, so the app is slower but still correct.

### 3. Authentication

1. On register or login, the password is checked with **bcrypt** and the server returns a **JWT** containing `user_id` and `role` (valid for 24h).
2. The frontend stores the token in `localStorage` and sends it as `Authorization: Bearer <token>`.
3. The `Auth` middleware verifies the token and puts the user id in the request **context**. The `Admin` middleware also checks `role == "admin"`.
4. The React `ProtectedRoute` only *hides* pages. The real protection is the Go middleware (a customer calling an admin endpoint gets `403`).

### 4. Frontend state without Redux

- **`AuthContext`** holds the logged-in user plus `login`, `register` and `logout`.
- **`CartContext`** holds the cart (loaded from the server when the user changes) plus `addToCart`, `updateQuantity`, `removeItem` and `checkout`.
- Everything else uses **props**, for example `Home` passes `onAddToCart` down to each `ProductCard`, and `AdminProducts` passes `onSubmit` / `onCancel` to `ProductForm`.

---

## 🧪 Testing

### Backend unit tests

```bash
cd backend
go test ./...
```

Covers JWT creation/validation, password hashing, and product validation (table-driven tests).

### End-to-end browser tests (Playwright)

A real Chrome browser clicks through the whole app. Start the backend (`go run .`) and frontend (`npm run dev`) first, as in [Run it locally](#-run-it-locally), then:

```bash
cd frontend
npm install
npx playwright install chromium   # only needed if you don't have Google Chrome
npm run test:e2e
```

3 tests, 27 steps and 63 assertions cover:

| Customer journey | Admin | Mobile |
|---|---|---|
| Product list, client-side search (no API calls while typing), category filter | Admin login → dashboard | No horizontal scroll at 390px width |
| Out-of-stock button disabled | Create / edit / delete product | |
| Login required for cart | Validation errors shown | |
| Register, duplicate email, wrong password | New product appears in the shop (cache cleared) | |
| Add to cart, badge count, change quantity, totals | Change order status | |
| Checkout reduces stock | Cancelled orders are locked | |
| Cancel order restores stock | Status filter, users list | |
| Customer cannot open `/admin` | | |

Screenshots of every screen are saved to `frontend/e2e/screenshots/`.

> To test a deployed copy instead: `BASE_URL=http://localhost:30080 npm run test:e2e`

### In CI

GitHub Actions runs `go vet`, `go test ./...` and the frontend build on every push and pull request. Images are only built and deployed if these pass.

---

## 🎓 Go concepts used (learning map)

| Concept | Where to look |
|---|---|
| `net/http` routing with methods and path params (`GET /products/{id}`, `r.PathValue`) | `routes.go`, `handlers/handler.go` |
| Structs, methods, pointer receivers | `models/models.go` (`Validate`), `handlers/` |
| Struct tags and JSON (`json:"-"` hides a field) | `models/models.go` |
| Interfaces (`scanner` works for both `*sql.Row` and `*sql.Rows`) | `handlers/products.go` |
| Struct embedding (`statusRecorder` embeds `http.ResponseWriter`) | `middleware/middleware.go` |
| Closures and middleware (functions returning functions) | `middleware/middleware.go`, `routes.go` |
| `context.Context`: request-scoped values and timeouts | `middleware/middleware.go`, `handlers/handler.go` |
| Error handling: sentinel errors, custom error types, `%w` wrapping, `errors.Is` / `errors.As` | `handlers/orders.go` |
| `defer`: closing rows, rolling back transactions, `recover()` from panics | `handlers/orders.go`, `middleware/middleware.go` |
| `database/sql`: `Query`, `QueryRow`, `Exec`, `sql.ErrNoRows`, `sql.NullInt64` | `handlers/` |
| **Transactions + `SELECT ... FOR UPDATE`** | `handlers/orders.go` |
| Safe dynamic SQL with `$1` placeholders (no SQL injection) | `handlers/products.go` → `ListProducts` |
| Avoiding the N+1 query problem (`= ANY($1)` + `pq.Array`) | `handlers/orders.go` → `fetchOrders` |
| Goroutines, channels, OS signals, graceful shutdown | `main.go` |
| `//go:embed` to bundle files into the binary | `db/postgres.go` |
| Dependency injection via a `Handler` struct | `handlers/handler.go` |
| Table-driven tests, `t.Run` | `models/models_test.go`, `auth/auth_test.go` |
| Multi-stage Docker build → small static binary | `backend/Dockerfile` |

---

## 🛠 Troubleshooting

| Problem | Fix |
|---|---|
| `address already in use` (8080 or 5173) | Stop whatever uses the port, or run the backend with `PORT=8081` and start Vite with `API_URL=http://localhost:8081 npm run dev` |
| `role "postgres" does not exist` (local run) | Your Postgres user is your system user: use `postgres://$(whoami)@localhost:5432/gocommerce?sslmode=disable` |
| `database "gocommerce" does not exist` | Run `createdb gocommerce` |
| `redis ... not reachable` warning | Start Redis, or ignore it: the app works without the cache |
| Frontend shows `Request failed (500/502)` | The backend isn't running. Start it with `go run .` (locally) or check `kubectl logs -n gocommerce deploy/backend` |
| Changes to demo products don't appear | Demo data is only inserted when the products table is empty: `dropdb gocommerce && createdb gocommerce` to reset |
| Workflow fails at `git push` with 403 | Repo → Settings → Actions → General → **Read and write permissions** |
| Workflow fails at image push with `denied` | Image names must be lowercase; also check the package isn't linked to a different repo |
| Pod stuck in `ImagePullBackOff` | The ghcr.io package is private: make it public, or add an `imagePullSecret` |
| ArgoCD shows `OutOfSync` / doesn't deploy | Check `repoURL` in `argocd/application.yaml`, and that the repo is public (or added under ArgoCD → Settings → Repositories) |

---

## 🗺 Ideas for future work

- [ ] Promote a user to admin (`PATCH /api/admin/users/{id}/role`)
- [ ] Product reviews and average rating
- [ ] Rate-limit login attempts with Redis (`INCR` + `EXPIRE`)
- [ ] Handler tests with `net/http/httptest`
- [ ] Move SQL into a `repository` layer
- [ ] Fake payment step: `POST /api/orders/{id}/pay` (`pending → paid`)
- [ ] Image upload instead of image URLs
- [ ] Run the Playwright browser tests in GitHub Actions (Postgres + Redis as service containers)
- [ ] Persistent volume for Postgres, and Secrets instead of plain env values in `k8s/`
