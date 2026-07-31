# Build stage
FROM node:20-alpine AS builder

WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm install

COPY . .
RUN npm run build -- --configuration production

# Final stage - Nginx
FROM nginx:alpine

# Copy custom Nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy build output
COPY --from=builder /app/dist/minerva/browser /usr/share/nginx/html 

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
