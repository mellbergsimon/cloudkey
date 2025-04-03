#!/bin/bash
set -e

echo "Building the project..."
make

echo "Installing the project..."
make install

echo "Starting cloudkey..."
/usr/local/bin/cloudkey
