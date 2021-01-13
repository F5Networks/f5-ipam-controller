#!/bin/bash

# expects node:6.11.0-slim and f5-ipam-controller as volume

set -e

npm install jsonschema@1.2.11 nodeunit
schemas/scripts/run-tests.sh
rm -rf node_modules/
