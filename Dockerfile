FROM eclipse-temurin:17-jdk-jammy
WORKDIR /app
COPY sslr-plsql-toolkit-3.8.0.4948.jar ./
COPY out ./out
COPY run ./
ENTRYPOINT ["./run"]