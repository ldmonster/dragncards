# Backend image with all Elixir dependencies pre-fetched so the
# container can start without any network access.  Build this once while
# you still have connectivity, then subsequent `docker compose up` can be
# run offline.

FROM elixir:1.14.4

# install Hex and Rebar archives upfront
RUN mix local.hex --force && mix local.rebar --force

# copy the project; we vendor deps and build into /app so they’re baked in
WORKDIR /app
ENV MIX_ENV=prod
COPY mix.exs mix.lock ./
COPY config config

# fetch and compile dependencies now
RUN mix deps.get && mix deps.compile

# copy remaining source
COPY . ./

# compile project (optional, speeds startup)
RUN mix compile

# default entrypoint matches original dev script
ENTRYPOINT ["/app/scripts/local-dev.sh"]

CMD []
