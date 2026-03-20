
# DragnCards

Multiplayer online card game engine written in **Elixir / Phoenix** (backend) and **React / TypeScript** (frontend).

---

## Table of Contents

1. [Local Development](#local-development)
   - [First-time setup](#first-time-setup)
   - [Managing users](#managing-users)
   - [Local image proxy](#local-image-proxy)
2. [Offline Deployment](#offline-deployment)
   - [Build images (online machine)](#1-build-images-online-machine)
   - [Transfer & install (offline machine)](#2-transfer--install-offline-machine)
   - [Makefile reference](#makefile-reference)
3. [Architecture Overview](#architecture-overview)
4. [Version Reference](#version-reference)

---

## Local Development

The `compose.yml` file at the project root starts Postgres, the Elixir
backend, and the React dev-server. The backend image is pre-built with all
Mix dependencies baked in so subsequent restarts work without network access.

### First-time setup

```sh
# 1. Start Postgres and build/start the backend
docker compose up -d backend

# 2. Create the default dev user  (alias: dev_user  |  password: password)
docker compose exec backend mix run /app/priv/create_user.exs

# 3. Start the frontend dev-server
docker compose run --rm --service-ports frontend
```

Or use the Makefile shortcut which does all of the above:

```sh
make run
```

Browse to **http://localhost:3000** and follow the
[Plugin Documentation](https://github.com/seastan/dragncards/wiki/Plugin-Documentation)
to load a game plugin.

---

### Managing users

#### Single user

```sh
docker compose exec backend mix run /app/priv/create_user.exs
```

#### Bulk user creation

Create a plain-text file with one `email password` pair per line
(passwords must be ≥ 8 characters). Example `users.txt`:

```text
alice@example.com password123
bob@example.com hunter2secret
carol@example.com mysafepass99
```

Then pipe it into the backend.

Default stack:

```sh
cat users.txt | docker compose -f compose.yml exec -T backend mix run /app/priv/batch_create_users.exs
```

Offline stack:

```sh
cat users.txt | docker compose -f compose.offline.yml exec -T backend mix run /app/priv/batch_create_users.exs
```

For a single user you can also use the Makefile shortcuts:

```sh
make add-user EMAIL=user@example.com PASSWORD=secret
```

For the offline stack:

```sh
make offline-add-user EMAIL=user@example.com PASSWORD=secret
```

#### Test players (SQL)

The following inserts four pre-hashed test accounts:

| Email | Password |
|---|---|
| player1@dragncards.com | password1 |
| player2@dragncards.com | password2 |
| player3@dragncards.com | password3 |
| player4@dragncards.com | password4 |

```sql
INSERT INTO users (email, alias, inserted_at, updated_at, password_hash, email_confirmed_at, email_confirmation_token)
VALUES ('player1@dragncards.com', 'player1', 'now', 'now',
  '$pbkdf2-sha512$100000$lBo3zNe49wIoWrAvht6Mbg==$SDfV/L5fNapiox7OgAJNB5rwrUm9RRNPCUBLHKXnNoVHcu574up2Tquxaa6shenktv7sCOtUu6rh4q0CmtOR+w==',
  'now', 'c236e80a-2c34-44b9-92ab-312df26365f9');

INSERT INTO users (id, email, alias, inserted_at, updated_at, password_hash, email_confirmed_at, email_confirmation_token)
VALUES ('7', 'player2@dragncards.com', 'player2', 'now', 'now',
  '$pbkdf2-sha512$100000$Hiwfmqbz6R0/R/q3whjVnA==$BvGkKDB/YfRnU4aQcV6INNJ8gv25Quw7SgzG64H7By5EgRdlTXIsOVHcLk7+Lf+bPqLkejAbl4F8Aanl1tASPQ==',
  'now', '6a35ba55-fd0d-47e5-aff1-d53edd5af1ec');

INSERT INTO users (id, email, alias, inserted_at, updated_at, password_hash, email_confirmed_at, email_confirmation_token)
VALUES ('8', 'player3@dragncards.com', 'player3', 'now', 'now',
  '$pbkdf2-sha512$100000$Z0jyoOb1KfzCuTGh/xVrZA==$YAlsffctWUbxujs3woZGZO6KGW++LquQAmc9MRalCXqBhaJYiOxJFjkkRjMAtbwLziVxCFD/LiRGlHutGvSpzw==',
  'now', '45eacb70-01c4-4194-a3b3-fe927bef0d0b');

INSERT INTO users (id, email, alias, inserted_at, updated_at, password_hash, email_confirmed_at, email_confirmation_token)
VALUES ('9', 'player4@dragncards.com', 'player4', 'now', 'now',
  '$pbkdf2-sha512$100000$1pFAgFabRwWro2FoLewoXw==$z0RCI+KwM68hdCxX+z+pN0mKELAd8aqvuPy+XUxNNx/ebpxrxlrxZ1fvLZ7NJQKyZnoF89NoR3fIggAYOJmEGQ==',
  'now', '9c05358a-5b4f-477a-a201-8565a842ec2f');
```

Run against the local Postgres:

```sh
psql -d dragncards_dev -f users.sql -U postgres -h 127.0.0.1
```

---

### Local image proxy

During development the app fetches card images from
s3. 
To work fully offline or with custom images:

1. Uncomment the `nginx` service in `compose.yml`:
```yaml
   nginx:
     image: nginx:latest
     container_name: my-nginx-app
     ports:
       - '8080:80'
     volumes:
       - ./images:/usr/share/nginx/html
```

2. Place asset files under `./images/` and start the proxy:
```sh
   docker compose up -d nginx
```

   The container serves `./images/` over HTTP on port **8080** with a network
   alias matching the S3 hostname.

4. When done, remove the hosts entry or stop the service:
```sh
   docker compose stop nginx
```

---

## Offline Deployment

Use the `Makefile` to build self-contained image archives on an internet-connected
machine, then transfer and install them on an air-gapped machine.

```
compose.offline.yml   ← production-style compose (baked images, no source volume mounts)
Makefile              ← build / install targets
build/                ← saved .tar archives (git-ignored)
```

### 1. Build images (online machine)

```sh
make offline-build
```

This will:
- Build `dragncards/backend:latest` from `./backend/Dockerfile`
- Build `dragncards/frontend:latest` from `./frontend/Dockerfile`
- Pull `postgres:16` and `nginx:latest`
- Save all four as `.tar` archives in `./build/`

```
build/
  backend.tar    (~1.5 GB)
  frontend.tar   (~63 MB)
  postgres.tar   (~437 MB)
  nginx.tar      (~70 MB)
```

Transfer the entire project directory (including `./build/`) to the offline machine
via USB drive, `scp`, or any other method.

### 2. Transfer & install (offline machine)

```sh
make offline-install
```

This will:
- Load all four `.tar` archives into the local Docker daemon
- Start the stack with `docker compose -f compose.offline.yml up -d`

Services and ports:

| Service | URL |
|---|---|
| Frontend (nginx) | http://localhost:3000 |
| Backend (Phoenix) | http://localhost:4000 |
| Card image proxy (nginx)  | http://localhost:8080 |
| Postgres | localhost:5433 |

To stop the offline stack:

```sh
make offline-stop
```

---

## Architecture Overview

```
┌─────────────┐   WebSocket / HTTP    ┌────────────────────┐
│   Browser   │ ────────────────────▶ │  nginx (frontend)  │  :3000
│             │                       │  React SPA         │
└─────────────┘                       └─────────┬──────────┘
                                                 │ /be/* proxy
                                      ┌──────────▼──────────┐
                                      │  Phoenix (backend)  │  :4000
                                      │  Elixir / Ecto      │
                                      └──────────┬──────────┘
                                                 │
                                      ┌──────────▼──────────┐
                                      │     PostgreSQL      │  :5433
                                      └─────────────────────┘

                                      ┌─────────────────────┐
                                      │  nginx              │  :8080
                                      │  Card image proxy   │
                                      └─────────────────────┘
```

The nginx config in `frontend/nginx/nginx.conf` proxies `/be/*` to the backend.
In `compose.offline.yml` the backend service is named `backend`
to match that proxy target exactly.

---

## Version Reference

Versions known to work:

| Component       | Version   | Notes                  |
|-----------------|-----------|------------------------|
| Docker Engine   | 27.3.1    | linux/arm64, Go 1.22.7 |
| Docker Compose  | v2.29.7   |                        |
| containerd      | 1.7.22    |                        |
| runc            | 1.1.14    |                        |
| Elixir          | 1.14.4    |                        |
| Node.js         | 20        | alpine                 |
| PostgreSQL      | 16        |                        |
| nginx           | 1.25      | alpine                 |

Tested on Ubuntu 22.04 and 24.04 Server (linux/arm64 & linux/amd64).






