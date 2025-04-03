FROM debian:stable-slim
COPY ./fake_ollama_linux_amd64 /usr/local/bin/fake_ollama_linux_amd64
RUN chmod +x /usr/local/bin/fake_ollama_linux_amd64
CMD ["/usr/local/bin/fake_ollama_linux_amd64"]
