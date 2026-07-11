FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl ca-certificates openjdk-17-jre-headless python3 \
  && rm -rf /var/lib/apt/lists/*

# Nu Html Checker — project ships production jar under the "latest" release tag
RUN curl -fsSL "https://github.com/validator/validator/releases/download/latest/vnu.jar" -o /vnu.jar \
  && printf '#!/bin/sh\nexec java -jar /vnu.jar "$@"\n' > /usr/local/bin/vnu \
  && chmod +x /usr/local/bin/vnu

COPY entrypoint.sh /runner/entrypoint.sh
RUN chmod +x /runner/entrypoint.sh \
  && sed -i 's/\r$//' /runner/entrypoint.sh
ENTRYPOINT ["/runner/entrypoint.sh"]
