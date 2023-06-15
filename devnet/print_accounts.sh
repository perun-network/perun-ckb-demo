#!/bin/bash

for entry in ./accounts/*; do
  echo $entry
  cat $entry
  echo "------------------"
done
