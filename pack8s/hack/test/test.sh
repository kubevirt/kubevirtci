#!/bin/bash

ginkgo build -r -race
fileNames=$(find . -name "*.test")
for file in ${fileNames[@]}; do
    sudo $file
done

