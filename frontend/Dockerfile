# build environment
FROM node:20-alpine AS build
WORKDIR /app
ENV PATH=/app/node_modules/.bin:$PATH
# Use relative asset paths so the built image works at any subpath without rebuild.
# The actual base href is injected into index.html at container start by the entrypoint.
ENV PUBLIC_URL=.
COPY package.json /app/package.json
RUN npm install --legacy-peer-deps --silent
COPY . /app
# compile tailwind CSS then build JS bundle
RUN npm run build:css && npm run build

# production environment
FROM nginx:1.25-alpine
COPY --from=build /app/build /usr/share/nginx/html
RUN rm /etc/nginx/conf.d/default.conf
COPY nginx/nginx.conf.template /etc/nginx/nginx.conf.template
COPY nginx/entrypoint.sh /docker-entrypoint-app.sh
RUN chmod +x /docker-entrypoint-app.sh
EXPOSE 80
CMD ["/docker-entrypoint-app.sh"]
