FROM node:22-alpine AS base

# Dev stage
FROM base AS dev
WORKDIR /app
ARG NEXT_PUBLIC_UMAMI_WEBSITE_ID
ENV NEXT_PUBLIC_UMAMI_WEBSITE_ID=$NEXT_PUBLIC_UMAMI_WEBSITE_ID
COPY package.json ./
RUN npm install
COPY . .
ENV NEXT_PUBLIC_API_URL=https://sharepwd.io
ENV NEXT_PUBLIC_APP_URL=https://sharepwd.io
CMD ["npm", "run", "dev"]

# Build stage
FROM base AS builder
WORKDIR /app
ARG NEXT_PUBLIC_UMAMI_WEBSITE_ID
ENV NEXT_PUBLIC_UMAMI_WEBSITE_ID=$NEXT_PUBLIC_UMAMI_WEBSITE_ID
COPY package.json ./
RUN npm install
COPY . .
ENV NEXT_PUBLIC_API_URL=https://sharepwd.io
ENV NEXT_PUBLIC_APP_URL=https://sharepwd.io
RUN npm run build

# Production stage
FROM base AS production
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public
EXPOSE 3000
USER node
CMD ["node", "server.js"]
