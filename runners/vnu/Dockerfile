FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl ca-certificates openjdk-17-jre-headless python3 unzip \
  && rm -rf /var/lib/apt/lists/*

# Nu Html Checker standalone jar
ARG VNU_VERSION=24.10.17
RUN curl -fsSL "https://github.com/validator/validator/releases/download/${VNU_VERSION}/vnu.jar_${VNU_VERSION}.zip" -o /tmp/vnu.zip \
  && unzip -q /tmp/vnu.zip -d /opt \
  && mv /opt/dist/vnu.jar /vnu.jar \
  && rm -rf /tmp/vnu.zip /opt/dist \
  && printf '#!/bin/sh\nexec java -jar /vnu.jar "$@"\n' > /usr/local/bin/vnu \
  && chmod +x /usr/local/bin/vnu

COPY entrypoint.sh /runner/entrypoint.sh
RUN chmod +x /runner/entrypoint.sh
ENTRYPOINT ["/runner/entrypoint.sh"]
