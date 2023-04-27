FROM eclipse-temurin:17-jdk-jammy
WORKDIR /app
COPY sslr-plsql-toolkit-3.8.0.4948.jar commons-lang3-3.12.0.jar commons-text-1.10.0.jar ./
COPY out ./out
COPY run ./
ENTRYPOINT ["./run"]