#!/bin/sh
curl -s -X POST http://127.0.0.1:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"phone":"+79999999999","password":"admin123"}'
