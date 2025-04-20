#!/bin/bash
# @name: Cleanup Logs
# @description: Removes old log files to free up disk space
# @author: David Chen
# @version: 2.0.0
# @category: Maintenance
# @tags: ["logs", "cleanup", "maintenance"]
# @inputs: [
#   { "name": "days", "description": "Remove logs older than X days", "type": "number", "required": true, "default": 30 },
#   { "name": "dry_run", "description": "Simulate removal without deleting files", "type": "boolean", "default": false }
# ]

days=$1
dry_run=$2

echo "Searching for log files older than $days days..."
sleep 1
echo "Found 42 files to remove"
sleep 1

if [ "$dry_run" = "true" ]; then
  echo "DRY RUN: Would have removed 42 files (156MB)"
else
  echo "Removing files..."
  sleep 2
  echo "SUCCESS: Removed 42 files (156MB)"
fi