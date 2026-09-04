#!/usr/bin/env bash
# Всё в одном: Linux-сборка, Windows .exe и (если есть) NSIS-установщик.
set -euo pipefail

echo "=== Linux build ==="
go build -ldflags="-s -w" -o vizor-space .

echo "=== Windows build ==="
CGO_ENABLED=1 GOOS=windows CC=x86_64-w64-mingw32-gcc \
  go build -ldflags="-s -w -H=windowsgui" -o VizorSpace.exe .

echo "=== NSIS installer ==="
if command -v makensis >/dev/null 2>&1; then
  makensis -V2 installer.nsi
else
  echo "makensis не установлен — пропускаю."
  echo "Поставьте одним махом с sudo apt install nsis"
fi

echo "Готово."