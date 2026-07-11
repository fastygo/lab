FROM returntocorp/semgrep:latest
USER root
RUN (apk add --no-cache unzip python3 2>/dev/null) \
  || (apt-get update && apt-get install -y --no-install-recommends unzip python3 && rm -rf /var/lib/apt/lists/*) \
  || true
COPY rules /rules
COPY entrypoint.sh /runner/entrypoint.sh
RUN sed -i 's/\r$//' /runner/entrypoint.sh && chmod +x /runner/entrypoint.sh
ENTRYPOINT ["/runner/entrypoint.sh"]
