FROM node:22-bookworm-slim
RUN apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq unzip python3 >/dev/null \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /runner
COPY package.json /runner/package.json
COPY stylelint.config.mjs /runner/stylelint.config.mjs
COPY entrypoint.js /runner/entrypoint.js
RUN npm install --omit=dev
ENTRYPOINT ["node", "/runner/entrypoint.js"]
