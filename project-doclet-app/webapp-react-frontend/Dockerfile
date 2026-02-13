FROM node:20-alpine AS builder
WORKDIR /app
ARG VITE_DOC_SERVICE_URL
ARG VITE_COLLAB_WS_URL
ENV VITE_DOC_SERVICE_URL=$VITE_DOC_SERVICE_URL
ENV VITE_COLLAB_WS_URL=$VITE_COLLAB_WS_URL
COPY package.json package-lock.json ./
RUN npm ci
COPY . ./
RUN npm run build

FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
