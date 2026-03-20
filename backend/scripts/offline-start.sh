#!/bin/bash

# Offline startup – all deps and source are baked into the image by the
# Dockerfile, so we skip `mix local.hex` and `mix deps.get`.

# Run DB setup (create + migrate + seed); safe to re-run on restart
MIX_ENV=prod mix ecto.setup || true

# Start Phoenix
exec mix phx.server
