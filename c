#!/bin/sh
set -eu
go install
javac -cp sslr-plsql-toolkit-3.8.0.4948.jar:commons-lang3-3.12.0.jar:commons-text-1.10.0.jar -d out/production/sslr sslr/src/Main.java
sudo podman build -t tgulacsi/sslr .
sudo podman rm -f sslr_server || echo $?
sudo systemctl try-restart sslr.service