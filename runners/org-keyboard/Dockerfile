FROM mcr.microsoft.com/playwright:v1.49.1-jammy

WORKDIR /runner
COPY package.json /runner/package.json
COPY entrypoint.js /runner/entrypoint.js
RUN npm install --omit=dev

ENTRYPOINT ["node", "/runner/entrypoint.js"]
