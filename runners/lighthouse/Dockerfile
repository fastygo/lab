FROM node:22-bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    chromium \
    ca-certificates \
    fonts-liberation \
  && rm -rf /var/lib/apt/lists/*

ENV CHROME_PATH=/usr/bin/chromium
ENV PUPPETEER_SKIP_DOWNLOAD=true

WORKDIR /runner
COPY entrypoint.js /runner/entrypoint.js
RUN npm install -g lighthouse@12

ENTRYPOINT ["node", "/runner/entrypoint.js"]
