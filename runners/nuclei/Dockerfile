FROM projectdiscovery/nuclei:latest
USER root
# python for JSONL → findings mapping (alpine-based images)
RUN (apk add --no-cache python3 2>/dev/null) || (microdnf install -y python3 2>/dev/null) || (apt-get update && apt-get install -y --no-install-recommends python3 && rm -rf /var/lib/apt/lists/*) || true
COPY entrypoint.sh /runner/entrypoint.sh
RUN sed -i 's/\r$//' /runner/entrypoint.sh && chmod +x /runner/entrypoint.sh
ENTRYPOINT ["/runner/entrypoint.sh"]
